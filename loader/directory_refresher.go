package loader

import "path/filepath"

type DirectoryRefresher struct {
	currDir string
}

func (d *DirectoryRefresher) WatchDirectory(runtimePath string, appDirPath string) string {
	d.currDir = filepath.Join(runtimePath, appDirPath)
	return d.currDir
}

func (d *DirectoryRefresher) ShouldRefresh(path string, op FileSystemOp) bool {
	if filepath.Dir(path) == d.currDir &&
		(op == Write || op == Create || op == Chmod) {
		return true
	}
	return false
}
