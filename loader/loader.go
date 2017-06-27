package loader

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/lyft/goruntime/snapshot"
	"github.com/lyft/goruntime/snapshot/entry"
	stats "github.com/lyft/gostats"

	logger "github.com/Sirupsen/logrus"
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
	watchPath       string
	subdirectory    string
	currentSnapshot snapshot.IFace
	nextSnapshot    snapshot.IFace
	updateLock      sync.RWMutex
	callbacks       []chan<- int
	stats           loaderStats
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
	targetDir := filepath.Join(l.watchPath, l.subdirectory)
	logger.Debugf("runtime changed. loading new snapshot at %s",
		targetDir)

	l.nextSnapshot = snapshot.New()
	filepath.Walk(targetDir, l.walkDirectoryCallback)

	// This could probably be done with an atomic pointer but the unsafe pointers the atomics
	// take scared me so skipping for now.
	l.stats.loadAttempts.Inc()
	l.stats.numValues.Set(uint64(len(l.nextSnapshot.Entries())))
	l.updateLock.Lock()
	l.currentSnapshot = l.nextSnapshot
	l.updateLock.Unlock()

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
	if !info.IsDir() {
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
		entry := entry.New(stringValue, 0, false)

		uint64Value, err := strconv.ParseUint(strings.TrimSpace(stringValue), 10, 64)
		if err == nil {
			entry.Uint64Value = uint64Value
			entry.Uint64Valid = true
		}

		logger.Debugf("runtime: adding key=%s value=%s uint=%t", key,
			stringValue, entry.Uint64Valid)
		l.nextSnapshot.SetEntry(key, entry)
	}

	return nil
}

// Function checks if the runtime needs to be reloaded.
// For Production the check is to see if runtime which is a symlink has changed
// For devbox/onebox the check is relaxed to see if any file has been modified then reload.
func reloadRuntime(ev fsnotify.Event, runtimePath string, is_dev bool) bool {
	if is_dev {
		if ev.Op&fsnotify.Write == fsnotify.Write || ev.Op&fsnotify.Create == fsnotify.Create || ev.Op&fsnotify.Chmod == fsnotify.Chmod {
			return true
		}
	} else {
		if ev.Name == runtimePath &&
			(ev.Op&fsnotify.Write == fsnotify.Write || ev.Op&fsnotify.Create == fsnotify.Create) {
			return true
		}
	}
	return false
}

func New(runtimePath string, runtimeSubdirectory string, scope stats.Scope, opts ...string) IFace {
	if runtimePath == "" || runtimeSubdirectory == "" {
		logger.Warnf("no runtime configuration. using nil loader.")
		return NewNil()
	}
	is_dev := false
	watchedPath := runtimePath
	for _, opt := range opts {
		if opt == "development" {
			is_dev = true
			// Need to watch the actual subdir where the changes happen in dev
			watchedPath = filepath.Join(runtimePath, runtimeSubdirectory)
			logger.Infof("Running in dev environment. Watching %s", watchedPath)
			break
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Fatalf("unable to create runtime watcher %+v", err)
	}

	if is_dev {
		// In dev environments watch the actual path
		err = watcher.Add(watchedPath)
	} else {
		// We need to watch the directory that the symlink is in vs. the symlink itself.
		err = watcher.Add(filepath.Dir(watchedPath))
	}

	if err != nil {
		logger.Fatalf("unable to create runtime watcher %+v", err)
	}

	newLoader := Loader{
		watcher, runtimePath, runtimeSubdirectory, nil, nil, sync.RWMutex{}, nil,
		newLoaderStats(scope)}
	newLoader.onRuntimeChanged()

	go func() {
		for {
			select {
			case ev := <-watcher.Events:
				logger.Debugf("Got event %s", ev)
				if reloadRuntime(ev, runtimePath, is_dev) {
					newLoader.onRuntimeChanged()
				}
			case err := <-watcher.Errors:
				logger.Warnf("runtime watch error:", err)
			}
		}
	}()

	return &newLoader
}
