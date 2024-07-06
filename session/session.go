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
	FilePath    string
	Content     string
	LastReadErr error
}

func OnSrcFileEvents(removed []string, added []string, changed []string) {
	for _, file_path := range removed {
		delete(allSrcFiles, file_path)
	}
	for _, file_path := range added {
		ensureSrcFile(file_path, nil)
	}
	for _, file_path := range changed {
		ensureSrcFile(file_path, nil)
	}
}

func OnSrcFileEdit(srcFilePath string, curFullContent string) {
	ensureSrcFile(srcFilePath, &curFullContent)
}

func ensureSrcFile(srcFilePath string, curFullContent *string) *SrcFile {
	util.Assert(filepath.IsAbs(srcFilePath), srcFilePath)
	src_file := allSrcFiles[srcFilePath]
	if src_file == nil {
		src_file = &SrcFile{FilePath: srcFilePath}
		allSrcFiles[srcFilePath] = src_file
	}
	if curFullContent != nil {
		src_file.Content, src_file.LastReadErr = *curFullContent, nil
	} else {
		src_file_bytes, err := os.ReadFile(srcFilePath)
		src_file.Content, src_file.LastReadErr = string(src_file_bytes), err
	}
	return src_file
}

func IsSrcFilePath(filePath string) bool {
	return filepath.Ext(filePath) == ".at" && (!strings.Contains(filePath, string(filepath.Separator)+".")) && (!util.FsIsDir(filePath))
}
