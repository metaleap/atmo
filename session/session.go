package session

import (
	"sync"
)

var (
	state struct {
		stateAccess
		srcFiles map[string]*SrcFile
		srcPkgs  map[string]*SrcPkg
	}
)

type StateAccess interface {
	OnSrcFileEdit(srcFilePath string, curFullContent string)
	OnSrcFileEvents(removed []string, canSkipFileRead bool, current ...string)

	AllCurrentSrcFileNotices() map[string][]*SrcFileNotice
	SrcFile(srcFilePath string, canSkipFileRead bool) *SrcFile
}

func init() {
	state.srcFiles, state.srcPkgs = map[string]*SrcFile{}, map[string]*SrcPkg{}
}

func LockedDo(do func(sess StateAccess)) {
	state.Lock()
	defer state.Unlock()
	do(&state.stateAccess)
}

type stateAccess struct{ sync.Mutex }

func (*stateAccess) OnSrcFileEdit(srcFilePath string, curFullContent string) {
	ensureSrcFile(srcFilePath, &curFullContent, true)
	refreshAndPublishNotices(srcFilePath)
}

func (*stateAccess) OnSrcFileEvents(removed []string, canSkipFileRead bool, current ...string) {
	for _, file_path := range removed {
		delete(state.srcFiles, file_path)
	}
	for _, file_path := range current {
		ensureSrcFile(file_path, nil, canSkipFileRead)
	}
	refreshAndPublishNotices(append(removed, current...)...)
}

func (*stateAccess) AllCurrentSrcFileNotices() map[string][]*SrcFileNotice {
	return allNotices
}

func (*stateAccess) SrcFile(srcFilePath string, canSkipFileRead bool) *SrcFile {
	src_file := ensureSrcFile(srcFilePath, nil, canSkipFileRead)
	if src_file == nil { // file might be gone from diags by now
		refreshAndPublishNotices(srcFilePath)
	}
	return src_file
}
