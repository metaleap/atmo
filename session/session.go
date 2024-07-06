package atmo_session

import (
	"path/filepath"

	"atmo/util"
)

type Session struct {
	allSrcPkgs  map[string]*SrcPackage
	allSrcFiles map[string]*SrcFile
}

func New() *Session {
	me := Session{
		allSrcPkgs: map[string]*SrcPackage{}, allSrcFiles: map[string]*SrcFile{},
	}
	return &me
}

type SrcPackage struct {
	DirPath string
	Files   []*SrcFile
}

type SrcFile struct {
	FilePath string
	Pkg      *SrcPackage
	Content  string
}

func (me *Session) OnSourceFileEvents(removed []string, added []string, changed []string) {
}

func (me *Session) OnSourceFileEdit(srcFilePath string, curFullContent string) {
	me.ensureSrcFile(srcFilePath, &curFullContent)
}

func (me *Session) ensureSrcPkg(dirPath string) *SrcPackage {
	util.Assert(filepath.IsAbs(dirPath), dirPath)
	src_pkg := me.allSrcPkgs[dirPath]
	dir_exists := util.FsIsDir(dirPath)
	if src_pkg != nil && !dir_exists {
		me.allSrcPkgs[dirPath] = nil
		for file_path := range me.allSrcFiles {
			if filepath.Clean(dirPath) == filepath.Clean(filepath.Dir(file_path)) {
				delete(me.allSrcFiles, file_path)
			}
		}
		for _, src_file := range src_pkg.Files {
			delete(me.allSrcFiles, src_file.FilePath)
		}
		src_pkg = nil
	}
	if src_pkg == nil && dir_exists {
		src_pkg = &SrcPackage{DirPath: dirPath}

	}
	return src_pkg
}

func (me *Session) ensureSrcFile(filePath string, curFullContent *string) *SrcFile {
	util.Assert(filepath.IsAbs(filePath), filePath)
	src_file := me.allSrcFiles[filePath]
	if src_file == nil {
		pkg_dir_path := filepath.Dir(filePath)
		src_file = &SrcFile{FilePath: filePath, Pkg: me.allSrcPkgs[pkg_dir_path]}
		me.allSrcFiles[filePath] = src_file
		if src_file.Pkg == nil {
			src_file.Pkg = &SrcPackage{DirPath: pkg_dir_path}
		}
	}
	return src_file
}
