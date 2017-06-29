package loader

type FileSystemOp int32

// Filesystem operations that are monitored for changes
const (
	Create FileSystemOp = iota
	Write
	Remove
	Rename
	Chmod
)

// A Refresher is used to determine when to refresh the runtime
type Refresher interface {
	// @return The directory path to watch for changes.
	// @param runtimePath The root of the runtime path
	// @param appDirPath Any app specific path
	WatchDirectory(runtimePath string, appDirPath string) string

	// @return If the runtime needs to be refreshed
	// @param path The path that triggered the FileSystemOp
	// @param The Filesystem op that happened on the directory returned from WatchDirectory
	ShouldRefresh(path string, op FileSystemOp) bool
}
