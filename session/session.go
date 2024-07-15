package session

import (
	"os"
	"sync"

	"atmo/util"
)

var (
	allSrcFiles = map[string]*SrcFile{}
	allPackages = map[string]*SrcPkg{}
	sharedState sync.Mutex
	stateAccess = StateAccess{
		AllCurrentSrcFileNotices: func() map[string][]*SrcFileNotice { return allNotices },
		SrcFile: func(srcFilePath string, canSkipFileRead bool) *SrcFile {
			src_file := ensureSrcFile(srcFilePath, nil, canSkipFileRead)
			if src_file == nil { // file might be gone from diags by now
				defer refreshAndPublishNotices(srcFilePath)
			}
			return src_file
		},
	}
)

type StateAccess struct {
	AllCurrentSrcFileNotices func() map[string][]*SrcFileNotice
	SrcFile                  func(srcFilePath string, canSkipFileRead bool) *SrcFile
}

func OnSrcFileEvents(removed []string, canSkipFileRead bool, current ...string) {
	sharedState.Lock()
	defer sharedState.Unlock()

	for _, file_path := range removed {
		delete(allSrcFiles, file_path)
	}
	for _, file_path := range current {
		ensureSrcFile(file_path, nil, canSkipFileRead)
	}
	refreshAndPublishNotices(append(removed, current...)...)
}

func OnSrcFileEdit(srcFilePath string, curFullContent string) {
	sharedState.Lock()
	defer sharedState.Unlock()

	ensureSrcFile(srcFilePath, &curFullContent, true)
	refreshAndPublishNotices(srcFilePath)
}

func WithState(do func(sess *StateAccess)) {
	sharedState.Lock()
	defer sharedState.Unlock()
	do(&stateAccess)
}

func ensureSrcFile(srcFilePath string, curFullContent *string, canSkipFileRead bool) *SrcFile {
	util.Assert(IsSrcFilePath(srcFilePath), srcFilePath)

	if !util.FsIsFile(srcFilePath) {
		delete(allSrcFiles, srcFilePath)
		return nil
	}

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
		if os.IsNotExist(err) {
			delete(allSrcFiles, srcFilePath)
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
