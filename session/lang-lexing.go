package session

import (
	"strings"
	"text/scanner"
	"unicode"

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
type TokKind int

const (
	_              TokKind = iota
	TokKindComment         // both /* multi-line */ and // single-line
	TokKindBrace           // parens, square brackets, curly braces
	TokKindSep             // comma
	// below: only toks that, if no sep-or-ws between them, will `huddle` together
	// into their own single contiguous expr as if parensed (above: those that won't)
	TokKindIdentWord  // lexemes that pass the `IsIdentRune` predicate below
	TokKindIdentOpish // all lexemes that dont match any other TokKind
	TokKindLitRune    // eg. 'ö' or '\''
	TokKindLitStr     // eg. "foo:\"bar\"" or `bar:"baz"`
	TokKindLitInt     // eg. 123 or -321
	TokKindLitFloat   // eg. 12.3 or -3.21
)

type Tok struct {
	byteOffset int
	Pos        SrcFilePos
	Kind       TokKind
	Src        string
}

// only called by `EnsureSrcFile`
func (me *SrcFile) tokenize() (toks Toks, toksChunks toksChunks, errs []*SrcFileNotice) {
	toks = make(Toks, 0, len(me.Content.Src)/3)
	var scan scanner.Scanner
	scan.Init(strings.NewReader(me.Content.Src))
	scan.Whitespace = 1<<'\n' | 1<<' '
	scan.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanChars | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanComments
	scan.Error = func(_ *scanner.Scanner, msg string) {
		errs = append(errs, &SrcFileNotice{Kind: NoticeKindErr, Message: msg, Code: NoticeCodeLexingError,
			Span: (&SrcFilePos{Line: scan.Line, Char: scan.Column}).ToSpan()})
	}
	var last_ident_first_char rune
	scan.IsIdentRune = func(char rune, i int) bool {
		last_ident_first_char = util.If(i == 0, char, last_ident_first_char)
		return (char == '_') || unicode.IsLetter(char) ||
			((i == 0) && ((char == '#') || (char == '%') || (char == '@') || (char == '$'))) ||
			((i > 0) && (unicode.IsDigit(char) || (unicode.IsUpper(last_ident_first_char) && (char == '/'))))
	}
	scan.Filename = me.FilePath

	for lexeme := scan.Scan(); lexeme != scanner.EOF; lexeme = scan.Scan() {
		tok := Tok{Pos: SrcFilePos{Line: scan.Line, Char: scan.Column}, byteOffset: scan.Offset}
		tok.Src = me.Content.Src[tok.byteOffset : tok.byteOffset+len(scan.TokenText())] // to avoid all those string copies we'd have if we just did tok.Src=scan.TokenText()
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
		case '(', ')', '{', '}', '[', ']':
			tok.Kind = TokKindBrace
		case ',':
			tok.Kind = TokKindSep
		default: // in case we want back to case-of-op, here's what we had: '<', '>', '+', '-', '*', '/', '\\', '^', '~', '×', '÷', '…', '·', '.', '|', '&', '!', '?', '%', '=':
			tok.Kind = TokKindIdentOpish
		}
		toks = append(toks, &tok)
	}

	// split dot-ending float toks like `10.` into 2 int-then-dot toks, to allow for dot-methods on int literals like `10.timesDo fn` etc.
	for i := 0; i < len(toks); i++ {
		if tok := toks[i]; tok.Kind == TokKindLitFloat && str.Ends(tok.Src, ".") {
			toks = append(toks[:i+1], append(Toks{{
				byteOffset: tok.byteOffset + (len(tok.Src) - 1),
				Pos:        SrcFilePos{Line: tok.Pos.Line, Char: tok.Pos.Char + (len(tok.Src) - 1)},
				Kind:       TokKindIdentOpish,
				Src:        tok.Src[len(tok.Src)-1:]},
			}, toks[i+1:]...)...)
			tok.Kind = TokKindLitInt
			tok.Src = tok.Src[:len(tok.Src)-1]
		}
	}

	// multi-char op chars such as `!=` are at this point single-char toks ie. '!', '='. we stitch them together:
	for i := 1; i < len(toks); i++ {
		if (toks[i-1].Kind == TokKindIdentOpish) && (toks[i].Kind == TokKindIdentOpish) && ((toks[i-1].Pos.Char + len(toks[i-1].Src)) == toks[i].Pos.Char) {
			toks[i-1].Src += toks[i].Src
			toks = append(toks[:i], toks[i+1:]...)
			i--
		}
	}

	if len(toks) > 0 && toks[0].Pos.Char > 1 {
		me.Notices.LexErrs = append(me.Notices.LexErrs, toks[0].newIndentErr())
	}
	if len(errs) == 0 {
		var err *SrcFileNotice
		toksChunks, err = toksChunked(toks)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return
}

func (me *Tok) braceMatch() rune {
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
	return 0
}
func (me *Tok) isBraceClosing(open rune) bool {
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
	return (me.Src[0] == '(' && it.Src[0] == ')') || (me.Src[0] == '[' && it.Src[0] == ']') || (me.Src[0] == '{' && it.Src[0] == '}')
}

func (me Toks) huddle() (huddled Toks, rest Toks) {
	var idx_until int
	for i := 1; i < len(me); i++ {
		cur, prev := me[i], me[i-1]
		if cur.byteOffset == (prev.byteOffset+len(prev.Src)) && cur.Kind >= TokKindIdentOpish && prev.Kind >= TokKindIdentOpish {
			idx_until = i + 1
		} else {
			break
		}
	}
	return me[:idx_until], me[idx_until:]
}

func (me *Tok) newErr(code SrcFileNoticeCode, msg string) *SrcFileNotice {
	return &SrcFileNotice{Kind: NoticeKindErr, Code: code, Span: me.span(), Message: msg}
}

func (me *Tok) newIndentErr() *SrcFileNotice {
	return me.newErr(NoticeCodeIndentation, "ambiguous indentation")
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

func (me Toks) allOfKind(kind TokKind) bool {
	return sl.All(me, func(it *Tok) bool { return it.Kind == kind })
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
	err_msg := "no matching opening and closing " +
		util.If((me[0].Src[0] == '(') || (me[0].Src[0] == ')'), "parens",
			util.If((me[0].Src[0] == '[') || (me[0].Src[0] == ']'), "brackets",
				"braces"))
	return nil, nil, &SrcFileNotice{Kind: NoticeKindErr, Span: me.Span(), Code: NoticeCodeBracesMismatch, Message: err_msg}
}

func (me Toks) isMultiLine() bool {
	return (me[len(me)-1].Pos.Line > me[0].Pos.Line)
}

func (me Toks) newErr(code SrcFileNoticeCode, msg string) *SrcFileNotice {
	return &SrcFileNotice{Kind: NoticeKindErr, Code: code, Span: me.Span(), Message: msg}
}

func (me Toks) Span() (ret SrcFileSpan) {
	ret.Start, ret.End = me[0].Pos, me[len(me)-1].span().End
	return
}

func (me Toks) split(by TokKind) (ret []Toks) {
	var cur Toks
	var skip int
	for _, tok := range me {
		if tok.Kind == TokKindBrace {
			skip = util.If(tok.isBraceOpening(0), skip+1, skip-1)
		}
		if (skip <= 0) && tok.Kind == by {
			ret, cur = append(ret, cur), nil
		} else {
			cur = append(cur, tok)
		}
	}
	if len(cur) > 0 {
		ret = append(ret, cur)
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
	return strings.Join(sl.As(me, func(it *Tok) string { return it.Src }), " ")
}

func (me Toks) withoutComments() Toks {
	return sl.Where(me, func(it *Tok) bool { return it.Kind != TokKindComment })
}

type toksChunks []*toksChunk

type toksChunk struct {
	self Toks
	subs toksChunks
	full Toks
}

func (me *toksChunk) str(indent int) (ret string) {
	ret = str.Repeat(" ", indent) + ">" + str.FromInt(indent) + ">" + me.self.str()
	for _, it := range me.subs {
		ret += "\n" + it.str(indent+2)
	}
	ret += "<" + str.FromInt(indent) + "<"
	return
}

func (me toksChunks) str() string {
	return str.Join(sl.As(me, func(it *toksChunk) string { return it.str(0) }), "\n")
}

func toksChunked(toks Toks) (ret toksChunks, err *SrcFileNotice) {
	if len(toks) == 0 {
		return nil, nil
	}
	pos_char, pos_line, last_line_start :=
		toks[0].Pos.Char, toks[0].Pos.Line, toks[0].Pos
	cur := &toksChunk{}
	var cur_indent_toks Toks
	cur_done := func(nextChunkStartTok *Tok) {
		if len(cur.self) > 0 {
			cur.subs, err = toksChunked(cur_indent_toks)
			ret = append(ret, cur)
		}
		if nextChunkStartTok != nil {
			cur, cur_indent_toks, pos_line =
				&toksChunk{self: Toks{nextChunkStartTok}, full: Toks{nextChunkStartTok}}, nil, nextChunkStartTok.Pos.Line
		}
	}

	// TODO: replace one-by-one appends with gathering idxs for sub-slicing from `toks`
	for len(toks) > 0 {
		tok := toks[0]
		if tok.Pos.Line > last_line_start.Line {
			last_line_start = tok.Pos
		}
		switch {
		case tok.isBraceOpening(0): // not indent-chunking parens (just yet), brackets or braces
			inner_toks, rest_toks, err := toks.braceMatch()
			if err != nil {
				return nil, err
			}
			full_brace_toks := toks[0 : len(inner_toks)+2]
			if tok.Pos.Line == pos_line {
				cur.self = append(cur.self, full_brace_toks...)
			} else {
				cur_indent_toks = append(cur_indent_toks, full_brace_toks...)
			}
			cur.full = append(cur.full, full_brace_toks...)
			toks = rest_toks
			continue

		case tok.Pos.Line == pos_line:
			cur.self, cur.full = append(cur.self, tok), append(cur.full, tok)
		case tok.Pos.Char < pos_char:
			return nil, tok.newIndentErr()
		case tok.Pos.Char > pos_char:
			cur_indent_toks, cur.full = append(cur_indent_toks, tok), append(cur.full, tok)
		case tok.Pos.Char == pos_char:
			if cur_done(tok); err != nil {
				return
			}
		}
		toks = toks[1:]
	}

	cur_done(nil)
	return
}
