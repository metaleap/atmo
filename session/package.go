package session

import (
	"os"
	"path/filepath"
	"strings"

	"atmo/util"
)

type SrcPkg struct {
	DirPath string
}

type SrcFile struct {
	FilePath string
	Pkg      *SrcPkg
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

func ensureSrcFile(srcFilePath string, curFullContent *string, canSkipFileRead bool) *SrcFile {
	util.Assert(IsSrcFilePath(srcFilePath), srcFilePath)

	if !util.FsIsFile(srcFilePath) {
		delete(state.srcFiles, srcFilePath)
		return nil
	}

	me := state.srcFiles[srcFilePath]
	if me == nil {
		me = &SrcFile{FilePath: srcFilePath}
		state.srcFiles[srcFilePath] = me
	}

	old_content, had_last_read_err := me.Content.Src, (me.Notices.LastReadErr != nil)
	if curFullContent != nil {
		me.Content.Src, me.Notices.LastReadErr = *curFullContent, nil
	} else if (!canSkipFileRead) || had_last_read_err {
		src_file_bytes, err := os.ReadFile(srcFilePath)
		if os.IsNotExist(err) {
			delete(state.srcFiles, srcFilePath)
			return nil
		} else {
			me.Content.Src, me.Notices.LastReadErr = string(src_file_bytes), errToNotice(err, NoticeCodeFileReadError, me.Span())
		}
	}

	if (me.Content.Src != old_content) || had_last_read_err || (me.Notices.LastReadErr != nil) {
		me.Content.Ast, me.Content.Est, me.Content.Toks, me.Notices.LexErrs = nil, nil, nil, nil
		if me.Notices.LastReadErr == nil {
			me.Content.Toks, me.Notices.LexErrs = tokenize(me.FilePath, me.Content.Src)
			if len(me.Notices.LexErrs) == 0 {
				me.parse()
				me.expand()
			}
		}
	}
	return me
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
	if (me.Content.Src != "") && (me.Content.Src[len(me.Content.Src)-1] != '\n') {
		ret.End.Line++
	}
	return
}
