package session

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"atmo/util"
)

var (
	allSrcFiles      map[string]*SrcFile
	allSrcFilesMutex sync.Mutex
)

func init() {
	allSrcFiles = map[string]*SrcFile{}
}

type SrcFile struct {
	FilePath string
	Content  struct {
		Src  string
		Toks Toks
		Ast  AstNodes
	}
	Notices struct {
		LastReadErr *SrcFileNotice
		LexErrs     []*SrcFileNotice
	}
}

func OnSrcFileEvents(removed []string, canSkipFileRead bool, current ...string) {
	allSrcFilesMutex.Lock()
	defer allSrcFilesMutex.Unlock()

	for _, file_path := range removed {
		delete(allSrcFiles, file_path)
	}
	for _, file_path := range current {
		ensureSrcFile(file_path, nil, canSkipFileRead)
	}
	refreshAndPublishNotices(append(removed, current...)...)
}

func OnSrcFileEdit(srcFilePath string, curFullContent string) {
	allSrcFilesMutex.Lock()
	defer allSrcFilesMutex.Unlock()

	ensureSrcFile(srcFilePath, &curFullContent, true)
	refreshAndPublishNotices(srcFilePath)
}

func WithSrcFileDo(srcFilePath string, canSkipFileRead bool, do func(srcFile *SrcFile)) {
	allSrcFilesMutex.Lock()
	defer allSrcFilesMutex.Unlock()

	if src_file := ensureSrcFile(srcFilePath, nil, canSkipFileRead); src_file != nil {
		do(src_file)
	}
}

func ensureSrcFile(srcFilePath string, curFullContent *string, canSkipFileRead bool) *SrcFile {
	util.Assert(IsSrcFilePath(srcFilePath), srcFilePath)

	me := allSrcFiles[srcFilePath]
	if me == nil {
		me = &SrcFile{FilePath: srcFilePath}
		allSrcFiles[srcFilePath] = me
	}
	old_content, had_last_read_err := me.Content.Src, (me.Notices.LastReadErr != nil)
	if curFullContent != nil {
		me.Content.Src, me.Notices.LastReadErr = *curFullContent, nil
	} else if (!canSkipFileRead) || had_last_read_err {
		src_file_bytes, err := os.ReadFile(srcFilePath)
		me.Content.Src, me.Notices.LastReadErr = string(src_file_bytes), errToNotice(err, NoticeCodeFileReadError, nil)
	}
	if (me.Content.Src == "") || (me.Content.Src != old_content) || had_last_read_err || (me.Notices.LastReadErr != nil) {
		me.Content.Ast, me.Content.Toks, me.Notices.LexErrs = nil, nil, nil
		if me.Notices.LastReadErr == nil {
			me.Content.Toks, me.Notices.LexErrs = tokenize(me.FilePath, me.Content.Src)
			if len(me.Notices.LexErrs) == 0 {
				me.parse()
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
	if me.Content.Src[len(me.Content.Src)-1] != '\n' {
		ret.End.Line++
	}
	return
}
