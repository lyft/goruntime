package loader

import "path/filepath"

type DirectoryRefresher struct {
	currDir  string
	watchOps map[FileSystemOp]struct{}
}

var defaultFileSystemOps = map[FileSystemOp]struct{}{
	Write:  {},
	Create: {},
	Chmod:  {},
}

func (d *DirectoryRefresher) WatchDirectory(runtimePath string, appDirPath string) string {
	d.currDir = filepath.Join(runtimePath, appDirPath)
	return d.currDir
}

func (d *DirectoryRefresher) WatchFileSystemOps(fsops ...FileSystemOp) map[FileSystemOp]struct{} {
	d.watchOps = map[FileSystemOp]struct{}{}
	for _, op := range fsops {
		d.watchOps[op] = struct{}{}
	}

	return d.watchOps
}

func (d *DirectoryRefresher) ShouldRefresh(path string, op FileSystemOp) bool {
	var watchOps map[FileSystemOp]struct{}

	if d.watchOps == nil {
		watchOps = defaultFileSystemOps
	} else {
		watchOps = d.watchOps
	}

	if _, opMatches := watchOps[op]; opMatches && filepath.Dir(path) == d.currDir {
		return true
	}
	return false
}
