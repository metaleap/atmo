package session

import (
	"os"
	"path/filepath"
	"strings"

	"atmo/util"
	"atmo/util/sl"
)

type SrcPack struct {
	DirPath string
	Files   []*SrcFile
	Est     EstNodes
}

type SrcFile struct {
	FilePath string
	pack     *SrcPack
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

func packsFsRefresh() {
	var gone_files []string
	var gone_packs []string
	for src_file_path := range state.srcFiles {
		if !util.FsIsFile(src_file_path) {
			gone_files = append(gone_files, src_file_path)
		}
	}
	for pack_dir_path, src_pack := range state.srcPacks {
		if !util.FsIsDir(pack_dir_path) {
			gone_files = append(gone_files, src_pack.srcFilePaths()...)
			gone_packs = append(gone_packs, pack_dir_path)
		}
	}
	removeSrcFiles(gone_files...)
	for _, pack_dir_path := range gone_packs {
		delete(state.srcPacks, pack_dir_path)
	}
}

func removeSrcFiles(srcFilePaths ...string) {
	if len(srcFilePaths) == 0 {
		return
	}
	packs_to_drop, packs_encountered := map[string]*SrcPack{}, map[string]*SrcPack{}
	for _, src_file_path := range srcFilePaths {
		src_file := state.srcFiles[src_file_path]
		if (src_file != nil) && (src_file.pack != nil) {
			packs_encountered[src_file.pack.DirPath] = src_file.pack
			src_file.pack.Files = sl.Where(src_file.pack.Files,
				func(it *SrcFile) bool { return (it != src_file) && (it.FilePath != src_file.FilePath) })
			if len(src_file.pack.Files) == 0 {
				packs_to_drop[src_file.pack.DirPath] = src_file.pack
			}
		}
		delete(state.srcFiles, src_file_path)
	}

	var pack_file_paths []string
	for pack_dir_path := range packs_to_drop {
		delete(state.srcPacks, pack_dir_path)
	}
	for _, src_pack := range packs_encountered {
		pack_file_paths = append(pack_file_paths, src_pack.srcFilePaths()...)
		src_pack.refreshEst()
	}
	refreshAndPublishNotices(append(pack_file_paths, srcFilePaths...)...)
}

func ensureSrcFiles(curFullContent *string, canSkipFileRead bool, srcFilePaths ...string) (encounteredDiagsRelevantChanges []string) {
	if len(srcFilePaths) == 0 {
		return
	}
	packs_to_refresh := map[*SrcPack]bool{}
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
			// ensure SrcPack
			pack_dir_path := filepath.Dir(src_file.FilePath)
			src_file.pack = state.srcPacks[pack_dir_path]
			if src_file.pack == nil {
				src_file.pack = &SrcPack{DirPath: pack_dir_path}
				state.srcPacks[pack_dir_path] = src_file.pack
			}
			src_file.pack.Files = sl.With(src_file.pack.Files, src_file)
			canSkipFileRead = false
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
						old_ast_sans_comments := old_ast.withoutComments()
						for _, new_node := range new_ast.withoutComments() {
							for _, old_node := range old_ast_sans_comments {
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
						packs_to_refresh[src_file.pack] = true
					}
				}
			}
		}
	}

	for src_pack := range packs_to_refresh {
		if src_pack.refreshEst() {
			encounteredDiagsRelevantChanges = sl.With(encounteredDiagsRelevantChanges, src_pack.srcFilePaths()...)
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

func (me *SrcPack) srcFilePaths() []string {
	return sl.As(me.Files, func(it *SrcFile) string { return it.FilePath })
}
