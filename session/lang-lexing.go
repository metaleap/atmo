package session

import (
	"strings"
	"text/scanner"
	"unicode"

	"atmo/util"
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
type Toks []Tok
type TokKind int

const (
	_ TokKind = iota
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
	Pos        SrcFilePos
	Kind       TokKind
	ByteOffset int
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
	scan.IsIdentRune = func(char rune, i int) bool {
		return ((i == 0) && ((char == '@') || (char == '$') || (char == '.'))) ||
			((i > 0) && (char == '/' || char == '\\' || unicode.IsDigit(char))) ||
			unicode.IsLetter(char)
	}
	scan.Filename = filePath

	var toks_flat Toks
	for lexeme := scan.Scan(); lexeme != scanner.EOF; lexeme = scan.Scan() {
		tok := Tok{Pos: SrcFilePos{Line: scan.Line, Char: scan.Column}, ByteOffset: scan.Offset, Src: scan.TokenText()}
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
		toks_flat = append(toks_flat, tok)
	}
	for i := 1; i < len(toks_flat); i++ {
		if (toks_flat[i-1].Kind == TokKindOp) && (toks_flat[i].Kind == TokKindOp) && ((toks_flat[i-1].Pos.Char + len(toks_flat[i-1].Src)) == toks_flat[i].Pos.Char) {
			toks_flat[i-1].Src += toks_flat[i].Src
			toks_flat = append(toks_flat[:i], toks_flat[i+1:]...)
			i--
		}
	}
	ret = append(ret, toks_flat)
	return
}
