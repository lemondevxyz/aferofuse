package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fuse/nodefs"
	"github.com/hanwen/go-fuse/v2/fuse/pathfs"
	"github.com/lemondevxyz/aferofuse"
	"github.com/spf13/afero"
)

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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		server.Unmount()
	}()

	server.Serve()
}
