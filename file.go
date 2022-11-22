package aferofuse

import (
	"errors"
	"io"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
	"github.com/spf13/afero"
)

type aferoFuseFile struct {
	nodefs.File
	file afero.File
}

func (f *aferoFuseFile) Read(dest []byte, off int64) (fuse.ReadResult, fuse.Status) {
	_, err := f.file.ReadAt(dest, off)
	if errors.Is(err, io.EOF) {
		err = nil
	}
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return fuse.ReadResultData(dest), fuse.ToStatus(nil)
}

func (f *aferoFuseFile) Write(dest []byte, off int64) (uint32, fuse.Status) {
	n, err := f.file.WriteAt(dest, off)
	if errors.Is(err, io.EOF) {
		err = nil
	}
	return uint32(n), fuse.ToStatus(err)
}

func (f *aferoFuseFile) Release() {
	f.file.Close()
}

func (f *aferoFuseFile) Truncate(size uint64) fuse.Status {
	return fuse.ToStatus(f.file.Truncate(int64(size)))
}
