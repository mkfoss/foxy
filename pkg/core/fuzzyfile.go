package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileType represents the type of file to search for
type FileType int

const (
	FTDbf FileType = iota
	FTCdx
	FTFpt
)

// String returns the string representation of the FileType
func (ft FileType) String() string {
	switch ft {
	case FTDbf:
		return "dbf"
	case FTCdx:
		return "cdx"
	case FTFpt:
		return "fpt"
	default:
		return "unknown"
	}
}

// FuzzyFindFile searches for a file with a case-insensitive match using an OpenerLister
// baseFilename can be just a filename or include a path like /path/to/filename
// fileType indicates what type of file to look for (dbf, cdx, or fpt)
// lister is an OpenerLister that provides the ListFiles capability
// Returns the actual filename found, or an error if none or multiple matches found
func FuzzyFindFile(baseFilename string, fileType FileType, lister OpenerLister) (string, error) {
	// Split into directory and base name
	dir := filepath.Dir(baseFilename)
	if dir == "." {
		// No directory specified, use current working directory
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %w", err)
		}
	}

	// Get the base filename without directory and without extension
	base := filepath.Base(baseFilename)
	// Strip any existing extension
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}

	// Determine the extension we're looking for
	var targetExt string
	switch fileType {
	case FTDbf:
		targetExt = ".dbf"
	case FTCdx:
		targetExt = ".cdx"
	case FTFpt:
		targetExt = ".fpt"
	default:
		return "", fmt.Errorf("unknown file type: %d", fileType)
	}

	// Get list of files from the lister
	filenames, err := lister.ListFiles(dir)
	if err != nil {
		return "", fmt.Errorf("failed to list files in directory %s: %w", dir, err)
	}

	// Search for matching files (case-insensitive)
	var matches []string
	baseLower := strings.ToLower(base)
	targetExtLower := strings.ToLower(targetExt)

	for _, filename := range filenames {
		filenameLower := strings.ToLower(filename)

		// Check if filename matches base + target extension (case-insensitive)
		expectedName := baseLower + targetExtLower
		if filenameLower == expectedName {
			matches = append(matches, filepath.Join(dir, filename))
		}
	}

	// Check results
	if len(matches) == 0 {
		return "", fmt.Errorf("no %s file found matching %s in directory %s", fileType.String(), base, dir)
	}

	if len(matches) > 1 {
		return "", fmt.Errorf("multiple %s files found matching %s in directory %s: %v", fileType.String(), base, dir, matches)
	}

	return matches[0], nil
}
