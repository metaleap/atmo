package session

import (
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
	"cmp"
	"maps"
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
	NoticeCodeExprExpected   SrcFileNoticeCode = "ExprExpected"
)

var (
	allNotices       = map[string][]*SrcFileNotice{}
	OnNoticesChanged = func(map[string][]*SrcFileNotice) {}
	OnDbgMsg         = func(bool, string, ...any) {}
	errMsgs          = map[SrcFileNoticeCode]string{
		NoticeCodeFileReadError:  "%s", // actual error msg in %s
		NoticeCodeWhitespace:     "unsupported white-space; ensure both: no leading tabs and only LF (no CR) line endings",
		NoticeCodeLexingError:    "invalid token: %s",   // actual error msg in %s
		NoticeCodeLitSyntax:      "invalid literal: %s", // actual error msg in %s
		NoticeCodeIndentation:    "ambiguous indentation",
		NoticeCodeMisplaced:      "unexpected: '%s'",
		NoticeCodeBracesMismatch: "no matching opening and closing %s",
		NoticeCodeExprExpected:   "expression expected",
	}
)

type SrcFileNotice struct {
	Kind    SrcFileNoticeKind
	Message string
	Span    SrcFileSpan
	Code    SrcFileNoticeCode
}

func (me *SrcFileNotice) Error() string  { return me.Message }
func (me *SrcFileNotice) String() string { return me.Message }

func errToNotice(err error, code SrcFileNoticeCode, span *SrcFileSpan) (ret *SrcFileNotice) {
	if ret, _ = err.(*SrcFileNotice); (ret == nil) && (err != nil) {
		err_msg := errMsgs[code]
		err_msg = util.If(err_msg == "", err.Error(), str.Fmt(err_msg, err.Error()))
		ret = &SrcFileNotice{Kind: NoticeKindErr, Message: err_msg, Code: code}
	}
	if ret != nil && span != nil {
		ret.Span = *span
	}
	return
}

func refreshAndPublishNotices(provokingFilePaths ...string) {
	all_notices := map[string][]*SrcFileNotice{}
	for _, src_file_path := range provokingFilePaths {
		var file_notices []*SrcFileNotice
		if src_file := allSrcFiles[src_file_path]; src_file != nil {
			if src_file.Notices.LastReadErr != nil {
				file_notices = append(file_notices, src_file.Notices.LastReadErr)
			}
			file_notices = append(file_notices, src_file.Notices.LexErrs...)
			for _, top_level_node := range src_file.Content.Ast {
				top_level_node.walk(nil, func(node *AstNode) {
					if node.err != nil {
						file_notices = append(file_notices, node.err)
					}
				})
			}
		}
		// sorting is mainly for the later equality-comparison further down below
		file_notices = sl.SortedPer(file_notices, func(diag1 *SrcFileNotice, diag2 *SrcFileNotice) int {
			if diag1.Span.Start.Line == diag2.Span.Start.Line {
				return cmp.Compare(diag1.Span.Start.Char, diag2.Span.Start.Char)
			}
			return cmp.Compare(diag1.Span.Start.Line, diag2.Span.Start.Line)
		})
		all_notices[src_file_path] = file_notices
	}

	if !maps.EqualFunc(all_notices, allNotices, func(diags1 []*SrcFileNotice, diags2 []*SrcFileNotice) bool {
		return slices.EqualFunc(diags1, diags2, func(diag1 *SrcFileNotice, diag2 *SrcFileNotice) bool {
			return (diag1 == diag2) || (*diag1 == *diag2)
		})
	}) {
		allNotices = all_notices
		OnNoticesChanged(allNotices)
	}
}
