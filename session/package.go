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

func IsSrcFilePath(filePath string) bool {
	return filepath.IsAbs(filePath) && filepath.Ext(filePath) == ".at" &&
		(!strings.Contains(filePath, string(filepath.Separator)+".")) && (!util.FsIsDir(filePath))
}

func removeSrcFiles(srcFilePaths ...string) {
	src_files := sl.As(srcFilePaths, func(it string) *SrcFile { return state.srcFiles[it] })
	del_pkgs := map[string]*SrcPkg{}
	for _, src_file := range src_files {
		if (src_file != nil) && (src_file.Pkg != nil) {
			src_file.Pkg.Files = sl.Where(src_file.Pkg.Files,
				func(it *SrcFile) bool { return it.FilePath != src_file.FilePath })
			if len(src_file.Pkg.Files) == 0 {
				del_pkgs[src_file.Pkg.DirPath] = src_file.Pkg
			}
		}
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

	old_content, had_last_read_err := me.Content.Src, (me.Notices.LastReadErr != nil)
	if curFullContent != nil {
		me.Content.Src, me.Notices.LastReadErr = *curFullContent, nil
	} else if (!canSkipFileRead) || had_last_read_err {
		src_file_bytes, err := os.ReadFile(srcFilePath)
		if os.IsNotExist(err) {
			removeSrcFiles(srcFilePath)
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

func (me *SrcFile) ensureSrcPkg() {
	if me.Pkg == nil {
		dir_path := filepath.Dir(me.FilePath)
		me.Pkg = state.srcPkgs[dir_path]
		if me.Pkg == nil {
			me.Pkg = &SrcPkg{DirPath: dir_path}
			state.srcPkgs[dir_path] = me.Pkg
		}
	}
	me.Pkg.Files = sl.With(me.Pkg.Files, me)
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
