package session

import (
	"strings"
	"text/scanner"
	"unicode"
	"unicode/utf8"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type Toks []*Tok
type Tok struct {
	byteOffset int
	Kind       TokKind
	Pos        SrcFilePos
	Src        string
}
type TokKind int

const (
	_ TokKind = iota
	TokKindBegin
	TokKindEnd
	TokKindComment    // both /* multi-line */ and // single-line
	TokKindBracketing // parens, square brackets, curly braces
	// below: only toks that, if no sep-or-ws between them, will `huddle` together
	// into their own single contiguous expr as if parensed (above: those that won't)
	TokKindIdentWord  // lexemes that pass the `IsIdentRune` predicate below
	TokKindIdentOpish // all lexemes that dont match any other TokKind
	TokKindLitRune    // eg. 'ö' or '\''
	TokKindLitStr     // eg. "foo:\"bar\"" or `bar:"baz"`
	TokKindLitInt     // eg. 123 or -321
	TokKindLitFloat   // eg. 12.3 or -3.21
)

// only called by `EnsureSrcFile`
func tokenize(srcFilePath string, curFullSrcFileContent string) (ret Toks, errs SrcFileNotices) {
	if len(curFullSrcFileContent) == 0 {
		return
	}

	var scan scanner.Scanner
	scan.Init(strings.NewReader(curFullSrcFileContent))
	scan.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanChars | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanComments
	scan.Error = func(_ *scanner.Scanner, msg string) {
		errs.Add(&SrcFileNotice{Kind: NoticeKindErr, Code: NoticeCodeLexingError,
			Message: errMsg(NoticeCodeLexingError, msg), Span: (&SrcFilePos{Line: scan.Line, Char: scan.Column}).ToSpan()})
	}
	var last_ident_first_char rune
	var prev *Tok
	scan.IsIdentRune = func(char rune, i int) bool {
		last_ident_first_char = util.If(i == 0, char, last_ident_first_char)
		return (char == '_') || unicode.IsLetter(char) ||
			// if at start of token: ident can begin with `@` or `:` if separate from prev, so `foo:bar` will not make `:bar` but `foo :bar` will
			((i == 0) && ((char == '@') || ((char == ':') && ((prev == nil) || (prev.Kind == TokKindIdentOpish) || (prev.isBracketingOpening(0)) || ((prev.byteOffset + len(prev.Src)) < scan.Offset))))) ||
			// not at start of token: also allow numbers, or `/` if ident started uppercase
			((i > 0) && (unicode.IsDigit(char) || (unicode.IsUpper(last_ident_first_char) && (char == '/'))))
	}
	scan.Filename = srcFilePath

	ret = make(Toks, 0, len(curFullSrcFileContent)/3)
	var had_ws_err bool
	var brac_level int
	var stack []int
	for lexeme := scan.Scan(); lexeme != scanner.EOF; lexeme = scan.Scan() {
		tok := &Tok{Pos: SrcFilePos{Line: scan.Line, Char: scan.Column}, byteOffset: scan.Offset}
		tok.Src = curFullSrcFileContent[tok.byteOffset : tok.byteOffset+len(scan.TokenText())] // to avoid all those string copies we'd have if we just did tok.Src=scan.TokenText()
		switch lexeme {
		case scanner.Int:
			tok.Kind = TokKindLitInt
		case scanner.Float:
			tok.Kind = TokKindLitFloat
		case scanner.Char:
			tok.Kind = TokKindLitRune
		case scanner.Comment:
			tok.Kind = TokKindComment
		case scanner.String, scanner.RawString:
			tok.Kind = TokKindLitStr
		case scanner.Ident:
			tok.Kind = TokKindIdentWord
			if (prev != nil) && ((prev.Kind == TokKindLitFloat) || (prev.Kind == TokKindLitInt)) && tok.isWhitespacelesslyRightAfter(prev) {
				errs = append(errs, tok.newErr(NoticeCodeLexingError, "separate `"+prev.Src+"` from `"+tok.Src+"`"))
			}
		case '(', ')', '{', '}', '[', ']':
			tok.Kind = TokKindBracketing
		default:
			tok.Kind = TokKindIdentOpish
		}

		if prev == nil { // we're at first token in source
			stack = append(stack, tok.Pos.Char)
			ret = append(ret, &Tok{Kind: TokKindBegin, byteOffset: tok.byteOffset, Pos: tok.Pos, Src: tok.Src})
			if tok.Pos.Char > 1 {
				ret, errs = append(ret, tok), append(errs, tok.newIndentErr())
				return
			}
		} else if is_new_line := (brac_level <= 0) && (tok.Pos.Line > prev.Pos.Line); is_new_line {
			// on newline: indent/dedent/newline handling, taken from https://docs.python.org/3/reference/lexical_analysis.html#indentation
			stack_top := stack[len(stack)-1]
			if tok.Pos.Char < stack_top {
				for ; stack_top > tok.Pos.Char; stack_top = stack[len(stack)-1] {
					stack = stack[:len(stack)-1]
					ret = append(ret, &Tok{Kind: TokKindEnd, byteOffset: tok.byteOffset, Pos: tok.Pos, Src: tok.Src})
				}
				if stack_top != tok.Pos.Char {
					errs.Add(tok.newIndentErr())
				}
				ret = append(ret, &Tok{Kind: TokKindEnd, byteOffset: tok.byteOffset, Pos: tok.Pos, Src: tok.Src})
				ret = append(ret, &Tok{Kind: TokKindBegin, byteOffset: tok.byteOffset, Pos: tok.Pos, Src: tok.Src})
			} else if tok.Pos.Char > stack_top {
				stack = append(stack, tok.Pos.Char)
				ret = append(ret, &Tok{Kind: TokKindBegin, byteOffset: tok.byteOffset, Pos: tok.Pos, Src: tok.Src})
			} else {
				ret = append(ret, &Tok{Kind: TokKindEnd, byteOffset: tok.byteOffset, Pos: tok.Pos, Src: tok.Src})
				ret = append(ret, &Tok{Kind: TokKindBegin, byteOffset: tok.byteOffset, Pos: tok.Pos, Src: tok.Src})
			}
			// also on newline: check for any carriage-return or leading tabs since last tok
			src_since_prev := curFullSrcFileContent[prev.byteOffset+len(prev.Src) : tok.byteOffset]
			if (!had_ws_err) && str.Idx(src_since_prev, '\r') >= 0 {
				had_ws_err, errs = true, append(errs, tok.newErr(NoticeCodeWhitespace))
			}
			src_since_prev = src_since_prev[1+str.Idx(src_since_prev, '\n'):]
			if (!had_ws_err) && str.Idx(src_since_prev, '\t') >= 0 {
				had_ws_err, errs = true, append(errs, tok.newErr(NoticeCodeWhitespace))
			}
		}

		// only now can the brac_level be adjusted
		if tok.Kind == TokKindBracketing {
			brac_level += util.If(tok.isBracketingOpening(0), 1, -1)
		}

		switch {
		default:
			ret = append(ret, tok)
		case (prev != nil) && (prev.Kind == TokKindIdentOpish) && (tok.Kind == TokKindIdentOpish) &&
			(!prev.isSep()) && (!tok.isSep()) && ((prev.Pos.Char + len(prev.Src)) == tok.Pos.Char):
			// multi-char op toks such as `!=` are at this point single-char toks ie. '!', '='. we stitch them together:
			prev.Src += tok.Src
			continue // to avoid the further-below setting of `prev = tok` in this `case`
		case ((tok.Kind == TokKindLitFloat) && str.Ends(tok.Src, ".")):
			// split dot-ending float toks like `10.` into 2 toks (int then dot), to allow for dot-methods on int literals like `10.timesDo fn` etc.
			dot := &Tok{
				Kind:       TokKindIdentOpish,
				byteOffset: tok.byteOffset + (len(tok.Src) - 1),
				Pos:        SrcFilePos{Line: tok.Pos.Line, Char: tok.Pos.Char + (len(tok.Src) - 1)},
				Src:        tok.Src[len(tok.Src)-1:],
			}
			tok.Kind, tok.Src = TokKindLitInt, tok.Src[:len(tok.Src)-1]
			ret = append(ret, tok, dot)
			tok = dot // so that `prev` will be correct
			// case (prev != nil) && ((tok.Kind == TokKindLitInt) || (tok.Kind == TokKindLitFloat)) && (prev.Src == "-") && tok.isWhitespacelesslyRightAfter(prev):
		}

		prev = tok
	}

	for len(stack) > 0 {
		stack = stack[:len(stack)-1]
		ret = append(ret, &Tok{Kind: TokKindEnd, byteOffset: prev.byteOffset + len(prev.Src), Src: "",
			Pos: SrcFilePos{Line: prev.Pos.Line, Char: prev.Pos.Char + utf8.RuneCountInString(prev.Src)}})
	}

	return
}

func (me *Tok) bracketingMatch() rune {
	if len(me.Src) > 0 {
		switch me.Src[0] {
		case '(':
			return ')'
		case '[':
			return ']'
		case '{':
			return '}'
		case ')':
			return '('
		case ']':
			return '['
		case '}':
			return '{'
		}
	}
	return 0
}
func (me *Tok) isBracketing() bool { return me.isBracketingClosing(0) || me.isBracketingOpening(0) }
func (me *Tok) isBracketingClosing(open rune) bool {
	if len(me.Src) == 0 {
		return false
	}
	switch open {
	case '(':
		return me.Src[0] == ')'
	case '[':
		return me.Src[0] == ']'
	case '{':
		return me.Src[0] == '}'
	}
	return (me.Src[0] == ')') || (me.Src[0] == ']') || (me.Src[0] == '}')
}
func (me *Tok) isBracketingOpening(close rune) bool {
	if len(me.Src) == 0 {
		return false
	}
	switch close {
	case ')':
		return me.Src[0] == '('
	case ']':
		return me.Src[0] == '['
	case '}':
		return (me.Src[0] == '{')
	}
	return (me.Src[0] == '(') || (me.Src[0] == '[') || (me.Src[0] == '{')
}
func (me *Tok) isBracketingMatch(it *Tok) bool {
	return (len(me.Src) > 0) && ((me.Src[0] == '(' && it.Src[0] == ')') || (me.Src[0] == '[' && it.Src[0] == ']') || (me.Src[0] == '{' && it.Src[0] == '}'))
}

func (me *Tok) isSep() bool {
	return (len(me.Src) == 1) && ((me.Src[0] == ',') || (me.Src[0] == ':'))
}

func (me *Tok) isWhitespacelesslyRightAfter(it *Tok) bool {
	return me.byteOffset == (it.byteOffset + len(it.Src))
}

func (me *Tok) newErr(code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return &SrcFileNotice{Kind: NoticeKindErr, Code: code, Span: me.span(), Message: errMsg(code, args...)}
}

func (me *Tok) newIndentErr() *SrcFileNotice {
	return me.newErr(NoticeCodeIndentation)
}

func (me *Tok) span() (ret SrcFileSpan) {
	ret.Start, ret.End = me.Pos, me.Pos
	for _, r := range me.Src {
		if r == '\n' {
			ret.End.Line, ret.End.Char = ret.End.Line+1, 1
		} else {
			ret.End.Char += len(string(r))
		}
	}
	return
}

func (me Toks) bracketingMatch() (inner Toks, tail Toks, err *SrcFileNotice) {
	var level int
	brac_open := rune(me[0].Src[0])
	brac_close := me[0].bracketingMatch()
	if (brac_close != 0) && me[0].isBracketingOpening(brac_close) {
		for i, tok := range me {
			if tok.isBracketingOpening(brac_close) {
				level++
			} else if tok.isBracketingClosing(brac_open) {
				level--
				if level == 0 {
					if !me[0].isBracketingMatch(tok) {
						break
					}
					return me[1:i], me[i+1:], nil
				}
			}
		}
	}
	return nil, nil, &SrcFileNotice{Kind: NoticeKindErr, Span: me.Span(), Code: NoticeCodeBracketingMismatch,
		Message: errMsg(NoticeCodeBracketingMismatch,
			util.If((me[0].Src[0] == '(') || (me[0].Src[0] == ')'), "parens",
				util.If((me[0].Src[0] == '[') || (me[0].Src[0] == ']'), "brackets",
					"braces")))}
}

func (me Toks) newDiag(kind SrcFileNoticeKind, atEnd bool, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return &SrcFileNotice{Kind: kind, Code: code, Span: util.If(atEnd, Toks.SpanEnd, Toks.Span)(me), Message: errMsg(code, args...)}
}
func (me Toks) newDiagInfo(atEnd bool, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(NoticeKindInfo, atEnd, code, args...)
}
func (me Toks) newDiagHint(atEnd bool, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(NoticeKindHint, atEnd, code, args...)
}
func (me Toks) newDiagWarn(atEnd bool, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(NoticeKindWarn, atEnd, code, args...)
}
func (me Toks) newDiagErr(atEnd bool, code SrcFileNoticeCode, args ...any) *SrcFileNotice {
	return me.newDiag(NoticeKindErr, atEnd, code, args...)
}

func (me Toks) Span() (ret SrcFileSpan) {
	ret.Start, ret.End = me[0].Pos, me[len(me)-1].span().End
	return
}

func (me Toks) SpanEnd() (ret SrcFileSpan) {
	return me[len(me)-1].span().End.ToSpan()
}

func (me Toks) split(delimChar byte) (ret []Toks) {
	var idx int
	var brac_level int
	for i, tok := range me {
		if (tok.Src[0] == delimChar) && (len(tok.Src) == 1) && (brac_level == 0) {
			ret = append(ret, me[idx:i])
			idx = i + 1
		} else if tok.isBracketingOpening(0) {
			brac_level++
		} else if tok.isBracketingClosing(0) {
			brac_level--
		}
	}
	if idx == 0 {
		ret = []Toks{me}
	} else if rest := me[idx:]; len(rest) > 0 {
		ret = append(ret, rest)
	}
	return
}

func (me Toks) src(curFullSrcFileContent string) string {
	if len(me) == 0 {
		return ""
	}
	first, last := me[0], me[len(me)-1]
	return curFullSrcFileContent[first.byteOffset:(last.byteOffset + len(last.Src))]
}

func (me Toks) str() string { // only for occasional debug prints
	return strings.Join(sl.To(me, func(it *Tok) string { return it.Src }), " ")
}
