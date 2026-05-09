package foxy

import (
	"io"
	"os"

	"github.com/mkfoss/foxy/pkg/core"
)

// Opener is an alias to core.Opener for backward compatibility.
// It defines the interface for opening files, allowing for custom storage backends.
type Opener = core.Opener

// OsOpener implements both Opener and OpenerLister for the standard OS filesystem.
type OsOpener struct{}

// OpenFile opens a file from the local filesystem.
func (o *OsOpener) OpenFile(name string, flag int, perm os.FileMode) (io.ReadSeekCloser, error) {
	return os.OpenFile(name, flag, perm)
}

// ListFiles lists all files in the specified directory.
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
