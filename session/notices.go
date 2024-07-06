package atmo_session

type SrcFileNoticeKind int

const (
	_ SrcFileNoticeKind = iota
	SrcFileNoticeKindErr
	SrcFileNoticeKindWarn
	SrcFileNoticeKindInfo
	SrcFileNoticeKindHint
)

type SrcFileNoticeCode int

const (
	_ SrcFileNoticeCode = iota
	SrcFileNoticeCodeFileReadError
)

type SrcFileNotice struct {
	Kind    SrcFileNoticeKind
	Message string
	Span    SrcFileSpan
	Code    SrcFileNoticeCode
}

func (me *SrcFile) Notices() []*SrcFileNotice {
	if me.LastReadErr != nil {
		return []*SrcFileNotice{{Kind: SrcFileNoticeKindErr, Message: me.LastReadErr.Error(), Code: SrcFileNoticeCodeFileReadError}}
	}
	return nil
}
