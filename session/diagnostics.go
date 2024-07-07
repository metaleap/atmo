package session

type SrcFileNoticeKind int

const (
	_ SrcFileNoticeKind = iota
	NoticeKindErr
	NoticeKindWarn
	NoticeKindInfo
	NoticeKindHint
)

type SrcFileNoticeCode int

const (
	_ SrcFileNoticeCode = iota
	NoticeCodeOtherError
	NoticeCodeFileReadError
	NoticeCodeLexingUnknownLexeme
	NoticeCodeLexingOtherError
	NoticeCodeUnmatchedBrace
	NoticeCodeNoPrecedence
)

type SrcFileNotice struct {
	Kind    SrcFileNoticeKind
	Message string
	Span    SrcFileSpan
	Code    SrcFileNoticeCode
}

func (me *SrcFileNotice) Error() string  { return me.Message }
func (me *SrcFileNotice) String() string { return me.Message }

func errToNotice(err error, code SrcFileNoticeCode) (ret *SrcFileNotice) {
	if ret, _ = err.(*SrcFileNotice); (ret == nil) && (err != nil) {
		ret = &SrcFileNotice{Kind: NoticeKindErr, Message: err.Error(), Code: code}
	}
	return
}

func refreshAndPublishNotices(provokingFilePaths ...string) {
	pub := map[string][]*SrcFileNotice{}
	for _, src_file_path := range provokingFilePaths {
		pub[src_file_path] = nil
		if src_file := allSrcFiles[src_file_path]; src_file != nil {
			pub[src_file_path] = src_file.Notices.ParseErrs
			if src_file.Notices.LastReadErr != nil {
				pub[src_file_path] = append(pub[src_file_path], src_file.Notices.LastReadErr)
			}
		}
	}
	OnNoticesChanged(pub)
}
