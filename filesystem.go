package foxy

import (
	"io"
	"os"
)

type Walla func(name string) (io.ReadSeekCloser, error)

type Opener interface {
	OpenFile(name string, flag int, perm os.FileMode) (io.ReadSeekCloser, error)
}

type OsOpener struct{}

func (o *OsOpener) OpenFile(name string, flag int, perm os.FileMode) (io.ReadSeekCloser, error) {
	return os.OpenFile(name, flag, perm)
}
