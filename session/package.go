package session

import (
	"os"
	"path/filepath"
	"strings"

	"atmo/util"
	"atmo/util/sl"
)

type SrcPkg struct {
	DirPath string
	Files   []*SrcFile
}

type SrcFile struct {
	FilePath string
	pkg      *SrcPkg
	Content  struct {
		Src  string
		Toks Toks
		Ast  AstNodes
		Est  EstNodes
	} `json:"-"`
	notices struct {
		LastReadErr *SrcFileNotice
		LexErrs     []*SrcFileNotice
	}
}

func IsSrcFilePath(filePath string) bool {
	return filepath.IsAbs(filePath) && filepath.Ext(filePath) == ".at" &&
		(!strings.Contains(filePath, string(filepath.Separator)+".")) && (!util.FsIsDir(filePath))
}

func removeSrcFiles(srcFilePaths ...string) {
	src_files := sl.As(srcFilePaths, func(it string) *SrcFile { return state.srcFiles[it] })
	del_pkgs := map[string]*SrcPkg{}
	for _, src_file := range src_files {
		if (src_file != nil) && (src_file.pkg != nil) {
			src_file.pkg.Files = sl.Where(src_file.pkg.Files,
				func(it *SrcFile) bool { return (it != src_file) && (it.FilePath != src_file.FilePath) })
			if len(src_file.pkg.Files) == 0 {
				del_pkgs[src_file.pkg.DirPath] = src_file.pkg
			}
		}
	}
	for dir_path := range del_pkgs {
		delete(state.srcPkgs, dir_path)
	}
}

func ensureSrcFile(srcFilePath string, curFullContent *string, canSkipFileRead bool) *SrcFile {
	util.Assert(IsSrcFilePath(srcFilePath), srcFilePath)

	if !util.FsIsFile(srcFilePath) {
		removeSrcFiles(srcFilePath)
		return nil
	}

	me := state.srcFiles[srcFilePath]
	if me == nil {
		me = &SrcFile{FilePath: srcFilePath}
		state.srcFiles[srcFilePath] = me
		me.ensureSrcPkg()
	}

	old_content, had_last_read_err := me.Content.Src, (me.notices.LastReadErr != nil)
	if curFullContent != nil {
		me.Content.Src, me.notices.LastReadErr = *curFullContent, nil
	} else if (!canSkipFileRead) || had_last_read_err {
		src_file_bytes, err := os.ReadFile(srcFilePath)
		if os.IsNotExist(err) {
			removeSrcFiles(srcFilePath)
			return nil
		} else {
			me.Content.Src, me.notices.LastReadErr = string(src_file_bytes), errToNotice(err, NoticeCodeFileReadError, me.Span())
		}
	}

	if (me.Content.Src != old_content) || had_last_read_err || (me.notices.LastReadErr != nil) {
		me.Content.Ast, me.Content.Est, me.Content.Toks, me.notices.LexErrs = nil, nil, nil, nil
		if me.notices.LastReadErr == nil {
			me.Content.Toks, me.notices.LexErrs = tokenize(me.FilePath, me.Content.Src)
			if len(me.notices.LexErrs) == 0 {
				me.parse()
				me.expand()
			}
		}
	}
	return me
}

func (me *SrcFile) ensureSrcPkg() {
	if me.pkg == nil {
		dir_path := filepath.Dir(me.FilePath)
		me.pkg = state.srcPkgs[dir_path]
		if me.pkg == nil {
			me.pkg = &SrcPkg{DirPath: dir_path}
			state.srcPkgs[dir_path] = me.pkg
		}
	}
	me.pkg.Files = sl.With(me.pkg.Files, me)
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
