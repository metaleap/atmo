package session

import (
	"cmp"
	"fmt"
	"path/filepath"
	"slices"

	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type SrcFileNoticeKind int

const (
	_ SrcFileNoticeKind = iota
	NoticeKindErr
	NoticeKindWarn
	NoticeKindInfo
	NoticeKindHint
)

type SrcFileNoticeCode string

const (
	NoticeCodeAtmoTodo      SrcFileNoticeCode = "AtmoTodo"
	NoticeCodeFileReadError SrcFileNoticeCode = "FileReadError"

	// lexing
	NoticeCodeWhitespace  SrcFileNoticeCode = "Whitespace"
	NoticeCodeLexingError SrcFileNoticeCode = "LexingError"
	NoticeCodeIndentation SrcFileNoticeCode = "Indentation"

	// parsing
	NoticeCodeBracesMismatch SrcFileNoticeCode = "BracesMismatch"
	NoticeCodeLitSyntax      SrcFileNoticeCode = "LitSyntax"

	// semantic
	NoticeCodeExpectedFooHere SrcFileNoticeCode = "Unexpected"
)

var (
	allNotices       = map[string]SrcFileNotices{}
	OnNoticesChanged = func() {}
	OnDbgMsg         = func(showIf bool, fmt string, args ...any) {}
	errMsgs          = map[SrcFileNoticeCode]string{
		NoticeCodeAtmoTodo:      "TODO by Atmo team, please report: %s",
		NoticeCodeFileReadError: "%s", // actual error msg in %s

		NoticeCodeWhitespace:  "unsupported white-space; ensure both: no line-leading tabs, and LF-only line endings (no CR or CRLF)",
		NoticeCodeLexingError: "invalid token: %s", // actual error msg in %s
		NoticeCodeIndentation: "incorrect indentation",

		NoticeCodeLitSyntax:      "invalid literal: %s", // actual error msg in %s
		NoticeCodeBracesMismatch: "opening and closing %s don't match up",

		NoticeCodeExpectedFooHere: "expected %s here %s",
	}
)

type SrcFileNotices = sl.Of[*SrcFileNotice]
type SrcFileNotice struct {
	Kind    SrcFileNoticeKind
	Message string
	Span    SrcFileSpan
	Code    SrcFileNoticeCode
}

func (me *SrcFileNotice) equals(it *SrcFileNotice) bool {
	return (me == it) || ((me != nil) && (it != nil) &&
		(me.Code == it.Code) && (me.Kind == it.Kind) && (me.Message == it.Message))
}

func (me *SrcFileNotice) Error() string  { return me.String() }
func (me *SrcFileNotice) String() string { return util.JsonFrom(me) }
func (me *SrcFileNotice) LocStr(srcFilePath string) string {
	if tmp, err := filepath.Rel(".", srcFilePath); err != nil && tmp != "" {
		srcFilePath = tmp
	}
	return fmt.Sprintf("%s:%d,%d", srcFilePath, me.Span.Start.Line, me.Span.Start.Char)
}

func errMsg(code SrcFileNoticeCode, args ...any) string {
	return str.Trim(str.Fmt(errMsgs[code], args...))
}

func errToNotice(err error, code SrcFileNoticeCode, span SrcFileSpan) *SrcFileNotice {
	if err == nil {
		return nil
	}
	err_msg, err_msg_fmt := err.Error(), errMsgs[code]
	err_msg = str.Trim(util.If(err_msg_fmt == "", err_msg, str.Fmt(err_msg_fmt, err_msg)))
	return &SrcFileNotice{Kind: NoticeKindErr, Message: err_msg, Code: code, Span: span}
}

func (me *SrcFile) allNotices() (ret sl.Of[*SrcFileNotice]) {
	has_brace_err := me.Src.Ast.has(true, func(node *AstNode) bool {
		return (node.errParsing != nil) && (node.errParsing.Code == NoticeCodeBracesMismatch)
	})
	if me.notices.LastReadErr != nil {
		ret.Add(me.notices.LastReadErr)
	}
	ret.Add(me.notices.LexErrs...)
	me.Src.Ast.walk(nil, func(node *AstNode) {
		if node.errParsing != nil {
			ret.Add(node.errParsing)
		}
		if !has_brace_err {
			ret.Add(node.errsExpansion...)
		}
	})
	return
}

// callers have already `sharedState.Lock`ed
func refreshAndPublishNotices(provokingFilePaths ...string) {
	if len(provokingFilePaths) == 0 {
		return
	}
	new_notices := map[string]SrcFileNotices{}

	for _, src_file_path := range provokingFilePaths {
		var file_notices SrcFileNotices
		if src_file := state.srcFiles[src_file_path]; src_file != nil {
			file_notices.Add(src_file.allNotices()...)
		}
		new_notices[src_file_path] = file_notices
	}

	// sorting is mainly for the later equality-comparison further down below
	for src_file_path := range new_notices {
		new_notices[src_file_path] = sl.SortedPer(new_notices[src_file_path], func(diag1 *SrcFileNotice, diag2 *SrcFileNotice) int {
			if diag1.Span.Start.Line == diag2.Span.Start.Line {
				return cmp.Compare(diag1.Span.Start.Char, diag2.Span.Start.Char)
			}
			return cmp.Compare(diag1.Span.Start.Line, diag2.Span.Start.Line)
		})
	}

	var have_changes bool
	for src_file_path := range allNotices {
		if _, still_exists := state.srcFiles[src_file_path]; !still_exists {
			have_changes = true
			delete(allNotices, src_file_path)
		}
	}
	for src_file_path, new_notices := range new_notices {
		old_notices := allNotices[src_file_path]
		if !slices.EqualFunc(old_notices, new_notices, func(diag1 *SrcFileNotice, diag2 *SrcFileNotice) bool {
			return (diag1 == diag2) || (*diag1 == *diag2)
		}) {
			have_changes = true
			break
		}
	}
	if have_changes {
		for src_file_path, new_notices := range new_notices {
			allNotices[src_file_path] = new_notices
		}
	}

	if have_changes {
		go OnNoticesChanged()
	}
}
