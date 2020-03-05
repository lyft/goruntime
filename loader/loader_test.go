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
	tmpdir, err := ioutil.TempDir("", "")
	assert.NoError(err)
	defer os.RemoveAll(tmpdir)

	tmpfile := filepath.Join(tmpdir, filepath.Base(path))

	err = ioutil.WriteFile(tmpfile, []byte(text), os.ModePerm)
	assert.NoError(err)

	err = os.MkdirAll(filepath.Dir(path), os.ModeDir|os.ModePerm)
	assert.NoError(err)

	// We use rename since creating a file and writing to it is too slow.
	// This is because creating the directory triggers the loader's watcher
	// causing it to scan the directory and if we need to create + write to
	// the file there is a chance the loader will store the contents of an
	// empty file, which is a race.
	//
	// This is okay because in prod we symlink files into place so we don't
	// need to worry about reading empty/partial files.
	//
	err = os.Rename(tmpfile, path)
	assert.NoError(err)
}

func TestNilRuntime(t *testing.T) {
	assert := require.New(t)

	loader := New("", "", nullScope, &SymlinkRefresher{RuntimePath: ""}, AllowDotFiles)
	snapshot := loader.Snapshot()
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

	loader := New(tempDir+"/current", "app", nullScope, &SymlinkRefresher{RuntimePath: tempDir + "/current"}, AllowDotFiles)
	runtime_update := make(chan int)
	loader.AddUpdateCallback(runtime_update)
	snapshot := loader.Snapshot()
	assert.Equal("", snapshot.Get("foo"))
	assert.Equal(uint64(5), snapshot.GetInteger("foo", 5))
	assert.Equal("hello", snapshot.Get("file1"))
	assert.Equal(uint64(6), snapshot.GetInteger("file1", 6))
	assert.Equal("world", snapshot.Get("dir.file2"))
	assert.Equal(uint64(7), snapshot.GetInteger("dir.file2", 7))
	assert.Equal(uint64(34), snapshot.GetInteger("dir2.file3", 100))

	info, _ := os.Stat(tempDir + "/testdir1/app/file1")
	assert.Equal(info.ModTime(), snapshot.GetModified("file1"))

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

	<-runtime_update

	time.Sleep(100 * time.Millisecond)

	snapshot = loader.Snapshot()
	assert.Equal("", snapshot.Get("foo"))
	assert.Equal("hello2", snapshot.Get("file1"))
	assert.Equal("world2", snapshot.Get("dir.file2"))
	assert.Equal(uint64(100), snapshot.GetInteger("dir2.file3", 0))
	assert.True(snapshot.FeatureEnabled("dir2.file3", 0))

	keys = snapshot.Keys()
	sort.Strings(keys)
	assert.EqualValues([]string{"dir.file2", "dir2.file3", "file1"}, keys)
}

func TestIgnoreDotfiles(t *testing.T) {
	assert := require.New(t)

	// Setup base test directory.
	tempDir, err := ioutil.TempDir("", "runtime_test")
	assert.NoError(err)
	defer os.RemoveAll(tempDir)

	// Make test files for runtime snapshot.
	makeFileInDir(assert, tempDir+"/testdir1/app/dir3/.file4", ".file4")
	makeFileInDir(assert, tempDir+"/testdir1/app/.dir/file5", ".dir")
	assert.NoError(err)

	loaderIgnoreDotfiles := New(tempDir+"/testdir1", "app", nullScope, &SymlinkRefresher{RuntimePath: tempDir + "/testdir1"}, IgnoreDotFiles)
	snapshot := loaderIgnoreDotfiles.Snapshot()
	assert.Equal("", snapshot.Get("dir3..file4"))
	assert.Equal("", snapshot.Get(".dir.file5"))

	loaderIncludeDotfiles := New(tempDir+"/testdir1", "app", nullScope, &SymlinkRefresher{RuntimePath: tempDir + "/testdir1"}, AllowDotFiles)
	snapshot = loaderIncludeDotfiles.Snapshot()
	assert.Equal(".file4", snapshot.Get("dir3..file4"))
	assert.Equal(".dir", snapshot.Get(".dir.file5"))
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

	loader := New(tempDir, "app", nullScope, &DirectoryRefresher{}, AllowDotFiles)
	runtime_update := make(chan int)
	loader.AddUpdateCallback(runtime_update)
	snapshot := loader.Snapshot()
	assert.Equal("", snapshot.Get("file1"))
	makeFileInDir(assert, appDir+"/file1", "hello")

	// Wait for the update
	<-runtime_update

	snapshot = loader.Snapshot()
	assert.Equal("hello", snapshot.Get("file1"))

	// Mimic a file change in directory
	makeFileInDir(assert, appDir+"/file2", "hello2")

	// Wait for the update
	<-runtime_update

	snapshot = loader.Snapshot()
	assert.Equal("hello2", snapshot.Get("file2"))

	// Write to the file
	f, err := os.OpenFile(appDir+"/file2", os.O_RDWR, os.ModeAppend)
	assert.NoError(err)
	_, err = f.WriteString("hello3")
	assert.NoError(err)
	f.Sync()

	// Wait for the update
	<-runtime_update

	snapshot = loader.Snapshot()
	assert.Equal("hello3", snapshot.Get("file2"))
}

func BenchmarkSnapsot(b *testing.B) {
	var ll Loader
	for i := 0; i < b.N; i++ {
		ll.Snapshot()
	}
}

func BenchmarkSnapsot_Parallel(b *testing.B) {
	ll := new(Loader)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ll.Snapshot()
		}
	})
}
