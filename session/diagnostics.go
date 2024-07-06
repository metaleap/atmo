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
	NoticeCodeFileReadError
	NoticeCodeUnmatchedBrace
	NoticeCodeNoPrecedence
	NoticeCodeMultipleAstNodes
)

type SrcFileNotice struct {
	Kind    SrcFileNoticeKind
	Message string
	Span    SrcFileSpan
	Code    SrcFileNoticeCode
}

func (me *SrcFileNotice) Error() string  { return me.Message }
func (me *SrcFileNotice) String() string { return me.Message }

func (me *SrcFile) Notices() []*SrcFileNotice {
	if me.LastReadErr != nil {
		return []*SrcFileNotice{{Kind: NoticeKindErr, Message: me.LastReadErr.Error(), Code: NoticeCodeFileReadError}}
	}
	return nil
}
