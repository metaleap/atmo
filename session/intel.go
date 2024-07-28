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
	/*sl.As([]lsp.SymbolKind{lsp.SymbolKindArray, lsp.SymbolKindBoolean, lsp.SymbolKindClass, lsp.SymbolKindConstant, lsp.SymbolKindConstructor, lsp.SymbolKindEnum, lsp.SymbolKindEnumMember, lsp.SymbolKindEvent, lsp.SymbolKindField, lsp.SymbolKindFile, lsp.SymbolKindFunction, lsp.SymbolKindInterface, lsp.SymbolKindKey, lsp.SymbolKindMethod, lsp.SymbolKindModule, lsp.SymbolKindNamespace, lsp.SymbolKindNull, lsp.SymbolKindNumber, lsp.SymbolKindObject, lsp.SymbolKindOperator, lsp.SymbolKindPackage, lsp.SymbolKindProperty, lsp.SymbolKindString, lsp.SymbolKindStruct, lsp.SymbolKindTypeParameter, lsp.SymbolKindVariable},
	func(it lsp.SymbolKind) lsp.DocumentSymbol {
		return lsp.DocumentSymbol{
			Name:           it.String(),
			Detail:         fmt.Sprintf("**TODO:** documentSymbols for `%v`", src_file_path),
			Kind:           it,
			Range:          lsp.Range{Start: lsp.Position{Line: 2, Character: 1}, End: lsp.Position{Line: 2, Character: 8}},
			SelectionRange: lsp.Range{Start: lsp.Position{Line: 2, Character: 3}, End: lsp.Position{Line: 2, Character: 6}},
		}
	}), nil*/
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
