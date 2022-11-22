package aferofuse

import (
	"os"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/pathfs"
	"github.com/matryer/is"
	"github.com/spf13/afero"
)

type testableAferoFs struct {
	afero.Fs
	name         string
	stat         os.FileInfo
	err          error
	calls        []string
	file         afero.File
	link         string
	openFileInt  int
	openFileMode os.FileMode
	symlinkOld   string
	symlinkNew   string
}

func NewTestableAferoFs() *testableAferoFs {
	return &testableAferoFs{
		Fs: afero.NewMemMapFs(),
	}
}

func (t *testableAferoFs) SymlinkIfPossible(a string, b string) error {
	t.symlinkOld, t.symlinkNew = a, b
	return t.err
}
func (t *testableAferoFs) ReadlinkIfPossible(string) (string, error)  { return t.link, t.err }
func (t *testableAferoFs) Name() string                               { return t.name }
func (t *testableAferoFs) Stat(string) (os.FileInfo, error)           { return t.stat, t.err }
func (t *testableAferoFs) Chmod(string, os.FileMode) error            { return t.err }
func (t *testableAferoFs) Chown(string, int, int) error               { return t.err }
func (t *testableAferoFs) Chtimes(string, time.Time, time.Time) error { return t.err }
func (t *testableAferoFs) Mkdir(string, os.FileMode) error            { return t.err }
func (t *testableAferoFs) Remove(string) error                        { return t.err }
func (t *testableAferoFs) Rename(string, string) error                { return t.err }
func (t *testableAferoFs) Open(string) (afero.File, error)            { return t.file, t.err }
func (t *testableAferoFs) OpenFile(a string, b int, c os.FileMode) (afero.File, error) {
	t.calls = append(t.calls, "OpenFile")
	t.openFileInt, t.openFileMode = b, c

	return t.file, t.err
}

func TestTestableAferoFsInterface(t *testing.T) {
	fn := func(afero.Fs) {}

	fn(&testableAferoFs{})
}

func fsBootstrap(t *testing.T) (*is.I, *testableAferoFs, pathfs.FileSystem) {
	is := is.New(t)
	afs := &testableAferoFs{}
	fuseFs := NewFuseFileSystem(afs)

	return is, afs, fuseFs
}

func TestTestableAferoName(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	is.Equal(fuseFs.String(), afs.Name())
}

func testErr(is *is.I, afs *testableAferoFs, fuseFs pathfs.FileSystem, fn func() fuse.Status) {
	afs.err = os.ErrNotExist
	is.Equal(fn(), fuse.ToStatus(afs.err))
	afs.err = nil
	is.Equal(fn(), fuse.ToStatus(afs.err))
}

func TestTestableAferoChmod(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	testErr(is, afs, fuseFs, func() fuse.Status { return fuseFs.Chmod("sadf", 0644, nil) })
}

func TestTestableAferoChown(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	testErr(is, afs, fuseFs, func() fuse.Status { return fuseFs.Chown("sadf", 0, 0, nil) })
}

func TestTestableAferoUtimens(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	is.Equal(fuseFs.Utimens("sadf", nil, nil, nil), fuse.EINVAL)
	is.Equal(fuseFs.Utimens("sadf", nil, &time.Time{}, nil), fuse.EINVAL)
	is.Equal(fuseFs.Utimens("sadf", &time.Time{}, nil, nil), fuse.EINVAL)
	testErr(is, afs, fuseFs, func() fuse.Status { return fuseFs.Utimens("sadf", &time.Time{}, &time.Time{}, nil) })
}

func TestTestableAferoTruncate(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	file := &testableFile{}
	afs.file = file
	is.Equal(fuseFs.Truncate("sadf", 124, nil), fuse.OK)
	is.Equal(afs.calls[0], "OpenFile")
	is.Equal(file.trunc, int64(124))
	testErr(is, afs, fuseFs, func() fuse.Status { return fuseFs.Truncate("sadf", 124, nil) })
}

func TestTestableAferoMkdir(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	testErr(is, afs, fuseFs, func() fuse.Status { return fuseFs.Mkdir("sadf", 0, nil) })
}

