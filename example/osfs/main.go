package main

import (
	"flag"
	"log"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
	"github.com/hanwen/go-fuse/v2/fuse/pathfs"
	"github.com/lemondevxyz/aferofuse"
	"github.com/spf13/afero"
)

type embeddedINode struct {
	inode *fs.Inode
}

func (e *embeddedINode) EmbeddedInode() *fs.Inode {
	return e.inode
}

func main() {
	base := flag.String("base", "", "the base path to 'contain' the os filesystem")
	debug := flag.Bool("debug", false, "print debug data")
	flag.Parse()
	if len(flag.Args()) < 1 {
		log.Fatal("Usage:\n  osfs MOUNTPOINT")
	}

	afs := afero.NewBasePathFs(afero.NewOsFs(), *base)
	fusefs := aferofuse.NewFuseFileSystem(afs)

	opts := &nodefs.Options{}
	opts.Debug = *debug

	mynodefs := pathfs.NewPathNodeFs(fusefs, &pathfs.PathNodeFsOptions{
		ClientInodes: true,
	})
	server, _, err := nodefs.MountRoot(flag.Arg(0), mynodefs.Root(), opts)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}
	log.Println("Mounted!")
	server.Serve()
}
