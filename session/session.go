package session

import (
	"cmp"
	"io/fs"
	"sync"

	"atmo/util"
	"atmo/util/kv"
	"atmo/util/sl"
)

var (
	state struct {
		stateAccess
		srcFiles map[string]*SrcFile
		srcPacks map[string]*SrcPack
	}
)

type StateAccess interface {
	OnSrcFileEdit(srcFilePath string, curFullContent string)
	OnSrcFileEvents(removed []string, canSkipFileRead bool, current ...string)

	AllCurrentSrcFileNotices() map[string]SrcFileNotices
	AllCurrentSrcPacks() []*SrcPack
	PacksFsRefresh()
	GetSrcPack(dirPath string, loadIfMissing bool) *SrcPack
	Interpreter(dirPath string) *Interp
	SrcFile(srcFilePath string, canSkipFileRead bool) *SrcFile
}

func init() {
	state.srcFiles, state.srcPacks = map[string]*SrcFile{}, map[string]*SrcPack{}
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
	packsFsRefresh()
	removeSrcFiles(removed...) // does refreshAndPublishNotices for removed
	refreshAndPublishNotices(ensureSrcFiles(nil, canSkipFileRead, current...)...)
}

func (*stateAccess) AllCurrentSrcFileNotices() map[string]SrcFileNotices {
	return allNotices
}

func (*stateAccess) AllCurrentSrcPacks() []*SrcPack {
	return sl.SortedPer(kv.Values(state.srcPacks), func(pack1 *SrcPack, pack2 *SrcPack) int {
		return cmp.Compare(pack1.DirPath, pack2.DirPath)
	})
}

func (*stateAccess) PacksFsRefresh() {
	packsFsRefresh()
}

func (*stateAccess) GetSrcPack(dirPath string, loadIfMissing bool) (ret *SrcPack) {
	ret = state.srcPacks[dirPath]
	if (ret == nil) && loadIfMissing {
		var src_file_paths []string
		util.FsDirWalk(dirPath, func(fsPath string, fsEntry fs.DirEntry) {
			if IsSrcFilePath(fsPath) {
				src_file_paths = append(src_file_paths, fsPath)
			}
		})
		if refr_diags_for := ensureSrcFiles(nil, true, src_file_paths...); len(refr_diags_for) > 0 {
			refreshAndPublishNotices(refr_diags_for...)
		}
	}
	return
}

func (me *stateAccess) Interpreter(dirPath string) *Interp {
	src_file_path := newSrcFilePathFakeAndReplish(dirPath)
	me.SrcFile(src_file_path, true)
	src_file := state.srcFiles[src_file_path]
	util.Assert(src_file != nil, nil)
	return newInterp(src_file, nil)
}

func (*stateAccess) SrcFile(srcFilePath string, canSkipFileRead bool) *SrcFile {
	refr_diags_for := ensureSrcFiles(nil, canSkipFileRead, srcFilePath)
	src_file := state.srcFiles[srcFilePath]
	if (src_file == nil) || (len(refr_diags_for) > 0) { // the latter, if non-empty, WILL have srcFilePath
		refreshAndPublishNotices(refr_diags_for...)
	}
	return src_file
}
