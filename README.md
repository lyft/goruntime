<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Goruntime](#goruntime)
  - [Overview](#overview)
  - [Installation](#installation)
  - [Building](#building)
  - [Usage](#usage)
    - [Intended Use](#intended-use)
    - [Components](#components)
      - [Loader](#loader)
      - [Snapshot](#snapshot)
    - [Example of Usage](#example-of-usage)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Goruntime

## Overview

Goruntime is a Go client for Runtime application level feature flags and configuration.

## Installation

```
go get github.com/lyft/goruntime
```

## Building

```
make bootstrap && make tests
```

## Usage

In order to start using goruntime, import it to your project with:

```Go
import "github.com/lyft/runtime"
```

### Intended Use

The runtime system is meant to support small amounts of data, such
as feature flags, kill switches, regional configuration, experiment
settings, etc.  Individual files should typically contain a single key/value pair
(filename as key, content as value).

### Components

The runtime system is composed of a Loader interface, Runtime interface and a Snapshot interface. The Snapshot holds a version of
the runtime data from disk, and is used to retrieve information from that data. The Loader loads the current snapshot, and
gets file system updates when the runtime data gets updated. The Loader also uses the Refresher to watch the runtime directory
and refreshes the snapshot when prompted.

#### Refresher
The Refresher [interface](https://github.com/lyft/goruntime/blob/master/loader/refresher_iface.go) is defined like this:

```Go
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
```

The Refresher determines what directory to watch for file system changes and if there are any changes when to refresh.

Two refreshers are provided out of the box
* [Symlink Refresher](https://github.com/lyft/goruntime/blob/master/loader/symlink_refresher.go) : Watches the runtime directory as if it were a symlink and prompts a refresh if the symlink changes.
* [Directory Refresher](https://github.com/lyft/goruntime/blob/master/loader/directory_refresher.go) : Watches the runtime directory as a regular directory and prompts a refresh if the content of that directory change (not its subdirectories). 

#### Loader

The Loader [interface](https://github.com/lyft/goruntime/blob/master/loader/iface.go) is defined like this:

```Go
type IFace interface {
	// @return Snapshot the current snapshot. This reference is safe to use forever, but will grow
	//         stale so should not be stored beyond when it is immediately needed.
	Snapshot() snapshot.IFace

	// Add a channel that will be written to when a new snapshot is available. "1" will be written
	// to the channel as a sentinel.
	// @param callback supplies the callback to add.
	AddUpdateCallback(callback chan<- int)
}
```

To create a new Loader:

```Go
import (
        "github.com/lyft/goruntime/loader"
        "github.com/lyft/gostats"
)

// for full docs on gostats visit https://github.com/lyft/gostats
store := stats.NewDefaultStore()
runtime := loader.New("runtime_path", "runtime_subdirectory", store.Scope("runtime"), &DirectoryRefresher{})
```

The Loader will use filesystem events to update the filesystem snapshot it has.

#### Snapshot

The Snapshot [interface](https://github.com/lyft/goruntime/blob/master/snapshot/iface.go) is defined like this:

```Go
type IFace interface {
	FeatureEnabled(key string, defaultValue uint64) bool

	// Fetch raw runtime data based on key.
	// @param key supplies the key to fetch.
	// @return const std::string& the value or empty string if the key does not exist.
	Get(key string) string

	// Fetch an integer runtime key.
	// @param key supplies the key to fetch.
	// @param defaultValue supplies the value to return if the key does not exist or it does not
	//        contain an integer.
	// @return uint64 the runtime value or the default value.
	GetInteger(key string, defaultValue uint64) uint64

	// Fetch all keys inside the snapshot.
	// @return []string all of the keys.
	Keys() []string

	Entries() map[string]*entry.Entry

	SetEntry(string, *entry.Entry)
}
```

A Snapshot is composed of a map of [`Entry`s](https://github.com/lyft/goruntime/blob/master/snapshot/entry/entry.go).
Each entry represents a file in the runtime path. The Snapshot can be used to `Get` the value of an entry (or `GetInteger`
if the file contains an integer).

Keys are built by joining paths with `.` relative to the runtime subdirectory. For example if this is your filesystem:

```
/runtime/
└── config
    ├── file1
    └── more_files
        ├── file2
        └── file3
```

And the runtime loader is setup like so:

```Go
store := stats.NewDefaultStore()
runtime := loader.New("/runtime", "config", stats.Scope("runtime"))
```

The values in all three files can be obtained the following way:

```Go
s := runtime.Snapshot()
s.Get("file1")
s.Get("more_files.file2")
//Supposed file3 contains an integer, or you want to use a default integer if file3 does not contain one
s.GetInteger("more_files.file3", 8)
```
