package aferofuse

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
	"github.com/matryer/is"
)

type testableFile struct {
	fns   []string
	err   error
	n     int
	read  []byte
	write []byte
	trunc int64
	close bool
}

func (file *testableFile) Close() error {
	file.fns = append(file.fns, "close")
	if file.err != nil {
		return file.err
	}
	file.close = true
	return nil
}

// useless (in the context of this package) and unused file functions
func (file *testableFile) Read(p []byte) (n int, err error)           { return }
func (file *testableFile) Readdir(p int) (n []os.FileInfo, err error) { return }
func (file *testableFile) Readdirnames(p int) (n []string, err error) { return }
func (file *testableFile) Write(p []byte) (n int, err error)          { return }
func (file *testableFile) Seek(int64, int) (n int64, err error)       { return }
func (file *testableFile) Stat() (n os.FileInfo, err error)           { return }
func (file *testableFile) Sync() (err error)                          { return }
func (file *testableFile) WriteString(s string) (n int, err error)    { return }
func (file *testableFile) Name() (str string)                         { return }

func (file *testableFile) ReadAt(p []byte, off int64) (n int, err error) {
	file.fns = append(file.fns, fmt.Sprintf("readat %v %v", p, off))
	n, err = file.n, file.err
	if err == nil {
		n = copy(p, file.read)
	}
	return
}

func (file *testableFile) WriteAt(p []byte, off int64) (n int, err error) {
	file.fns = append(file.fns, fmt.Sprintf("writeat %v %v", p, off))
	n, err = file.n, file.err
	if err == nil {
		n = copy(file.write, p)
	}
	return
}

func (file *testableFile) Truncate(size int64) (err error) {
	file.fns = append(file.fns, fmt.Sprintf("truncate %d", size))
	err = file.err
	if err != nil {
		return
	}

	file.trunc = size

	return
}

func TestFileRead(t *testing.T) {
	is := is.New(t)
	file := &testableFile{}

	fuseFile := &aferoFuseFile{
		File: nodefs.NewDefaultFile(),
		file: file,
	}

	file.err = io.EOF

	res, status := fuseFile.Read([]byte{}, 0)
	is.Equal(res.Size(), file.n)
	is.Equal(fuse.OK, status)

	file.err = os.ErrNotExist
	file.n = 14
	res, status = fuseFile.Read([]byte{}, 0)
	is.True(res == nil)
	is.Equal(fuse.ToStatus(file.err), status)

	file.err = nil
	file.read = []byte{1, 2, 3, 4}
	aferoBytes := make([]byte, 64)
	res, status = fuseFile.Read(aferoBytes, 0)
	is.Equal(res.Size(), 64)
	fuseBytes, _ := res.Bytes(nil)
	is.Equal(aferoBytes[:4], fuseBytes[:4])
	is.Equal(fuse.ToStatus(file.err), status)
}

func TestFileWrite(t *testing.T) {
	is := is.New(t)
	file := &testableFile{}

	fuseFile := &aferoFuseFile{
		File: nodefs.NewDefaultFile(),
		file: file,
	}

	file.err = io.EOF

	n, status := fuseFile.Write([]byte{}, 0)
	is.Equal(n, uint32(file.n))
	is.Equal(fuse.OK, status)

	file.err = os.ErrNotExist
	file.n = 14
	n, status = fuseFile.Write([]byte{}, 0)
	is.Equal(fuse.ToStatus(file.err), status)

	file.err = nil
	file.write = make([]byte, 64)

	aferoBytes := make([]byte, 64)
	for i := range aferoBytes {
		aferoBytes[i] = byte(i)
	}

	n, status = fuseFile.Write(aferoBytes, 0)
	is.Equal(aferoBytes, file.write)
	is.Equal(fuse.OK, status)
}

func TestFileTruncate(t *testing.T) {
	is := is.New(t)
	file := &testableFile{}

	fuseFile := &aferoFuseFile{
		File: nodefs.NewDefaultFile(),
		file: file,
	}

	file.err = os.ErrNotExist
	is.Equal(fuse.ToStatus(file.err), fuseFile.Truncate(123))

	file.err = nil
	is.Equal(fuse.ToStatus(file.err), fuseFile.Truncate(123))
	is.Equal(file.trunc, int64(123))
}

func TestFileClose(t *testing.T) {
	is := is.New(t)
	file := &testableFile{}

	fuseFile := &aferoFuseFile{
		File: nodefs.NewDefaultFile(),
		file: file,
	}

	fuseFile.Release()
	is.Equal(file.close, true)
}

func TestEmptyTest(t *testing.T) {}
