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
	Est     EstNodes
}

type SrcFile struct {
	FilePath string
	pkg      *SrcPkg
	Content  struct {
		Src  string
		Toks Toks
		Ast  AstNodes
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

func pkgsFsRefresh() {
	var gone_files []string
	var gone_pkgs []string
	for src_file_path := range state.srcFiles {
		if !util.FsIsFile(src_file_path) {
			gone_files = append(gone_files, src_file_path)
		}
	}
	for pkg_dir_path, src_pkg := range state.srcPkgs {
		if !util.FsIsDir(pkg_dir_path) {
			gone_files = append(gone_files, src_pkg.srcFilePaths()...)
			gone_pkgs = append(gone_pkgs, pkg_dir_path)
		}
	}
	removeSrcFiles(gone_files...)
	for _, pkg_dir_path := range gone_pkgs {
		delete(state.srcPkgs, pkg_dir_path)
	}
}

func removeSrcFiles(srcFilePaths ...string) {
	if len(srcFilePaths) == 0 {
		return
	}
	pkgs_to_delete, pkgs_encountered := map[string]*SrcPkg{}, map[string]*SrcPkg{}
	for _, src_file_path := range srcFilePaths {
		src_file := state.srcFiles[src_file_path]
		if (src_file != nil) && (src_file.pkg != nil) {
			pkgs_encountered[src_file.pkg.DirPath] = src_file.pkg
			src_file.pkg.Files = sl.Where(src_file.pkg.Files,
				func(it *SrcFile) bool { return (it != src_file) && (it.FilePath != src_file.FilePath) })
			if len(src_file.pkg.Files) == 0 {
				pkgs_to_delete[src_file.pkg.DirPath] = src_file.pkg
			}
		}
		delete(state.srcFiles, src_file_path)
	}

	var pkg_file_paths []string
	for pkg_dir_path := range pkgs_to_delete {
		delete(state.srcPkgs, pkg_dir_path)
	}
	for _, src_pkg := range pkgs_encountered {
		pkg_file_paths = append(pkg_file_paths, src_pkg.srcFilePaths()...)
		src_pkg.refreshEst()
	}
	refreshAndPublishNotices(append(pkg_file_paths, srcFilePaths...)...)
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
	}
	{ // ensure SrcPkg
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
		old_ast := me.Content.Ast
		me.Content.Ast, me.Content.Toks, me.notices.LexErrs = nil, nil, nil
		if me.notices.LastReadErr == nil {
			me.Content.Toks, me.notices.LexErrs = tokenize(me.FilePath, me.Content.Src)
			if len(me.notices.LexErrs) == 0 {
				new_ast := me.parse()
				var num_same_nodes int
				if len(old_ast) == len(new_ast) {
					for _, old_node := range old_ast {
						for _, new_node := range new_ast {
							if old_node.equals(new_node, true) {
								num_same_nodes++
							}
						}
					}
				}
				have_any_changes := (num_same_nodes != len(old_ast)) || (num_same_nodes != len(new_ast))
				me.Content.Ast = new_ast
				if have_any_changes { // false if changes were in comments, whitespace (other than top-level indentation), or mere re-ordering of top-level nodes
					me.pkg.refreshEst()
				}
			}
		}
	}
	return me
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

func (me *SrcPkg) srcFilePaths() []string {
	return sl.As(me.Files, func(it *SrcFile) string { return it.FilePath })
}
