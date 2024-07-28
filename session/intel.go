package session

type IntelLookupKind int

const (
	IntelLookupKindDefs IntelLookupKind = iota
	IntelLookupKindDecls
	IntelLookupKindRefs
)

type IntelItemFormat int

const (
	IntelItemFormatPlainText IntelItemFormat = iota
	IntelItemFormatMarkdown
	IntelItemFormatAtmoSrcText
	IntelItemFormatPrimTypeTag
)

type IntelItemName int

const (
	IntelItemNameKind IntelItemName = iota
	IntelItemNameName
	IntelItemNameDescription
	IntelItemNameSignature
	IntelItemNamePrimType
	IntelItemNameExpansion
	IntelItemNameImport
	IntelItemNameUnused
	IntelItemNameTag // userland annotations like deprecated
	IntelItemNameStrBytesLen
	IntelItemNameStrUtf8RunesLen
	IntelItemNameNumHex
	IntelItemNameNumOct
	IntelItemNameNumDec
)

type Intel interface {
	Decls(pack *SrcPack, file *SrcFile, topLevelOnly bool, query string) (ret []IntelInfo)
	Lookup(kind IntelLookupKind, file *SrcFile, pos SrcFilePos, inFileOnly bool) (ret []SrcFileLocs)
	Completions(file *SrcFile, pos SrcFilePos) (ret []IntelInfo)
	Infos(file *SrcFile, pos SrcFilePos) *IntelInfo
	CanRename(file *SrcFile, pos SrcFilePos) *SrcFileSpan
}

type intel struct{}

type IntelItem struct {
	Name   IntelItemName
	Format IntelItemFormat
	Value  string
}

type IntelInfo struct {
	Infos     []IntelItem
	Sub       []*IntelInfo
	SpanIdent *SrcFileSpan
	SpanFull  *SrcFileSpan
}

func (intel) Decls(pack *SrcPack, file *SrcFile, topLevelOnly bool, query string) (ret []IntelInfo) {
	return
}

func (intel) Lookup(kind IntelLookupKind, file *SrcFile, pos SrcFilePos, inFileOnly bool) (ret []SrcFileLocs) {
	return
}

func (intel) Completions(file *SrcFile, pos SrcFilePos) (ret []IntelInfo) {
	return
}

func (intel) Infos(file *SrcFile, pos SrcFilePos) *IntelInfo {
	return nil
}

func (intel) CanRename(file *SrcFile, pos SrcFilePos) *SrcFileSpan {
	return nil
}
