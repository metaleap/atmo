package session

import (
	"path/filepath"
	"strings"

	"atmo/util"
)

type SrcPkg struct{}

type SrcFile struct {
	FilePath string
	Content  struct {
		Src  string
		Toks Toks
		Ast  AstNodes
		Est  EstNodes
	}
	Notices struct {
		LastReadErr *SrcFileNotice
		LexErrs     []*SrcFileNotice
	}
}

func IsSrcFilePath(filePath string) bool {
	return filepath.IsAbs(filePath) && filepath.Ext(filePath) == ".at" &&
		(!strings.Contains(filePath, string(filepath.Separator)+".")) && (!util.FsIsDir(filePath))
}

func (me *SrcFile) Span() (ret SrcFileSpan) {
	ret.Start, ret.End = SrcFilePos{Line: 1, Char: 1}, SrcFilePos{Line: 1, Char: 1}
	for i := 0; i < len(me.Content.Src); i++ {
		if me.Content.Src[i] == '\n' {
			ret.End.Line++
		}
	}
	if me.Content.Src[len(me.Content.Src)-1] != '\n' {
		ret.End.Line++
	}
	return
}
