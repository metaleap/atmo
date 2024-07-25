package util

import (
	"io/fs"
	"os"
	"path/filepath"

	"atmo/util/str"
)

func FsDelFile(filePath string) { _ = os.Remove(filePath) }
func FsDelDir(dirPath string)   { _ = os.RemoveAll(dirPath) }

func FsIsDir(dirPath string) bool   { return fsIs(dirPath, fs.FileInfo.IsDir, true) }
func FsIsFile(filePath string) bool { return fsIs(filePath, fs.FileInfo.IsDir, false) }

func fsIs(path string, check func(fs.FileInfo) bool, expect bool) bool {
	fs_info := fsStat(path)
	return (fs_info != nil) && (expect == check(fs_info))
}

func fsStat(path string) fs.FileInfo {
	fs_info, err := os.Stat(path)
	return If(err != nil, nil, fs_info)
}

func FsPathSwapExt(filePath string, oldExtInclDot string, newExtInclDot string) string {
	if str.Ends(filePath, oldExtInclDot) {
		filePath = filePath[:len(filePath)-len(oldExtInclDot)] + newExtInclDot
	}
	return filePath
}

func FsIsNewerThan(file1Path string, file2Path string) bool {
	fs_info1, fs_info2 := fsStat(file1Path), fsStat(file2Path)
	return (fs_info1 == nil) || (fs_info2 == nil) ||
		(fs_info1.IsDir()) || (fs_info2.IsDir()) ||
		fs_info1.ModTime().IsZero() || fs_info2.ModTime().IsZero() ||
		fs_info1.ModTime().After(fs_info2.ModTime())
}

func FsDirWalk(dirPath string, onDirEntry func(fsPath string, fsEntry fs.DirEntry)) error {
	dir_path, err := filepath.Abs(dirPath)
	if err != nil {
		return err
	}
	return fs.WalkDir(os.DirFS(dir_path), ".", func(path string, dirEntry fs.DirEntry, err error) error {
		if err == nil {
			fs_path := filepath.Join(dir_path, path)
			if fs_path != dir_path { // dont want that DirEntry with Name()=="." in *our* walks
				onDirEntry(fs_path, dirEntry)
			}
		}
		return nil
	})
}