func TestTestableAferoRename(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	testErr(is, afs, fuseFs, func() fuse.Status { return fuseFs.Rename("sadf", "asdf", nil) })
}

func TestTestableAferoRmdir(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	testErr(is, afs, fuseFs, func() fuse.Status { return fuseFs.Rmdir("sadf", nil) })
}

func TestTestableAferoUnlink(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	testErr(is, afs, fuseFs, func() fuse.Status { return fuseFs.Unlink("asdf", nil) })
}

func TestTestableAferoOpen(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	afs.err = os.ErrNotExist
	_, status := fuseFs.Open("", 0644, nil)
	is.Equal(status, fuse.ToStatus(afs.err))

	afs.err = nil
	afs.file = &testableFile{}
	file, status := fuseFs.Open("", 0644, nil)
	is.True(file != nil)
	is.Equal(status, fuse.OK)
}

func TestTestableAferoCreate(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	afs.err = os.ErrNotExist
	_, status := fuseFs.Create("", 0644, 0, nil)
	is.Equal(status, fuse.ToStatus(afs.err))

	afs.file = &testableFile{}
	afs.err = nil
	file, status := fuseFs.Create("", 0644, 123, nil)
	is.Equal(status, fuse.OK)
	is.True(file != nil)
	is.Equal(afs.openFileInt, 0644)
	is.Equal(afs.openFileMode, os.FileMode(123))
}

func TestTestableAferoSymlink(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	_, ok := (afero.NewMemMapFs()).(afero.Linker)
	is.Equal(ok, false)

	badFuseFs := NewFuseFileSystem(afero.NewMemMapFs())
	is.Equal(badFuseFs.Symlink("asd", "bcd", nil), fuse.ENOSYS)

	afs.err = os.ErrNotExist
	is.Equal(fuseFs.Symlink("asd", "bcd", nil), fuse.ToStatus(afs.err))

	is.Equal(afs.symlinkOld, "asd")
	is.Equal(afs.symlinkNew, "bcd")
}

func TestTestableAferoReadlink(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	_, ok := (afero.NewMemMapFs()).(afero.LinkReader)
	is.Equal(ok, false)

	badFuseFs := NewFuseFileSystem(afero.NewMemMapFs())
	_, status := badFuseFs.Readlink("asd", nil)
	is.Equal(status, fuse.ENOSYS)

	afs.link = "link"
	afs.err = os.ErrNotExist
	link, status := fuseFs.Readlink("asd", nil)
	is.Equal(status, fuse.ToStatus(afs.err))
	is.Equal(link, afs.link)
}

func TestTestableAferoOpenDir(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	memFs := afero.NewMemMapFs()
	is.NoErr(memFs.Mkdir("dir", 0755))
	is.NoErr(afero.WriteFile(memFs, "dir/ok.txt", []byte{}, 0755))
	is.NoErr(afero.WriteFile(memFs, "dir/ok2.txt", []byte{}, 0755))
	file, err := memFs.Open("dir")
	is.NoErr(err)
	afs.file = file

	afs.err = os.ErrNotExist
	_, status := fuseFs.OpenDir("dir", nil)
	is.Equal(status, fuse.ToStatus(afs.err))

	afs.err = nil
	stream, status := fuseFs.OpenDir("dir", nil)
	is.Equal(status, fuse.OK)
	is.Equal(len(stream), 2)
}

func TestTestableAferoStat(t *testing.T) {
	is, afs, fuseFs := fsBootstrap(t)

	afs.err = os.ErrExist
	_, status := fuseFs.GetAttr("", nil)
	is.Equal(status, fuse.ToStatus(afs.err))

	memFs := afero.NewMemMapFs()
	_, err := memFs.Create("asdf")
	is.NoErr(err)
	stat, err := memFs.Stat("asdf")
	is.NoErr(err)
	is.True(stat != nil)

	afs.err = nil
	afs.stat = stat
	attr, status := fuseFs.GetAttr("", nil)
	is.Equal(attr, fuse.ToAttr(stat))
	is.Equal(status, fuse.OK)

}
