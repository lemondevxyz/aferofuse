package aferofuse

import (
	"os"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
	"github.com/hanwen/go-fuse/v2/fuse/pathfs"
	"github.com/spf13/afero"
)

type aferoFuseFs struct {
	pathfs.FileSystem
	afs afero.Fs
}

func (a *aferoFuseFs) String() string {
	return a.afs.Name()
}

func (a *aferoFuseFs) GetAttr(name string, context *fuse.Context) (*fuse.Attr, fuse.Status) {
	stat, err := a.afs.Stat(name)
	return fuse.ToAttr(stat), fuse.ToStatus(err)
}

func (a *aferoFuseFs) Chmod(name string, mode uint32, context *fuse.Context) fuse.Status {
	err := a.afs.Chmod(name, os.FileMode(mode))
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fuse.OK
}

func (a *aferoFuseFs) Chown(name string, uid uint32, gid uint32, context *fuse.Context) fuse.Status {
	err := a.afs.Chown(name, int(uid), int(gid))
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fuse.OK
}

func (a *aferoFuseFs) Utimens(name string, atime *time.Time, mtime *time.Time, context *fuse.Context) fuse.Status {
	if atime == nil || mtime == nil {
		return fuse.EINVAL
	}

	err := a.afs.Chtimes(name, *atime, *mtime)
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fuse.OK
}

func (a *aferoFuseFs) Truncate(name string, size uint64, context *fuse.Context) fuse.Status {
	file, err := a.afs.OpenFile(name, os.O_WRONLY, 0)
	defer file.Close()
	if err != nil {
		return fuse.ToStatus(err)
	}

	return fuse.ToStatus(file.Truncate(int64(size)))
}

func (a *aferoFuseFs) Mkdir(name string, mode uint32, context *fuse.Context) fuse.Status {
	return fuse.ToStatus(a.afs.Mkdir(name, os.FileMode(mode)))
}

func (a *aferoFuseFs) Rename(oldName, newName string, context *fuse.Context) fuse.Status {
	return fuse.ToStatus(a.afs.Rename(oldName, newName))
}

func (a *aferoFuseFs) Rmdir(name string, context *fuse.Context) fuse.Status {
	return fuse.ToStatus(a.afs.Remove(name))
}

func (a *aferoFuseFs) Unlink(name string, context *fuse.Context) fuse.Status {
	return a.Rmdir(name, context)
}

func (a *aferoFuseFs) Open(name string, flags uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	file, err := a.afs.OpenFile(name, int(flags), 0755)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return &aferoFuseFile{
		File: nodefs.NewDefaultFile(),
		file: file,
	}, fuse.ToStatus(nil)
}

func (a *aferoFuseFs) Create(name string, flags uint32, mode uint32, context *fuse.Context) (nodefs.File, fuse.Status) {
	file, err := a.afs.OpenFile(name, int(flags), os.FileMode(mode))
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	return &aferoFuseFile{
		File: nodefs.NewDefaultFile(),
		file: file,
	}, fuse.ToStatus(nil)
}

func (a *aferoFuseFs) OpenDir(name string, context *fuse.Context) ([]fuse.DirEntry, fuse.Status) {
	files, err := afero.ReadDir(a.afs, name)
	if err != nil {
		return nil, fuse.ToStatus(err)
	}

	arr := make([]fuse.DirEntry, 0, len(files))
	for _, v := range files {
		arr = append(arr, fuse.DirEntry{
			Name: v.Name(),
			Mode: uint32(v.Mode()),
		})
	}

	return arr, fuse.ToStatus(nil)
}

func (a *aferoFuseFs) Symlink(oldname string, newname string, context *fuse.Context) fuse.Status {
	linker, ok := a.afs.(afero.Linker)
	if !ok {
		return fuse.ENOSYS
	}

	return fuse.ToStatus(linker.SymlinkIfPossible(oldname, newname))
}

func (a *aferoFuseFs) Readlink(name string, context *fuse.Context) (string, fuse.Status) {
	linker, ok := a.afs.(afero.LinkReader)
	if !ok {
		return "", fuse.ENOSYS
	}

	str, err := linker.ReadlinkIfPossible(name)
	return str, fuse.ToStatus(err)
}

func NewFuseFileSystem(afs afero.Fs) pathfs.FileSystem {
	return pathfs.NewLockingFileSystem(&aferoFuseFs{
		FileSystem: pathfs.NewDefaultFileSystem(),
		afs:        afs,
	})
}
