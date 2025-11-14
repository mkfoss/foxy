package core

import (
	"io"
	"os"
)

// Opener is the base interface for opening files
type Opener interface {
	OpenFile(name string, flag int, perm os.FileMode) (io.ReadSeekCloser, error)
}

// OpenerLister extends Opener with the ability to list files in a directory
// This is used for fuzzy file search functionality
type OpenerLister interface {
	Opener
	// ListFiles returns a list of filenames in the given directory path
	ListFiles(dirPath string) ([]string, error)
}