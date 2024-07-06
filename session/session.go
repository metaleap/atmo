package session

import (
	"os"
	"path/filepath"
	"strings"

	"atmo/util"
)

var (
	allSrcFiles map[string]*SrcFile

	OnNoticesChanged func(map[string][]*SrcFileNotice)
)

type SrcFilePos struct {
	Line int
	Char int
}

type SrcFileSpan struct {
	Start SrcFilePos
	End   SrcFilePos
}

func init() {
	allSrcFiles = map[string]*SrcFile{}
}

type SrcFile struct {
	FilePath string
	Content  struct {
		Src  string
		Toks Tokens
		Ast  *AstFile
	}
	Notices struct {
		LastReadErr *SrcFileNotice
		ParseErrs   []*SrcFileNotice
	}
}

func OnSrcFileEvents(removed []string, canSkipFileRead bool, current ...string) {
	for _, file_path := range removed {
		delete(allSrcFiles, file_path)
	}
	for _, file_path := range current {
		ensureSrcFile(file_path, nil, canSkipFileRead)
	}
	refreshAndPublishNotices(append(removed, current...)...)
}

func OnSrcFileEdit(srcFilePath string, curFullContent string) {
	ensureSrcFile(srcFilePath, &curFullContent, true)
	refreshAndPublishNotices(srcFilePath)
}

func ensureSrcFile(srcFilePath string, curFullContent *string, canSkipFileRead bool) *SrcFile {
	util.Assert(filepath.IsAbs(srcFilePath), srcFilePath)
	src_file := allSrcFiles[srcFilePath]
	if src_file == nil {
		src_file = &SrcFile{FilePath: srcFilePath}
		allSrcFiles[srcFilePath] = src_file
	}
	old_content, had_read_err := src_file.Content.Src, (src_file.Notices.LastReadErr != nil)
	if curFullContent != nil {
		src_file.Content.Src, src_file.Notices.LastReadErr = *curFullContent, nil
	} else if (!canSkipFileRead) || had_read_err || (old_content == "") {
		src_file_bytes, err := os.ReadFile(srcFilePath)
		src_file.Content.Src, src_file.Notices.LastReadErr = string(src_file_bytes), errToNotice(err, NoticeCodeFileReadError)
	}
	if (src_file.Content.Src != old_content) || had_read_err || (src_file.Notices.LastReadErr != nil) {
		src_file.Content.Ast, src_file.Content.Toks, src_file.Notices.ParseErrs = nil, nil, nil
		if src_file.Notices.LastReadErr == nil {
			src_file.Content.Toks = tokenize(src_file.Content.Src)
			src_file.Content.Ast, src_file.Notices.ParseErrs = parse(src_file.Content.Toks, src_file.Content.Src, srcFilePath)
		}
	}
	return src_file
}

func IsSrcFilePath(filePath string) bool {
	return filepath.Ext(filePath) == ".at" && (!strings.Contains(filePath, string(filepath.Separator)+".")) && (!util.FsIsDir(filePath))
}
