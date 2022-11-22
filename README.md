# aferofuse
[![GoDocs](https://godocs.io/github.com/lemondevxyz/aferofuse?status.svg)](https://godocs.io/github.com/lemondevxyz/aferofuse)
Have you ever thought: I really like X file system Go implementation but I sure wish I could expose it to the whole operating system. Well, you are in luck!

This is exactly what `aferofuse` is. Taken from the two libraries: afero and fuse - `aferofuse` aims to expose internal Go abstract filesystems to the linux operating system by using the `fuse` library.

## why?
My original reason to implement this was to faciliate a tool that could limit a directory's size. I searched all over the web for about 3.5 seconds, found nothing(or most likely didn't read through anything) and decided to implement this library.

## how?
Start by getting the library:
``` shell
go get github.com/lemondevxyz/aferofuse
```

Afterwards, just borrow the code from `example/osfs/main.go` and modify it to your liking! Too lazy to go there? Lucky for you I have a code snippet ready:
```go
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

```

## what's left?
- [x] Internal unit testing
- [ ] GitHub actions code coverage
- [x] ~~Testing through a stressor like `iozone`~~
- [ ] Release `v0.1.0`
