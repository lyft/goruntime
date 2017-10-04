package loader

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"sort"

	"time"

	stats "github.com/lyft/gostats"
	logger "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

var nullScope = stats.NewStore(stats.NewNullSink(), false)

func init() {
	lvl, _ := logger.ParseLevel("DEBUG")
	logger.SetLevel(lvl)
}

func makeFileInDir(assert *require.Assertions, path string, text string) {
	err := os.MkdirAll(filepath.Dir(path), os.ModeDir|os.ModePerm)
	assert.NoError(err)

	err = ioutil.WriteFile(path, []byte(text), os.ModePerm)
	assert.NoError(err)
}

func TestNilRuntime(t *testing.T) {
	assert := require.New(t)

	testLoader := New("", "", nullScope, &SymlinkRefresher{RuntimePath: ""})
	snapshot := testLoader.Snapshot()
	assert.Equal("", snapshot.Get("foo"))
	assert.Equal(uint64(100), snapshot.GetInteger("bar", 100))
	assert.True(snapshot.FeatureEnabled("baz", 100))
	assert.False(snapshot.FeatureEnabled("blah", 0))
}

func TestSymlinkRefresher(t *testing.T) {
	assert := require.New(t)

	// Setup base test directory.
	tempDir, err := ioutil.TempDir("", "runtime_test")
	assert.NoError(err)
	defer os.RemoveAll(tempDir)

	// Make test files for first runtime snapshot.
	makeFileInDir(assert, tempDir+"/testdir1/app/file1", "hello")
	makeFileInDir(assert, tempDir+"/testdir1/app/dir/file2", "world")
	makeFileInDir(assert, tempDir+"/testdir1/app/dir2/file3", "\n 34  ")
	assert.NoError(err)
	err = os.Symlink(tempDir+"/testdir1", tempDir+"/current")
	assert.NoError(err)

	testLoader := New(tempDir+"/current", "app", nullScope, &SymlinkRefresher{RuntimePath: tempDir + "/current"})
	runtimeUpdate := make(chan int)
	testLoader.AddUpdateCallback(runtimeUpdate)
	snapshot := testLoader.Snapshot()
	assert.Equal("", snapshot.Get("foo"))
	assert.Equal(uint64(5), snapshot.GetInteger("foo", 5))
	assert.Equal("hello", snapshot.Get("file1"))
	assert.Equal(uint64(6), snapshot.GetInteger("file1", 6))
	assert.Equal("world", snapshot.Get("dir.file2"))
	assert.Equal(uint64(7), snapshot.GetInteger("dir.file2", 7))
	assert.Equal(uint64(34), snapshot.GetInteger("dir2.file3", 100))

	keys := snapshot.Keys()
	sort.Strings(keys)
	assert.EqualValues([]string{"dir.file2", "dir2.file3", "file1"}, keys)

	//// Make test files for second runtime snapshot.
	makeFileInDir(assert, tempDir+"/testdir2/app/file1", "hello2")
	makeFileInDir(assert, tempDir+"/testdir2/app/dir/file2", "world2")
	makeFileInDir(assert, tempDir+"/testdir2/app/dir2/file3", "100")
	err = os.Symlink(tempDir+"/testdir2", tempDir+"/current_new")
	assert.NoError(err)
	err = os.Rename(tempDir+"/current_new", tempDir+"/current")
	assert.NoError(err)

	<-runtimeUpdate

	time.Sleep(100 * time.Millisecond)

	snapshot = testLoader.Snapshot()
	assert.Equal("", snapshot.Get("foo"))
	assert.Equal("hello2", snapshot.Get("file1"))
	assert.Equal("world2", snapshot.Get("dir.file2"))
	assert.Equal(uint64(100), snapshot.GetInteger("dir2.file3", 0))
	assert.True(snapshot.FeatureEnabled("dir2.file3", 0))

	keys = snapshot.Keys()
	sort.Strings(keys)
	assert.EqualValues([]string{"dir.file2", "dir2.file3", "file1"}, keys)
}

func TestDirectoryRefresher(t *testing.T) {
	assert := require.New(t)

	// Setup base test directory.
	tempDir, err := ioutil.TempDir("", "dir_runtime_test")
	assert.NoError(err)
	defer os.RemoveAll(tempDir)

	appDir := tempDir + "/app"
	err = os.MkdirAll(appDir, os.ModeDir|os.ModePerm)
	assert.NoError(err)

	testLoader := New(tempDir, "app", nullScope, &DirectoryRefresher{})
	runtimeUpdate := make(chan int)
	testLoader.AddUpdateCallback(runtimeUpdate)
	snapshot := testLoader.Snapshot()
	assert.Equal("", snapshot.Get("file1"))
	makeFileInDir(assert, appDir+"/file1", "hello")

	// Wait for the update
	<-runtimeUpdate

	snapshot = testLoader.Snapshot()
	assert.Equal("hello", snapshot.Get("file1"))

	// Mimic a file change in directory
	makeFileInDir(assert, appDir+"/file2", "hello2")

	// Wait for the update
	<-runtimeUpdate

	snapshot = testLoader.Snapshot()
	assert.Equal("hello2", snapshot.Get("file2"))
}
