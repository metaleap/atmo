package session

import "atmo/util/sl"

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

type IntelItemKind int

const (
	IntelItemKindName IntelItemKind = iota
	IntelItemKindDescription
	IntelItemKindKind // func, lit, var etc
	IntelItemKindSignature
	IntelItemKindPrimType
	IntelItemKindExpansion
	IntelItemKindImport
	IntelItemKindUnused
	IntelItemKindTag // userland annotations like deprecated
	IntelItemKindStrBytesLen
	IntelItemKindStrUtf8RunesLen
	IntelItemKindNumHex
	IntelItemKindNumOct
	IntelItemKindNumDec
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
	Kind   IntelItemKind
	Format IntelItemFormat
	Value  string
}
type IntelItems sl.Of[IntelItem]

type IntelInfo struct {
	Infos     IntelItems
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

func (me IntelItems) First(kind IntelItemKind) *IntelItem {
	for i := range me {
		if item := &me[i]; item.Kind == kind {
			return item
		}
	}
	return nil
}

func (me IntelItems) Where(kind IntelItemKind) IntelItems {
	return sl.Where(me, func(item IntelItem) bool { return item.Kind == kind })
}

func (me IntelItems) Name() *IntelItem {
	return me.First(IntelItemKindName)
}
