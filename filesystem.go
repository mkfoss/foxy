package foxy

import (
	"io"
	"os"
)

type Filer interface {
	io.ReadSeeker
	io.Closer
}

type Opener interface {
	OpenFile(name string, flag int, perm os.FileMode) (Filer, error)
}

type OsOpener struct{}

func (o *OsOpener) OpenFile(name string, flag int, perm os.FileMode) (Filer, error) {
	return os.OpenFile(name, flag, perm)
}
