package atmo_session

import (
	"os"
	"path/filepath"

	"atmo/util"
)

type SrcFilePos struct {
	Line int
	Char int
}

type SrcFileSpan struct {
	Start SrcFilePos
	End   SrcFilePos
}

type Session struct {
	allSrcFiles map[string]*SrcFile

	OnNoticesChanged func(map[string][]*SrcFileNotice)
}

func New() *Session {
	me := Session{
		allSrcFiles: map[string]*SrcFile{},
	}
	return &me
}

type SrcFile struct {
	FilePath    string
	Content     string
	LastReadErr error
}

func (me *Session) OnSrcFileEvents(removed []string, added []string, changed []string) {
	for _, file_path := range removed {
		delete(me.allSrcFiles, file_path)
	}
	for _, file_path := range added {
		me.ensureSrcFile(file_path, nil).Notices()
	}
	for _, file_path := range changed {
		me.ensureSrcFile(file_path, nil).Notices()
	}
}

func (me *Session) OnSrcFileEdit(srcFilePath string, curFullContent string) {
	me.ensureSrcFile(srcFilePath, &curFullContent)
}

func (me *Session) ensureSrcFile(srcFilePath string, curFullContent *string) *SrcFile {
	util.Assert(filepath.IsAbs(srcFilePath), srcFilePath)
	src_file := me.allSrcFiles[srcFilePath]
	if src_file == nil {
		src_file = &SrcFile{FilePath: srcFilePath}
		me.allSrcFiles[srcFilePath] = src_file
	}
	if curFullContent != nil {
		src_file.Content, src_file.LastReadErr = *curFullContent, nil
	} else {
		src_file_bytes, err := os.ReadFile(srcFilePath)
		src_file.Content, src_file.LastReadErr = string(src_file_bytes), err
	}
	return src_file
}
