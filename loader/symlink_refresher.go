package loader

import "path/filepath"

type SymlinkRefresher struct {
	RuntimePath string
}

func (s *SymlinkRefresher) WatchDirectory(runtimePath string, appDirPath string) string {
	return filepath.Dir(runtimePath)
}

func (s *SymlinkRefresher) ShouldRefresh(path string, op FileSystemOp) bool {
	if path == s.RuntimePath &&
		(op == Write || op == Create) {
		return true
	}
	return false
}
