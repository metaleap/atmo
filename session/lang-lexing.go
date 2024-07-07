package session

import (
	"strings"
	"text/scanner"
	"unicode"

	"atmo/util/str"
)

// SrcFilePos Line and Char both start at 1
type SrcFilePos struct {
	Line int
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

func (me *SrcFileSpan) IsEmpty() bool { return me.Start == me.End }

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
	SrcFilePos
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
		errs = append(errs, &SrcFileNotice{Kind: NoticeKindErr, Message: msg, Code: NoticeCodeLexingOtherError,
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
		/*
		   >>>/home/_/c/at/foo.at:1:1>>foo<<
		   >>>/home/_/c/at/foo.at:1:4>>-<<
		   >>>/home/_/c/at/foo.at:1:5>>bar<<
		   >>>/home/_/c/at/foo.at:1:8>>-<<
		   >>>/home/_/c/at/foo.at:1:9>>{<<
		   >>>/home/_/c/at/foo.at:1:10>>baz<<
		   >>>/home/_/c/at/foo.at:1:13>>}<<
		   >>>/home/_/c/at/foo.at:1:15>>:<<
		   >>>/home/_/c/at/foo.at:1:16>>=<<
		   >>>/home/_/c/at/foo.at:2:3>>(<<
		   >>>/home/_/c/at/foo.at:2:4>>"Hello World"<<
		   >>>/home/_/c/at/foo.at:2:17>>[<<
		   >>>/home/_/c/at/foo.at:2:18>>0<<
		   >>>/home/_/c/at/foo.at:2:19>>]<<
		   >>>/home/_/c/at/foo.at:2:20>>)<<
		*/
		tok := Tok{SrcFilePos: SrcFilePos{Line: scan.Line, Char: scan.Column}, ByteOffset: scan.Offset, Src: scan.TokenText()}
		switch lexeme {
		case scanner.Char:
			tok.Kind = TokKindLitChar
		case scanner.Comment:
			tok.Kind = TokKindComment
		case scanner.Int:
			tok.Kind = TokKindLitInt
		case scanner.Float:
			tok.Kind = TokKindLitFloat
		case '(', ')', '{', '}', '[', ']':
			tok.Kind = TokKindBrace
		case scanner.String, scanner.RawString:
			tok.Kind = TokKindLitStr
		case scanner.Ident:
			tok.Kind = TokKindIdent
		case ',', '.':
			tok.Kind = TokKindSep
		case ':':
			if scan.Peek() == ' ' || scan.Peek() == '\n' {
				tok.Kind = TokKindSep
			}
		case '<', '>', '+', '-', '*', '/', '\\', '^', '~', '×', '÷', '…', '·', '|', '&', '!', '?', '%', '=':
			tok.Kind = TokKindOp
		default:
			errs = append(errs, &SrcFileNotice{Kind: NoticeKindErr, Code: NoticeCodeLexingUnknownLexeme,
				Span: (&SrcFilePos{Line: scan.Line, Char: scan.Column}).ToSpan(), Message: str.Fmt("unknown lexeme: '%s'", string(lexeme))})
		}
		toks_flat = append(toks_flat, tok)
	}
	ret = append(ret, toks_flat)
	return
}
