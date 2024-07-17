package session

import (
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
	"cmp"
	"slices"
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
	OnDbgMsg         = func(bool, string, ...any) {}
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

func (me *SrcFileNotice) String() string { return me.Message }

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

// callers have already `sharedState.Lock`ed
func refreshAndPublishNotices(provokingFilePaths ...string) {
	if len(provokingFilePaths) == 0 {
		return
	}
	new_notices := map[string]SrcFileNotices{}

	for _, src_file_path := range provokingFilePaths {
		var file_notices SrcFileNotices
		if src_file := state.srcFiles[src_file_path]; src_file != nil {
			has_brace_err := src_file.Content.Ast.has(true, func(node *AstNode) bool {
				return (node.errParsing != nil) && (node.errParsing.Code == NoticeCodeBracesMismatch)
			})
			if src_file.notices.LastReadErr != nil {
				file_notices.Add(src_file.notices.LastReadErr)
			}
			file_notices.Add(src_file.notices.LexErrs...)
			src_file.Content.Ast.walk(nil, func(node *AstNode) {
				if node.errParsing != nil {
					file_notices.Add(node.errParsing)
				}
				if !has_brace_err {
					file_notices.Add(node.errsExpansion...)
				}
			})
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
