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

type ToksChunks []Toks
type Toks []*Tok
type TokKind int

const (
	TokKindErr TokKind = iota
	TokKindBrace
	TokKindSep
	TokKindOp
	TokKindIdent
	TokKindComment
	TokKindLitRune
	TokKindLitStr
	TokKindLitInt
	TokKindLitFloat
)

type Tok struct {
	byteOffset int
	Pos        SrcFilePos
	Kind       TokKind
	Src        string
}

// only called by `EnsureSrcFile`
func (me *SrcFile) tokenize() (ret ToksChunks, errs []*SrcFileNotice) {
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
			((i == 0) && ((char == '#') || (char == '%') || (char == '@') || (char == '$') || (char == '.'))) ||
			((i > 0) && (unicode.IsDigit(char) || (unicode.IsUpper(last_ident_first_char) && (char == '/'))))
	}
	scan.Filename = me.FilePath

	var flat_list Toks
	for lexeme := scan.Scan(); lexeme != scanner.EOF; lexeme = scan.Scan() {
		tok := Tok{Pos: SrcFilePos{Line: scan.Line, Char: scan.Column}, byteOffset: scan.Offset, Src: scan.TokenText()}
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
			tok.Kind = TokKindIdent
		case '(', ')', '{', '}', '[', ']':
			tok.Kind = TokKindBrace
		case ',':
			tok.Kind = TokKindSep
		default: // in case we want back to case-of-op, here's what we had: '<', '>', '+', '-', '*', '/', '\\', '^', '~', '×', '÷', '…', '·', '.', '|', '&', '!', '?', '%', '=':
			tok.Kind = TokKindOp
		}
		flat_list = append(flat_list, &tok)
	}

	// multi-char op-likes such as `!=` are at this point single-char toks ie. '!', '='. we stitch them together:
	for i := 1; i < len(flat_list); i++ {
		if (flat_list[i-1].Kind == TokKindOp) && (flat_list[i].Kind == TokKindOp) && ((flat_list[i-1].Pos.Char + len(flat_list[i-1].Src)) == flat_list[i].Pos.Char) {
			flat_list[i-1].Src += flat_list[i].Src
			flat_list = append(flat_list[:i], flat_list[i+1:]...)
			i--
		}
	}

	var err *SrcFileNotice
	if ret, err = flat_list.topChunks(); err != nil {
		errs = append(errs, err)
	}
	return
}

func (me *Tok) newIndentErr() *SrcFileNotice {
	return &SrcFileNotice{Kind: NoticeKindErr, Code: NoticeCodeMisindentation, Span: me.Span(), Message: "ambiguous indentation:"}
}
func (me *Tok) isBraceClosing() bool { return str.Has(")]}", me.Src) }
func (me *Tok) isBraceOpening() bool { return str.Has("([{", me.Src) }
func (me *Tok) isBraceMatch(it *Tok) bool {
	return (me.Src == "(" && it.Src == ")") || (me.Src == "[" && it.Src == "]") || (me.Src == "{" && it.Src == "}")
}

func (me *Tok) Span() (ret SrcFileSpan) {
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
	if me[0].isBraceOpening() {
		for i, tok := range me {
			if tok.isBraceOpening() {
				level++
			} else if tok.isBraceClosing() {
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
		util.If((me[0].Src == "(") || (me[0].Src == ")"), "parens",
			util.If((me[0].Src == "[") || (me[0].Src == "]"), "brackets",
				"braces"))
	return nil, nil, &SrcFileNotice{Kind: NoticeKindErr, Span: me.Span(), Code: NoticeCodeBracesMismatch, Message: err_msg}
}

func (me Toks) Span() (ret SrcFileSpan) {
	ret.Start, ret.End = me[0].Pos, me[len(me)-1].Span().End
	return
}

func (me Toks) split(by TokKind) (ret []Toks) {
	var cur Toks
	for _, tok := range me {
		if tok.Kind == by {
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

func (me Toks) subChunks() (head Toks, tail ToksChunks, err *SrcFileNotice) {
	idx_start := sl.IdxWhere(me, func(it *Tok) bool { return it.Pos.Line > me[0].Pos.Line })
	if idx_start < 0 {
		return me, nil, nil
	}

	indent_pos_char := me[idx_start].Pos.Char
	if indent_pos_char <= me[0].Pos.Char {
		return nil, nil, me[idx_start].newIndentErr()
	}

	head = me[:idx_start]
	cur_line := me[idx_start].Pos.Line
	var cur_chunk Toks
	for _, tok := range me[idx_start:] {
		is_new_line := (tok.Pos.Line > cur_line)
		if !is_new_line {
			cur_chunk = append(cur_chunk, tok)
		} else {
			cur_line++
			if tok.Pos.Char > indent_pos_char {
				cur_chunk = append(cur_chunk, tok)
			} else if tok.Pos.Char == indent_pos_char {
				tail, cur_chunk = append(tail, cur_chunk), Toks{tok}
			} else {
				return nil, nil, tok.newIndentErr()
			}
		}
	}
	if len(cur_chunk) > 0 {
		tail = append(tail, cur_chunk)
	}
	if len(tail) == 0 {
		util.Assert(len(me) == len(head), nil)
	}

	return
}

func (me Toks) topChunks() (ret ToksChunks, err *SrcFileNotice) {
	var cur_chunk Toks
	if me[0].Pos.Char != 1 {
		err = me[0].newIndentErr()
		return
	}
	for i, tok := range me {
		is_on_a_new_line := (i == 0) || (tok.Pos.Line != me[i-1].Pos.Line)
		if is_on_a_new_line && (tok.Pos.Char == 1) {
			if len(cur_chunk) > 0 {
				ret = append(ret, cur_chunk)
			}
			cur_chunk = Toks{tok}
		} else {
			cur_chunk = append(cur_chunk, tok)
		}
	}
	if len(cur_chunk) > 0 {
		ret = append(ret, cur_chunk)
	}
	// may now have comments as top-level chunks that should belong to the next non-comment top-level chunk, let's rectify:
	for i := 0; i < len(ret)-1; i++ {
		if ret[i].allOfKind(TokKindComment) && (ret[i+1][0].Pos.Line == (1 + ret[i][len(ret[i])-1].Pos.Line)) {
			ret[i+1] = append(ret[i], ret[i+1]...) // prepend cur chunk to next
			ret = append(ret[:i], ret[i+1:]...)    // remove cur chunk
			i--
		}
	}

	return
}

func (me Toks) withoutComments() Toks {
	return sl.Where(me, func(it *Tok) bool { return it.Kind != TokKindComment })
}

func (me ToksChunks) str() string {
	return str.Fmt("%#v", sl.As(me, func(it Toks) string { return it.str() }))
}
