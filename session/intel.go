package session

import (
	"atmo/util"
	"atmo/util/sl"
	"atmo/util/str"
)

type IntelLookupKind int

const (
	IntelLookupKindDefs IntelLookupKind = iota
	IntelLookupKindDecls
	IntelLookupKindRefs
	IntelLookupKindTypes
	IntelLookupKindImpls
)

type IntelItemKind int

const (
	IntelItemKindName IntelItemKind = iota
	IntelItemKindDescription
	IntelItemKindKind // func, lit, var etc
	IntelItemKindSrcPackDirPath
	IntelItemKindSrcFilePath
	IntelItemKindPrimType
	IntelItemKindExpansion
	IntelItemKindImport
	IntelItemKindTag // userland annotations like deprecated
	IntelItemKindStrBytesLen
	IntelItemKindStrUtf8RunesLen
	IntelItemKindNumHex
	IntelItemKindNumOct
	IntelItemKindNumDec
)

type IntelDeclKind string

const (
	IntelDeclKindFunc IntelDeclKind = "func"
	IntelDeclKindVar  IntelDeclKind = "var"
)

type Intel interface {
	Decls(pack *SrcPack, file *SrcFile, topLevelOnly bool, query string) (ret []*IntelInfo)
	Lookup(kind IntelLookupKind, file *SrcFile, pos SrcFilePos, inFileOnly bool) (ret []*SrcFileLocs)
	Completions(file *SrcFile, pos SrcFilePos) (ret []*IntelInfo)
	Info(file *SrcFile, pos SrcFilePos) *IntelInfo
	CanRename(file *SrcFile, pos SrcFilePos) *SrcFileSpan
}

type intel struct{}

type IntelItem struct {
	Kind     IntelItemKind
	Value    string
	CodeLang string
}
type IntelItems sl.Of[IntelItem]

type IntelInfo struct {
	Items     IntelItems
	Sub       []*IntelInfo
	SpanIdent *SrcFileSpan
	SpanFull  *SrcFileSpan
}

// temporary fake impl
func (me intel) Decls(pack *SrcPack, file *SrcFile, topLevelOnly bool, query string) (ret []*IntelInfo) {
	if file == nil { // for temporary fake impl
		file = me.dummyFile()
	}
	if (pack == nil) && (file != nil) {
		pack = file.pack
	}
	ret = append(ret, &IntelInfo{
		SpanIdent: &SrcFileSpan{Start: SrcFilePos{Line: 1, Char: 4}, End: SrcFilePos{Line: 1, Char: 11}},
		SpanFull:  &SrcFileSpan{Start: SrcFilePos{Line: 1, Char: 1}, End: SrcFilePos{Line: 4, Char: 123}},
		Items: IntelItems{
			IntelItem{Kind: IntelItemKindName, Value: "FakeSym1"},
			IntelItem{Kind: IntelItemKindDescription, Value: "Fake symbol 1"},
			IntelItem{Kind: IntelItemKindKind, Value: string(IntelDeclKindVar)},
			IntelItem{Kind: IntelItemKindSrcFilePath, Value: file.FilePath},
			IntelItem{Kind: IntelItemKindSrcPackDirPath, Value: pack.DirPath},
		},
	})
	if !topLevelOnly {
		ret[0].Sub = []*IntelInfo{{
			SpanIdent: &SrcFileSpan{Start: SrcFilePos{Line: 3, Char: 4}, End: SrcFilePos{Line: 3, Char: 11}},
			SpanFull:  &SrcFileSpan{Start: SrcFilePos{Line: 3, Char: 4}, End: SrcFilePos{Line: 3, Char: 123}},
			Items: IntelItems{
				IntelItem{Kind: IntelItemKindName, Value: "FakeSym2"},
				IntelItem{Kind: IntelItemKindDescription, Value: "Fake symbol 2"},
				IntelItem{Kind: IntelItemKindKind, Value: string(IntelDeclKindFunc)},
				IntelItem{Kind: IntelItemKindSrcFilePath, Value: file.FilePath},
				IntelItem{Kind: IntelItemKindSrcPackDirPath, Value: pack.DirPath},
			},
		}}
	}
	return
}

// temporary fake impl
func (me intel) Lookup(kind IntelLookupKind, file *SrcFile, pos SrcFilePos, inFileOnly bool) (ret []*SrcFileLocs) {
	return me.dummyLocs()
}

// temporary fake impl
func (intel) Completions(file *SrcFile, pos SrcFilePos) (ret []*IntelInfo) {
	return
}

func (intel) Info(file *SrcFile, pos SrcFilePos) (ret *IntelInfo) {
	node := file.pack.Trees.Sem.TopLevel.At(file, &pos)
	if node != nil {
		ret = &IntelInfo{SpanFull: node.From.SrcSpan}
		for it := node; it != nil; it = it.Parent {
			var str_val string
			if it.HasFact(SemFactPreComputed, nil, false, false) {
				str_val = it.String()
			}
			var str_facts string
			for k := range it.Facts {
				str_facts += k.String() + ", "
			}
			if str_facts = str.TrimSuff(str_facts, ", "); str_facts != "" {
				str_facts = "Facts: "
			}
			var str_src string
			if (it.From != nil) && (it.From.SrcNode != nil) {
				str_src = it.From.SrcNode.Src
			}
			ret.Items = append(ret.Items, IntelItem{
				Kind: IntelItemKindDescription, Value: str_src, CodeLang: "atmo",
			}, IntelItem{
				Kind: IntelItemKindDescription, Value: SemTypeToString(it.Type) + util.If(str_val == "", "", " — (`"+str_val+"`)"),
			}, IntelItem{
				Kind: IntelItemKindDescription, Value: str_facts,
			})
		}
	}
	return
}

// temporary fake impl
func (intel) CanRename(file *SrcFile, pos SrcFilePos) *SrcFileSpan {
	return &SrcFileSpan{Start: pos, End: SrcFilePos{Line: pos.Line, Char: 4 + pos.Char}}
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

func (intel) dummyFile() *SrcFile {
	for _, src_file := range state.srcFiles {
		if !src_file.IsInterpFauxFile() {
			return src_file
		}
	}
	return nil
}

func (me intel) dummyLocs() []*SrcFileLocs {
	file := me.dummyFile()
	return []*SrcFileLocs{
		{File: file, Spans: []*SrcFileSpan{{Start: SrcFilePos{Line: 2, Char: 1}, End: SrcFilePos{Line: 2, Char: 8}}}},
		{File: file, Spans: []*SrcFileSpan{{Start: SrcFilePos{Line: 4, Char: 1}, End: SrcFilePos{Line: 4, Char: 8}}}},
	}
}
