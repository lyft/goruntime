package loader

import "path/filepath"

type DirectoryRefresher struct {
}

func (d *DirectoryRefresher) WatchDirectory(runtimePath string, appDirPath string) string {
	return filepath.Join(runtimePath, appDirPath)
}

func (d *DirectoryRefresher) ShouldRefresh(path string, op FileSystemOp) bool {
	if op == Write || op == Create || op == Chmod {
		return true
	}
	return false
}
