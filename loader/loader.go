package loader

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

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
	watcher         *fsnotify.Watcher
	targetDirectory string
	currentSnapshot snapshot.IFace
	nextSnapshot    snapshot.IFace
	updateLock      sync.RWMutex
	callbacks       []chan<- int
	stats           loaderStats
	buf             bytes.Buffer
	ignoreDotfiles  bool
}

func (l *Loader) Snapshot() snapshot.IFace {
	// This could probably be done with an atomic pointer but the unsafe pointers the atomics
	// take scared me so skipping for now.
	l.updateLock.RLock()
	defer l.updateLock.RUnlock()
	return l.currentSnapshot
}

func (l *Loader) AddUpdateCallback(callback chan<- int) {
	l.callbacks = append(l.callbacks, callback)
}

func (l *Loader) onRuntimeChanged() {
	logger.Debugf("runtime changed. loading new snapshot at %s", l.targetDirectory)

	l.nextSnapshot = snapshot.New()
	filepath.Walk(l.targetDirectory, l.walkDirectoryCallback)

	// This could probably be done with an atomic pointer but the unsafe pointers the atomics
	// take scared me so skipping for now.
	l.stats.loadAttempts.Inc()
	l.stats.numValues.Set(uint64(len(l.nextSnapshot.Entries())))
	l.updateLock.Lock()
	l.currentSnapshot = l.nextSnapshot
	l.updateLock.Unlock()
	l.clearBuffer()

	l.nextSnapshot = nil
	for _, callback := range l.callbacks {
		// Arbitrary integer just to wake up channel.
		callback <- 1
	}
}

type walkError struct {
	err error
}

var keyReplacer = strings.NewReplacer("/", ".")

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

		contents, err := l.readFile(path)
		if err != nil {
			l.stats.loadFailures.Inc()
			logger.Warnf("runtime: error reading %s: %s", path, err)

			return nil
		}

		key, err := filepath.Rel(l.targetDirectory, path)
		if err != nil {
			l.stats.loadFailures.Inc()
			logger.Warnf("runtime: error parsing path %s: %s", path, err)

			return nil
		}

		key = keyReplacer.Replace(key)
		e := &entry.Entry{
			StringValue: contents,
			Uint64Value: 0,
			Uint64Valid: false,
			Modified:    info.ModTime(),
		}

		uint64Value, err := strconv.ParseUint(strings.TrimSpace(contents), 10, 64)
		if err == nil {
			e.Uint64Value = uint64Value
			e.Uint64Valid = true
		}

		logger.Debugf("runtime: adding key=%s value=%s uint=%t", key,
			contents, e.Uint64Valid)
		l.nextSnapshot.SetEntry(key, e)
	}

	return nil
}

func (l *Loader) readFile(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	l.buf.Reset()
	l.buf.Grow(4096)

	if _, err := l.buf.ReadFrom(f); err != nil {
		return "", err
	}

	return l.buf.String(), nil
}

// clearBuffer clears the current buffer so that it may be GC'd.
func (l *Loader) clearBuffer() { l.buf = bytes.Buffer{} }

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
		watcher:         watcher,
		targetDirectory: filepath.Join(runtimePath, runtimeSubdirectory),
		stats:           newLoaderStats(scope),
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
