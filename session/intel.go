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

func IntelDecls(pack *SrcPack, file *SrcFile, topLevelOnly bool, query string) (ret []IntelInfo) {
	return
}

func IntelLookup(kind IntelLookupKind, file *SrcFile, pos SrcFilePos, inFileOnly bool) (ret []SrcFileLocs) {
	return
}

func IntelCompletions(file *SrcFile, pos SrcFilePos) (ret []IntelInfo) {
	return
}

func IntelInfos(file *SrcFile, pos SrcFilePos) *IntelInfo {
	return nil
}

func IntelCanRename(file *SrcFile, pos SrcFilePos) *SrcFileSpan {
	return nil
}

func IntelSignatures(file *SrcFile, pos SrcFilePos) (ret []IntelItem) {
	return
}
