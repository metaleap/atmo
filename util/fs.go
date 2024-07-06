package util

import (
	"io/fs"
	"os"
	"path/filepath"

	"atmo/util/str"
)

func FsDelFile(filePath string) { _ = os.Remove(filePath) }
func FsDelDir(dirPath string)   { _ = os.RemoveAll(dirPath) }

func FsCopy(srcFilePath string, dstFilePath string) {
	FsWrite(dstFilePath, FsRead(srcFilePath))
}

func FsRead(filePath string) []byte {
	data, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	return data
}

func FsWrite(filePath string, data []byte) {
	err := os.WriteFile(filePath, data, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func FsIsDir(dirPath string) bool   { return fsIs(dirPath, fs.FileInfo.IsDir, true) }
func FsIsFile(filePath string) bool { return fsIs(filePath, fs.FileInfo.IsDir, false) }

func fsIs(path string, check func(fs.FileInfo) bool, expect bool) bool {
	fs_info := fsStat(path)
	return (fs_info != nil) && (expect == check(fs_info))
}

func fsStat(path string) fs.FileInfo {
	fs_info, err := os.Stat(path)
	is_not_exist := os.IsNotExist(err)
	if err != nil && !is_not_exist {
		panic(err)
	}
	return If(is_not_exist, nil, fs_info)
}

func FsDirPathHome() string {
	ret, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return ret
}

func FsDirPathCur() string {
	ret, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return ret
}

func FsPathAbs(fsPath string) string {
	ret, err := filepath.Abs(fsPath)
	if err != nil {
		panic(err)
	}
	return ret
}

func FsPathSwapExt(filePath string, oldExtInclDot string, newExtInclDot string) string {
	if str.Ends(filePath, oldExtInclDot) {
		filePath = filePath[:len(filePath)-len(oldExtInclDot)] + newExtInclDot
	}
	return filePath
}

func FsIsNewerThan(file1Path string, file2Path string) bool {
	fs_info1, fs_info2 := fsStat(file1Path), fsStat(file2Path)
	return (fs_info1 == nil) || (fs_info1.IsDir()) || (fs_info2 == nil) || (fs_info2.IsDir()) ||
		fs_info1.ModTime().After(fs_info2.ModTime())
}

func FsDirEnsure(dirPath string) (did bool) {
	if FsIsDir(dirPath) { // wouldn't think you'd need this, with the below, but do
		return false
	}
	err := os.MkdirAll(dirPath, os.ModePerm)
	if (err != nil) && !os.IsExist(err) {
		panic(err)
	}
	return (err == nil)
}

func FsLinkEnsure(linkLocationPath string, pointsToPath string, ensureDirInstead bool) (did bool) {
	if ensureDirInstead {
		did = FsDirEnsure(pointsToPath)
	} else if !FsIsFile(linkLocationPath) {
		did = FsDirEnsure(filepath.Dir(linkLocationPath))
		points_to_path, link_location_path := FsPathAbs(pointsToPath), FsPathAbs(linkLocationPath)
		if err := os.Symlink(points_to_path, link_location_path); (err != nil) && !os.IsExist(err) {
			panic(err)
		} else {
			did = (err == nil)
		}
	}
	return
}

func FsDirWalk(dirPath string, onDirEntry func(fsPath string, fsEntry fs.DirEntry)) {
	if err := fs.WalkDir(os.DirFS(dirPath), ".", func(path string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}
		fs_path := filepath.Join(dirPath, path)
		if fs_path != dirPath { // dont want that DirEntry with Name()=="." in *our* walks
			onDirEntry(fs_path, dirEntry)
		}
		return nil
	}); err != nil {
		panic(err)
	}
}
