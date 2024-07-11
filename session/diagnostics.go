package session

var (
	OnNoticesChanged = func(map[string][]*SrcFileNotice) {}
	OnDbgMsg         = func(string) {}
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
	NoticeCodeLexingError    SrcFileNoticeCode = "LexingError"
	NoticeCodeBracesMismatch SrcFileNoticeCode = "BracesMismatch"
	NoticeCodeBadLitSyntax   SrcFileNoticeCode = "BadLitSyntax"
	NoticeCodeMisplaced      SrcFileNoticeCode = "Misplaced"
	NoticeCodeExprExpected   SrcFileNoticeCode = "ExprExpected"
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
		ret = &SrcFileNotice{Kind: NoticeKindErr, Message: err.Error(), Code: code}
	}
	if span != nil {
		ret.Span = *span
	}
	return
}

func refreshAndPublishNotices(provokingFilePaths ...string) {
	pub := map[string][]*SrcFileNotice{}
	for _, src_file_path := range provokingFilePaths {
		pub[src_file_path] = nil
		if src_file := allSrcFiles[src_file_path]; src_file != nil {
			if src_file.Notices.LastReadErr != nil {
				pub[src_file_path] = append(pub[src_file_path], src_file.Notices.LastReadErr)
			}
			pub[src_file_path] = append(pub[src_file_path], src_file.Notices.LexErrs...)
			pub[src_file_path] = append(pub[src_file_path], src_file.Notices.ParseErrs...)
			for _, top_level_node := range src_file.Content.Ast {
				top_level_node.walk(nil, func(node *AstNode) {
					pub[src_file_path] = append(pub[src_file_path], node.errsParsing...)
				})
			}
		}
	}
	OnNoticesChanged(pub)
}
