package foxy

import (
	"io"
	"os"

	"github.com/mkfoss/foxy/pkg/core"
)

// Opener is an alias to core.Opener for backward compatibility
type Opener = core.Opener

// OsOpener implements both Opener and OpenerLister for the OS filesystem
type OsOpener struct{}

func (o *OsOpener) OpenFile(name string, flag int, perm os.FileMode) (io.ReadSeekCloser, error) {
	return os.OpenFile(name, flag, perm)
}

// ListFiles implements OpenerLister interface
func (o *OsOpener) ListFiles(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}
