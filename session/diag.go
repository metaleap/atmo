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
	NoticeCodeFileReadError  SrcFileNoticeCode = "FileReadError"
	NoticeCodeWhitespace     SrcFileNoticeCode = "Whitespace"
	NoticeCodeLexingError    SrcFileNoticeCode = "LexingError"
	NoticeCodeLitSyntax      SrcFileNoticeCode = "LitSyntax"
	NoticeCodeIndentation    SrcFileNoticeCode = "Indentation"
	NoticeCodeMisplaced      SrcFileNoticeCode = "Misplaced"
	NoticeCodeBracesMismatch SrcFileNoticeCode = "BracesMismatch"
)

var (
	allNotices       = map[string][]*SrcFileNotice{}
	OnNoticesChanged = func() {}
	OnDbgMsg         = func(bool, string, ...any) {}
	errMsgs          = map[SrcFileNoticeCode]string{
		NoticeCodeFileReadError:  "%s", // actual error msg in %s
		NoticeCodeWhitespace:     "unsupported white-space; ensure both: no line-leading tabs, and LF-only line endings (no CR or CRLF)",
		NoticeCodeLexingError:    "invalid token: %s",   // actual error msg in %s
		NoticeCodeLitSyntax:      "invalid literal: %s", // actual error msg in %s
		NoticeCodeIndentation:    "incorrect indentation",
		NoticeCodeMisplaced:      "unexpected: '%s'",
		NoticeCodeBracesMismatch: "opening and closing %s don't match up",
	}
)

type SrcFileNotice struct {
	Kind    SrcFileNoticeKind
	Message string
	Span    SrcFileSpan
	Code    SrcFileNoticeCode
}

func (me *SrcFileNotice) String() string { return me.Message }

func errToNotice(err error, code SrcFileNoticeCode, span SrcFileSpan) *SrcFileNotice {
	if err == nil {
		return nil
	}
	err_msg, err_msg_fmt := err.Error(), errMsgs[code]
	err_msg = util.If(err_msg_fmt == "", err_msg, str.Fmt(err_msg_fmt, err_msg))
	return &SrcFileNotice{Kind: NoticeKindErr, Message: err_msg, Code: code, Span: span}
}

// callers have already `sharedState.Lock`ed
func refreshAndPublishNotices(provokingFilePaths ...string) {
	new_notices := map[string][]*SrcFileNotice{}

	for _, src_file_path := range provokingFilePaths {
		var file_notices []*SrcFileNotice
		if src_file := state.srcFiles[src_file_path]; src_file != nil {
			if src_file.notices.LastReadErr != nil {
				file_notices = append(file_notices, src_file.notices.LastReadErr)
			}
			file_notices = append(file_notices, src_file.notices.LexErrs...)
			src_file.Content.Ast.walk(nil, func(node *AstNode) {
				if node.err != nil {
					file_notices = append(file_notices, node.err)
				}
			})
			src_file.Content.Est.walk(nil, func(node *EstNode) {
				file_notices = append(file_notices, node.diags...)
			})
		}
		// sorting is mainly for the later equality-comparison further down below
		file_notices = sl.SortedPer(file_notices, func(diag1 *SrcFileNotice, diag2 *SrcFileNotice) int {
			if diag1.Span.Start.Line == diag2.Span.Start.Line {
				return cmp.Compare(diag1.Span.Start.Char, diag2.Span.Start.Char)
			}
			return cmp.Compare(diag1.Span.Start.Line, diag2.Span.Start.Line)
		})
		new_notices[src_file_path] = file_notices
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
