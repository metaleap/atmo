package session

import (
	"cmp"
	"path/filepath"
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

	AllCurrentSrcFileDiags() map[string]Diags
	AllCurrentSrcPacks() []*SrcPack
	PacksFsRefresh()
	GetSrcPack(dirPath string, loadIfMissing bool) *SrcPack
	Interpreter(dirPath string) *Interp
	SrcFile(srcFilePath string) *SrcFile
}

func init() {
	state.srcFiles, state.srcPacks = map[string]*SrcFile{}, map[string]*SrcPack{}
}

func Access(do func(sess StateAccess, intel Intel)) {
	state.Lock()
	defer state.Unlock()
	do(&state.stateAccess, intel{})
}

type stateAccess struct{ sync.Mutex }

func (*stateAccess) OnSrcFileEdit(srcFilePath string, curFullContent string) {
	refreshAndPublishDiags(false, ensureSrcFiles(&curFullContent, true, srcFilePath)...)
}

func (*stateAccess) OnSrcFileEvents(removed []string, canSkipFileRead bool, current ...string) {
	packsFsRefresh()
	removeSrcFiles(removed...) // does refreshAndPublishDiags for removed
	refreshAndPublishDiags(false, ensureSrcFiles(nil, canSkipFileRead, current...)...)
}

func (*stateAccess) AllCurrentSrcFileDiags() map[string]Diags {
	return allDiags
}

func (*stateAccess) AllCurrentSrcPacks() []*SrcPack {
	return sl.SortedPer(kv.Values(state.srcPacks), func(pack1 *SrcPack, pack2 *SrcPack) int {
		return cmp.Compare(pack1.DirPath, pack2.DirPath)
	})
}

func (*stateAccess) PacksFsRefresh() {
	packsFsRefresh()
}

func (*stateAccess) GetSrcPack(packDirPath string, loadIfMissing bool) (ret *SrcPack) {
	util.Assert(filepath.IsAbs(packDirPath), nil)
	ret = state.srcPacks[packDirPath]
	if (ret == nil) && loadIfMissing {
		var src_file_paths []string
		for _, src_file_path := range util.FsDirFilesOnlyList(packDirPath) {
			println(src_file_path)
			if IsSrcFilePath(src_file_path) {
				src_file_paths = append(src_file_paths, src_file_path)
			}
		}
		if refr_diags_for := ensureSrcFiles(nil, true, src_file_paths...); len(refr_diags_for) > 0 {
			refreshAndPublishDiags(false, refr_diags_for...)
		}
		ret = state.srcPacks[packDirPath]
	}
	return
}

func (me *stateAccess) Interpreter(packDirPath string) *Interp {
	util.Assert(filepath.IsAbs(packDirPath), nil)
	src_pack := me.GetSrcPack(packDirPath, true)
	if src_pack != nil && src_pack.Interp != nil {
		return src_pack.Interp
	}

	src_file_path := newInterpFauxFilePath(packDirPath)
	_ = ensureSrcFiles(nil, true, src_file_path)
	src_file := state.srcFiles[src_file_path]
	if src_pack == nil {
		src_pack = me.GetSrcPack(packDirPath, true) // do this again in case the previous was `nil`, now it shouldnt be
	}
	util.Assert(src_file != nil, nil)
	util.Assert(src_pack != nil, nil)
	util.Assert(src_file.pack == src_pack, nil)
	defer refreshAndPublishDiags(false, src_pack.srcFilePaths()...)
	if src_pack.Interp != nil {
		return src_pack.Interp
	}
	return newInterp(src_pack, src_file)
}

func (*stateAccess) SrcFile(srcFilePath string) *SrcFile {
	refr_diags_for := ensureSrcFiles(nil, true, srcFilePath)
	src_file := state.srcFiles[srcFilePath]
	if (src_file == nil) || (len(refr_diags_for) > 0) { // the latter, if non-empty, WILL have srcFilePath
		refreshAndPublishDiags(false, refr_diags_for...)
	}
	return src_file
}
