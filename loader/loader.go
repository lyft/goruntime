package loader

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/lyft/goruntime/snapshot"
	"github.com/lyft/goruntime/snapshot/entry"
	stats "github.com/lyft/gostats"

	logger "github.com/sirupsen/logrus"
)

type loaderStats struct {
	loadAttempts stats.Counter
	loadFailures stats.Counter
	numValues    stats.Gauge
}

func newLoaderStats(scope stats.Scope) loaderStats {
	ret := loaderStats{}
	ret.loadAttempts = scope.NewCounter("load_attempts")
	ret.loadFailures = scope.NewCounter("load_failures")
	ret.numValues = scope.NewGauge("num_values")
	return ret
}

// Implementation of Loader that watches a symlink and reads from the filesystem.
type Loader struct {
	currentSnapshot atomic.Value
	watcher         *fsnotify.Watcher
	watchPath       string
	subdirectory    string
	nextSnapshot    snapshot.IFace
	callbacks       []chan<- int
	stats           loaderStats
	ignoreDotfiles  bool
}

func (l *Loader) Snapshot() snapshot.IFace {
	v, _ := l.currentSnapshot.Load().(snapshot.IFace)
	return v
}

func (l *Loader) AddUpdateCallback(callback chan<- int) {
	l.callbacks = append(l.callbacks, callback)
}

func (l *Loader) onRuntimeChanged() {
	targetDir := filepath.Join(l.watchPath, l.subdirectory)
	logger.Debugf("runtime changed. loading new snapshot at %s",
		targetDir)

	l.nextSnapshot = snapshot.New()
	filepath.Walk(targetDir, l.walkDirectoryCallback)

	l.stats.loadAttempts.Inc()
	l.stats.numValues.Set(uint64(len(l.nextSnapshot.Entries())))
	l.currentSnapshot.Store(l.nextSnapshot)

	l.nextSnapshot = nil
	for _, callback := range l.callbacks {
		// Arbitrary integer just to wake up channel.
		callback <- 1
	}
}

type walkError struct {
	err error
}

func (l *Loader) walkDirectoryCallback(path string, info os.FileInfo, err error) error {
	if err != nil {
		l.stats.loadFailures.Inc()
		logger.Warnf("runtime: error processing %s: %s", path, err)

		return nil
	}

	logger.Debugf("runtime: processing %s", path)
	if l.ignoreDotfiles && info.IsDir() && strings.HasPrefix(info.Name(), ".") {
		return filepath.SkipDir
	}

	if !info.IsDir() {
		if l.ignoreDotfiles && strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		contents, err := ioutil.ReadFile(path)

		if err != nil {
			l.stats.loadFailures.Inc()
			logger.Warnf("runtime: error reading %s: %s", path, err)

			return nil
		}

		key, err := filepath.Rel(filepath.Join(l.watchPath, l.subdirectory), path)

		if err != nil {
			l.stats.loadFailures.Inc()
			logger.Warnf("runtime: error parsing path %s: %s", path, err)

			return nil
		}

		key = strings.Replace(key, "/", ".", -1)
		stringValue := string(contents)
		e := &entry.Entry{
			StringValue: stringValue,
			Uint64Value: 0,
			Uint64Valid: false,
			Modified:    info.ModTime(),
		}

		uint64Value, err := strconv.ParseUint(strings.TrimSpace(stringValue), 10, 64)
		if err == nil {
			e.Uint64Value = uint64Value
			e.Uint64Valid = true
		}

		logger.Debugf("runtime: adding key=%s value=%s uint=%t", key,
			stringValue, e.Uint64Valid)
		l.nextSnapshot.SetEntry(key, e)
	}

	return nil
}

func getFileSystemOp(ev fsnotify.Event) FileSystemOp {
	switch ev.Op {
	case ev.Op & fsnotify.Write:
		return Write
	case ev.Op & fsnotify.Create:
		return Create
	case ev.Op & fsnotify.Chmod:
		return Chmod
	case ev.Op & fsnotify.Remove:
		return Remove
	case ev.Op & fsnotify.Rename:
		return Rename
	}
	return -1
}

type Option func(l *Loader)

func AllowDotFiles(l *Loader)  { l.ignoreDotfiles = false }
func IgnoreDotFiles(l *Loader) { l.ignoreDotfiles = true }

func New2(runtimePath, runtimeSubdirectory string, scope stats.Scope, refresher Refresher, opts ...Option) (IFace, error) {
	if runtimePath == "" || runtimeSubdirectory == "" {
		logger.Warn("no runtime configuration. using nil loader.")
		return NewNil(), nil
	}
	watchedPath := refresher.WatchDirectory(runtimePath, runtimeSubdirectory)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		// If this fails with EMFILE (0x18) it is likely due to
		// inotify_init1() and fs.inotify.max_user_instances.
		//
		// Include the error message, type and value - this is
		// particularly useful if the error is a syscall.Errno.
		return nil, fmt.Errorf("unable to create runtime watcher: %[1]s (%[1]T %#[1]v)\n", err)
	}

	err = watcher.Add(watchedPath)
	if err != nil {
		return nil, fmt.Errorf("unable to watch file (%[1]s): %[2]s (%[2]T %#[2]v)", watchedPath, err)
	}

	newLoader := Loader{
		watcher:      watcher,
		watchPath:    runtimePath,
		subdirectory: runtimeSubdirectory,
		stats:        newLoaderStats(scope),
	}

	for _, opt := range opts {
		opt(&newLoader)
	}

	newLoader.onRuntimeChanged()

	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				logger.Debugf("Got event %s", ev)
				if refresher.ShouldRefresh(ev.Name, getFileSystemOp(ev)) {
					newLoader.onRuntimeChanged()
				}
			case err := <-watcher.Errors:
				logger.Warnf("runtime watch error: %s", err)
			}
		}
	}()

	return &newLoader, nil
}

// Deprecated: use New2 instead
func New(runtimePath string, runtimeSubdirectory string, scope stats.Scope, refresher Refresher, opts ...Option) IFace {
	loader, err := New2(runtimePath, runtimeSubdirectory, scope, refresher, opts...)
	if err != nil {
		logger.Panic(err)
	}
	return loader
}
