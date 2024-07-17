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
		LexErrs     SrcFileNotices
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

func ensureSrcFiles(curFullContent *string, canSkipFileRead bool, srcFilePaths ...string) (encounteredDiagsRelevantChanges []string) {
	if len(srcFilePaths) == 0 {
		return
	}
	pkgs_to_refresh := map[*SrcPkg]bool{}
	util.Assert((curFullContent == nil) || (len(srcFilePaths) == 1), len(srcFilePaths))

	for _, src_file_path := range srcFilePaths {
		flag_for_diags_refr := func() { encounteredDiagsRelevantChanges = sl.With(encounteredDiagsRelevantChanges, src_file_path) }
		util.Assert(IsSrcFilePath(src_file_path), src_file_path)

		if !util.FsIsFile(src_file_path) {
			removeSrcFiles(src_file_path)
			flag_for_diags_refr()
			continue
		}

		src_file := state.srcFiles[src_file_path]
		if src_file == nil {
			flag_for_diags_refr()
			src_file = &SrcFile{FilePath: src_file_path}
			state.srcFiles[src_file_path] = src_file
			// ensure SrcPkg
			pkg_dir_path := filepath.Dir(src_file.FilePath)
			src_file.pkg = state.srcPkgs[pkg_dir_path]
			if src_file.pkg == nil {
				src_file.pkg = &SrcPkg{DirPath: pkg_dir_path}
				state.srcPkgs[pkg_dir_path] = src_file.pkg
			}
			src_file.pkg.Files = sl.With(src_file.pkg.Files, src_file)
		}

		old_content, had_last_read_err := src_file.Content.Src, (src_file.notices.LastReadErr != nil)
		if curFullContent != nil {
			src_file.Content.Src, src_file.notices.LastReadErr = *curFullContent, nil
		} else if (!canSkipFileRead) || had_last_read_err {
			src_file_bytes, err := os.ReadFile(src_file_path)
			if os.IsNotExist(err) {
				removeSrcFiles(src_file_path)
				flag_for_diags_refr()
				continue
			} else {
				src_file.Content.Src, src_file.notices.LastReadErr = string(src_file_bytes), errToNotice(err, NoticeCodeFileReadError, src_file.Span())
			}
		}

		if (src_file.Content.Src != old_content) || had_last_read_err || (src_file.notices.LastReadErr != nil) {
			old_ast := src_file.Content.Ast
			src_file.Content.Ast, src_file.Content.Toks, src_file.notices.LexErrs = nil, nil, nil
			if src_file.notices.LastReadErr != nil {
				flag_for_diags_refr()
			} else {
				src_file.Content.Toks, src_file.notices.LexErrs = tokenize(src_file.FilePath, src_file.Content.Src)
				if len(src_file.notices.LexErrs) > 0 {
					flag_for_diags_refr()
				} else {
					new_ast := src_file.parse()
					if new_ast.hasKind(AstNodeKindErr) {
						flag_for_diags_refr()
					}
					new_same_as_old := make(map[*AstNode]bool, len(old_ast)) // avoids double-counting
					if len(old_ast) == len(new_ast) {
						for _, new_node := range new_ast {
							for _, old_node := range old_ast {
								if old_node.equals(new_node, true) {
									new_same_as_old[new_node] = true
									break
								}
							}
						}
					}
					have_changes := (len(new_same_as_old) != len(old_ast)) || (len(new_same_as_old) != len(new_ast))

					src_file.Content.Ast = new_ast
					if have_changes { // false if changes were in comments, whitespace (other than top-level indentation), or mere re-ordering of top-level nodes
						pkgs_to_refresh[src_file.pkg] = true
					}
				}
			}
		}
	}

	for src_pkg := range pkgs_to_refresh {
		if src_pkg.refreshEst() {
			encounteredDiagsRelevantChanges = sl.With(encounteredDiagsRelevantChanges, src_pkg.srcFilePaths()...)
		}
	}
	return
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
