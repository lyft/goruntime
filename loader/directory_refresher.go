package loader

import "path/filepath"

type DirectoryRefresher struct {
	currDir  string
	watchOps map[FileSystemOp]bool
}

var defaultFileSystemOps = map[FileSystemOp]bool{
	Write:  true,
	Create: true,
	Chmod:  true,
}

func (d *DirectoryRefresher) WatchDirectory(runtimePath string, appDirPath string) string {
	d.currDir = filepath.Join(runtimePath, appDirPath)
	return d.currDir
}

func (d *DirectoryRefresher) WatchFileSystemOps(fsops ...FileSystemOp) {
	d.watchOps = map[FileSystemOp]bool{}
	for _, op := range fsops {
		d.watchOps[op] = true
	}
}

func (d *DirectoryRefresher) ShouldRefresh(path string, op FileSystemOp) bool {
	watchOps := d.watchOps
	if watchOps == nil {
		watchOps = defaultFileSystemOps
	}
	return filepath.Dir(path) == d.currDir && watchOps[op]
}
