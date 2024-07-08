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

func (me SrcFilePos) ToSpan() (ret SrcFileSpan) {
	ret.Start, ret.End = me, me
	return
}

type SrcFileSpan struct {
	Start SrcFilePos
	End   SrcFilePos
}

func (me *SrcFileSpan) IsSinglePos() bool { return me.Start == me.End }

type ToksChunks []Toks
type Toks []*Tok
type TokKind int

const (
	TokKindInvalid TokKind = iota
	TokKindBrace
	TokKindOp
	TokKindSep
	TokKindIdent
	TokKindComment
	TokKindLitChar
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

func tokenize(src string, filePath string) (ret ToksChunks, errs []*SrcFileNotice) {
	var scan scanner.Scanner
	scan.Init(strings.NewReader(src))
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
			((i == 0) && ((char == '%') || (char == '@') || (char == '$') || (char == '.'))) ||
			((i > 0) && (unicode.IsDigit(char) || (unicode.IsUpper(last_ident_first_char) && (char == '/'))))
	}
	scan.Filename = filePath

	var flat_list Toks
	for lexeme := scan.Scan(); lexeme != scanner.EOF; lexeme = scan.Scan() {
		tok := Tok{Pos: SrcFilePos{Line: scan.Line, Char: scan.Column}, byteOffset: scan.Offset, Src: scan.TokenText()}
		switch lexeme {
		case scanner.Char:
			tok.Kind = TokKindLitChar
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
		case ',', '.':
			tok.Kind = TokKindSep
		case ':':
			next_rune := scan.Peek()
			tok.Kind = util.If((next_rune == ' ') || (next_rune == '\n'), TokKindSep, TokKindOp)
		case '<', '>', '+', '-', '*', '/', '\\', '^', '~', '×', '÷', '…', '·', '|', '&', '!', '?', '%', '=':
			tok.Kind = TokKindOp
		default:
			errs = append(errs, &SrcFileNotice{Kind: NoticeKindErr, Code: NoticeCodeLexingError,
				Span: (&SrcFilePos{Line: scan.Line, Char: scan.Column}).ToSpan(), Message: str.Fmt("unknown lexeme: '%s'", string(lexeme))})
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

	ret = flat_list.chunks(1)
	return
}

func (me Toks) chunks(posCharIndentedIsGt int) (ret ToksChunks) {
	var cur_chunk Toks
	for i, tok := range me {
		is_on_a_new_line := (i == 0) || (tok.Pos.Line != me[i-1].Pos.Line)
		if is_on_a_new_line && (tok.Pos.Char == posCharIndentedIsGt) {
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

func (me Toks) allOfKind(kind TokKind) bool {
	return sl.All(me, func(it *Tok) bool { return it.Kind == kind })
}

func (me Toks) withoutComments() Toks {
	return sl.Where(me, func(it *Tok) bool { return it.Kind != TokKindComment })
}
