package session

import (
	"atmo/util/str"
	"strings"
	"text/scanner"
)

// SrcFilePos Line and Char start at 1, for error-msg UX etc.
type SrcFilePos struct {
	Line int
	Char int
}

type SrcFileSpan struct {
	Start SrcFilePos
	End   SrcFilePos
}

func (me *SrcFileSpan) IsEmpty() bool { return me.Start == me.End }

type Toks []Tok

type Tok struct {
	SrcFilePos
	byteOffset int
}

func tokenize(src string, filePath string) (ret Toks, errs []*SrcFileNotice) {
	var scan scanner.Scanner
	scan.Init(strings.NewReader(src))
	scan.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanChars | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanComments
	scan.Error = func(_ *scanner.Scanner, msg string) {
		errs = append(errs, &SrcFileNotice{Kind: NoticeKindErr, Message: msg, Code: NoticeCodeLexingError,
			Span: SrcFileSpan{Start: SrcFilePos{Line: scan.Line, Char: scan.Column}}})
	}
	scan.Filename = filePath
	for tok := scan.Scan(); tok != scanner.EOF; tok = scan.Scan() {
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
		println(str.Fmt(">>>%s>>%s<<", scan.Position, scan.TokenText()))
	}
	return
}
