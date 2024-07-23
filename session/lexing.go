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

// SrcFilePos Line and Char both start at 1
type SrcFilePos struct {
	// Line starts at 1
	Line int
	// Char starts at 1
	Char int
}

func (me *SrcFilePos) after(it *SrcFilePos) bool {
	return util.If(me.Line == it.Line, me.Char > it.Char, me.Line > it.Line)
}
func (me *SrcFilePos) afterOrAt(it *SrcFilePos) bool {
	return util.If(me.Line == it.Line, me.Char >= it.Char, me.Line > it.Line)
}
func (me *SrcFilePos) before(it *SrcFilePos) bool {
	return util.If(me.Line == it.Line, me.Char < it.Char, me.Line < it.Line)
}
func (me *SrcFilePos) beforeOrAt(it *SrcFilePos) bool {
	return util.If(me.Line == it.Line, me.Char <= it.Char, me.Line < it.Line)
}

func (me SrcFilePos) ToSpan() (ret SrcFileSpan) {
	ret.Start, ret.End = me, me
	return
}

type SrcFileSpan struct {
	Start SrcFilePos
	End   SrcFilePos
}

func (me SrcFileSpan) contains(it *SrcFilePos) bool {
	return it.afterOrAt(&me.Start) && it.beforeOrAt(&me.End)
}

func (me *SrcFileSpan) isSinglePos() bool { return me.Start == me.End }

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
	TokKindComment // both /* multi-line */ and // single-line
	TokKindBrace   // parens, square brackets, curly braces
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
	scan.IsIdentRune = func(char rune, i int) bool {
		last_ident_first_char = util.If(i == 0, char, last_ident_first_char)
		return (char == '_') || unicode.IsLetter(char) ||
			((i == 0) && (char == '@')) ||
			((i > 0) && (unicode.IsDigit(char) || (unicode.IsUpper(last_ident_first_char) && (char == '/'))))
	}
	scan.Filename = srcFilePath

	ret = make(Toks, 0, len(curFullSrcFileContent)/3)
	var prev *Tok
	var had_ws_err bool
	var brace_level int
	var stack []int
	for lexeme := scan.Scan(); lexeme != scanner.EOF; lexeme = scan.Scan() {
		tok := &Tok{Pos: SrcFilePos{Line: scan.Line, Char: scan.Column}, byteOffset: scan.Offset}
		tok.Src = curFullSrcFileContent[tok.byteOffset : tok.byteOffset+len(scan.TokenText())] // to avoid all those string copies we'd have if we just did tok.Src=scan.TokenText()
		switch lexeme {
		case scanner.Char:
			tok.Kind = TokKindLitRune
		case scanner.Comment:
			tok.Kind = TokKindComment
		case scanner.Int:
			tok.Kind = TokKindLitInt
		case scanner.Float:
			tok.Kind = TokKindLitFloat
		case scanner.String, scanner.RawString:
			tok.Kind = TokKindLitStr
		case scanner.Ident:
			tok.Kind = TokKindIdentWord
			if (prev != nil) && ((prev.Kind == TokKindLitFloat) || (prev.Kind == TokKindLitInt)) && tok.isWhitespacelesslyRightAfter(prev) {
				errs = append(errs, tok.newErr(NoticeCodeLexingError, "separate `"+prev.Src+"` from `"+tok.Src+"`"))
			}
		case '(', ')', '{', '}', '[', ']':
			tok.Kind = TokKindBrace
		default: // in case we want back to case-of-op, here's what we had: '<', '>', '+', '-', '*', '/', '\\', '^', '~', '×', '÷', '…', '·', '.', '|', '&', '!', '?', '%', '=':
			tok.Kind = TokKindIdentOpish
		}

		if prev == nil { // we're at first token in source
			stack = append(stack, tok.Pos.Char)
			ret = append(ret, &Tok{Kind: TokKindBegin, byteOffset: tok.byteOffset, Pos: tok.Pos, Src: tok.Src})
			if tok.Pos.Char > 1 {
				ret, errs = append(ret, tok), append(errs, tok.newIndentErr())
				return
			}
		} else if is_new_line := (brace_level <= 0) && (tok.Pos.Line > prev.Pos.Line); is_new_line {
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

		// only now can the brace_level be adjusted if needed
		if tok.Kind == TokKindBrace {
			brace_level += util.If(tok.isBraceOpening(0), 1, -1)
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

func (me *Tok) braceMatch() rune {
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
func (me *Tok) isBraceClosing(open rune) bool {
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
func (me *Tok) isBraceOpening(close rune) bool {
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
func (me *Tok) isBraceMatch(it *Tok) bool {
	return (len(me.Src) > 0) && ((me.Src[0] == '(' && it.Src[0] == ')') || (me.Src[0] == '[' && it.Src[0] == ']') || (me.Src[0] == '{' && it.Src[0] == '}'))
}

func (me *Tok) isSep() bool {
	return (len(me.Src) == 1) && ((me.Src[0] == ',') || (me.Src[0] == ';') || (me.Src[0] == ':') || (me.Src[0] == '.'))
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

func (me Toks) braceMatch() (inner Toks, tail Toks, err *SrcFileNotice) {
	var level int
	brace_open := rune(me[0].Src[0])
	brace_close := me[0].braceMatch()
	if (brace_close != 0) && me[0].isBraceOpening(brace_close) {
		for i, tok := range me {
			if tok.isBraceOpening(brace_close) {
				level++
			} else if tok.isBraceClosing(brace_open) {
				level--
				if level == 0 {
					if !me[0].isBraceMatch(tok) {
						break
					}
					return me[1:i], me[i+1:], nil
				}
			}
		}
	}
	return nil, nil, &SrcFileNotice{Kind: NoticeKindErr, Span: me.Span(), Code: NoticeCodeBracesMismatch,
		Message: errMsg(NoticeCodeBracesMismatch,
			util.If((me[0].Src[0] == '(') || (me[0].Src[0] == ')'), "parens",
				util.If((me[0].Src[0] == '[') || (me[0].Src[0] == ']'), "brackets",
					"braces")))}
}

func (me Toks) Span() (ret SrcFileSpan) {
	ret.Start, ret.End = me[0].Pos, me[len(me)-1].span().End
	return
}

func (me Toks) SpanEnd() (ret SrcFileSpan) {
	return me[len(me)-1].span().End.ToSpan()
}

func (me Toks) src(curFullSrcFileContent string) string {
	if len(me) == 0 {
		return ""
	}
	first, last := me[0], me[len(me)-1]
	return curFullSrcFileContent[first.byteOffset:(last.byteOffset + len(last.Src))]
}

func (me Toks) str() string { // only for occasional debug prints
	return strings.Join(sl.As(me, func(it *Tok) string { return it.Src }), " ")
}
