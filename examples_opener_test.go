package foxy

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mkfoss/foxy/pkg/core" // imported for interface documentation
)

// Example: ZipOpener implements both core.Opener and core.OpenerLister for zip files
// This demonstrates how to create a custom opener that supports fuzzy file search
type ZipOpener struct {
	zipReader *zip.ReadCloser
}

func NewZipOpener(zipPath string) (*ZipOpener, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	return &ZipOpener{zipReader: r}, nil
}

func (z *ZipOpener) Close() error {
	if z.zipReader != nil {
		return z.zipReader.Close()
	}
	return nil
}

// OpenFile implements core.Opener interface
func (z *ZipOpener) OpenFile(name string, flag int, perm os.FileMode) (io.ReadSeekCloser, error) {
	// Search for the file in the zip archive
	for _, f := range z.zipReader.File {
		if f.Name == name || filepath.Join("/", f.Name) == name {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			// Wrap in a ReadSeekCloser (zip files don't support seeking, but we can work around it)
			return &zipFileWrapper{rc: rc}, nil
		}
	}
	return nil, fmt.Errorf("file %s not found in zip archive", name)
}

// ListFiles implements core.OpenerLister interface
func (z *ZipOpener) ListFiles(dirPath string) ([]string, error) {
	var files []string
	
	// Normalize the directory path
	dirPath = strings.TrimPrefix(dirPath, "/")
	if dirPath != "" && !strings.HasSuffix(dirPath, "/") {
		dirPath += "/"
	}
	
	// List all files in the specified directory within the zip
	for _, f := range z.zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		
		// Check if file is in the specified directory
		if dirPath == "" || strings.HasPrefix(f.Name, dirPath) {
			// Get just the filename without the directory path
			relativePath := strings.TrimPrefix(f.Name, dirPath)
			// Only include files directly in this directory (not subdirectories)
			if !strings.Contains(relativePath, "/") {
				files = append(files, relativePath)
			}
		}
	}
	
	return files, nil
}

// zipFileWrapper wraps a zip file ReadCloser to add minimal Seek support
type zipFileWrapper struct {
	rc io.ReadCloser
}

func (z *zipFileWrapper) Read(p []byte) (n int, err error) {
	return z.rc.Read(p)
}

func (z *zipFileWrapper) Seek(offset int64, whence int) (int64, error) {
	// Basic seek support - in real implementation you'd need to buffer or re-open
	return 0, fmt.Errorf("seek not fully supported on zip files")
}

func (z *zipFileWrapper) Close() error {
	return z.rc.Close()
}

// Example usage:
// func ExampleZipOpener() {
//     zipOpener, err := NewZipOpener("data.zip")
//     if err != nil {
//         panic(err)
//     }
//     defer zipOpener.Close()
//     
//     dbf := &Dbf{UseFuzzyFileSearch: true}
//     err = dbf.OpenWithOpener("TestFile", zipOpener)
//     if err != nil {
//         panic(err)
//     }
//     // DBF file will be found regardless of case: testfile.dbf, TestFile.DBF, etc.
// }
