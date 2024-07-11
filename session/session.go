package session

import (
	"os"
	"path/filepath"
	"strings"

	"atmo/util"
)

var (
	allSrcFiles map[string]*SrcFile
)

func init() {
	allSrcFiles = map[string]*SrcFile{}
}

type SrcFile struct {
	FilePath string
	Content  struct {
		Src  string
		Toks Toks
		Ast  Nodes
	}
	Notices struct {
		LastReadErr *SrcFileNotice
		LexErrs     []*SrcFileNotice
		// ParseErrs has only those parsing errors that can't be in a `Node.Errs.Parsing`
		ParseErrs []*SrcFileNotice
	}
}

func OnSrcFileEvents(removed []string, canSkipFileRead bool, current ...string) {
	for _, file_path := range removed {
		delete(allSrcFiles, file_path)
	}
	for _, file_path := range current {
		EnsureSrcFile(file_path, nil, canSkipFileRead)
	}
	refreshAndPublishNotices(append(removed, current...)...)
}

func OnSrcFileEdit(srcFilePath string, curFullContent string) {
	EnsureSrcFile(srcFilePath, &curFullContent, true)
	refreshAndPublishNotices(srcFilePath)
}

func EnsureSrcFile(srcFilePath string, curFullContent *string, canSkipFileRead bool) *SrcFile {
	util.Assert(IsSrcFilePath(srcFilePath), srcFilePath)
	me := allSrcFiles[srcFilePath]
	if me == nil {
		me = &SrcFile{FilePath: srcFilePath}
		allSrcFiles[srcFilePath] = me
	}
	old_content, had_last_read_err := me.Content.Src, (me.Notices.LastReadErr != nil)
	if curFullContent != nil {
		me.Content.Src, me.Notices.LastReadErr = *curFullContent, nil
	} else if (!canSkipFileRead) || had_last_read_err || (old_content == "") {
		src_file_bytes, err := os.ReadFile(srcFilePath)
		me.Content.Src, me.Notices.LastReadErr = string(src_file_bytes), errToNotice(err, NoticeCodeFileReadError, nil)
	}
	if (me.Content.Src != old_content) || had_last_read_err || (me.Notices.LastReadErr != nil) {
		ast_prev := me.Content.Ast
		me.Content.Ast, me.Content.Toks, me.Notices.LexErrs, me.Notices.ParseErrs =
			nil, nil, nil, nil
		if me.Notices.LastReadErr == nil {
			me.Content.Toks, me.Notices.LexErrs = me.tokenize()
			if len(me.Notices.LexErrs) == 0 {
				me.parse(ast_prev)
			}
		}
	}
	return me
}

func IsSrcFilePath(filePath string) bool {
	return filepath.IsAbs(filePath) && filepath.Ext(filePath) == ".at" &&
		(!strings.Contains(filePath, string(filepath.Separator)+".")) && (!util.FsIsDir(filePath))
}
