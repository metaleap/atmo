package session

import (
	"cmp"
	"sync"

	"atmo/util/kv"
	"atmo/util/sl"
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

	AllCurrentSrcFileNotices() map[string]SrcFileNotices
	AllCurrentSrcPkgs() []*SrcPkg
	PkgsFsRefresh()
	GetSrcPkg(dirPath string) *SrcPkg
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
	refreshAndPublishNotices(ensureSrcFiles(&curFullContent, true, srcFilePath)...)
}

func (*stateAccess) OnSrcFileEvents(removed []string, canSkipFileRead bool, current ...string) {
	pkgsFsRefresh()
	removeSrcFiles(removed...) // does refreshAndPublishNotices for removed
	refreshAndPublishNotices(ensureSrcFiles(nil, canSkipFileRead, current...)...)
}

func (*stateAccess) AllCurrentSrcFileNotices() map[string]SrcFileNotices {
	return allNotices
}

func (*stateAccess) AllCurrentSrcPkgs() []*SrcPkg {
	return sl.SortedPer(kv.Values(state.srcPkgs), func(pkg1 *SrcPkg, pkg2 *SrcPkg) int {
		return cmp.Compare(pkg1.DirPath, pkg2.DirPath)
	})
}

func (*stateAccess) PkgsFsRefresh() {
	pkgsFsRefresh()
}

func (*stateAccess) GetSrcPkg(dirPath string) *SrcPkg {
	return state.srcPkgs[dirPath]
}

func (*stateAccess) SrcFile(srcFilePath string, canSkipFileRead bool) *SrcFile {
	refr_diags_for := ensureSrcFiles(nil, canSkipFileRead, srcFilePath)
	src_file := state.srcFiles[srcFilePath]
	if (src_file == nil) || (len(refr_diags_for) > 0) { // the latter, if non-empty, WILL have srcFilePath
		refreshAndPublishNotices(refr_diags_for...)
	}
	return src_file
}
