package core

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// testOpenerLister is a simple test implementation of OpenerLister
type testOpenerLister struct{}

func (t *testOpenerLister) OpenFile(name string, flag int, perm os.FileMode) (io.ReadSeekCloser, error) {
	return os.OpenFile(name, flag, perm)
}

func (t *testOpenerLister) ListFiles(dirPath string) ([]string, error) {
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

func TestFuzzyFindFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fuzzyfile_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	lister := &testOpenerLister{}

	// Create test files with various case combinations
	testFiles := []string{
		"TestFile.dbf",
		"testFile.CDX",
		"TestFile.Fpt",
		"AnotherFile.DBF",
		"yetanother.cdx",
	}

	for _, filename := range testFiles {
		fullPath := filepath.Join(tempDir, filename)
		f, err := os.Create(fullPath)
		if err != nil {
			t.Fatalf("failed to create test file %s: %s", filename, err)
		}
		f.Close()
	}

	tests := []struct {
		name        string
		baseFile    string
		fileType    FileType
		expected    string
		expectError bool
	}{
		{
			name:     "find dbf with lowercase input",
			baseFile: filepath.Join(tempDir, "testfile"),
			fileType: FTDbf,
			expected: filepath.Join(tempDir, "TestFile.dbf"),
		},
		{
			name:     "find cdx with lowercase input",
			baseFile: filepath.Join(tempDir, "testfile"),
			fileType: FTCdx,
			expected: filepath.Join(tempDir, "testFile.CDX"),
		},
		{
			name:     "find fpt with uppercase input",
			baseFile: filepath.Join(tempDir, "TESTFILE"),
			fileType: FTFpt,
			expected: filepath.Join(tempDir, "TestFile.Fpt"),
		},
		{
			name:     "find dbf with mixed case input",
			baseFile: filepath.Join(tempDir, "AnotherFile"),
			fileType: FTDbf,
			expected: filepath.Join(tempDir, "AnotherFile.DBF"),
		},
		{
			name:     "find cdx with all lowercase",
			baseFile: filepath.Join(tempDir, "YETANOTHER"),
			fileType: FTCdx,
			expected: filepath.Join(tempDir, "yetanother.cdx"),
		},
		{
			name:        "no file found",
			baseFile:    filepath.Join(tempDir, "nonexistent"),
			fileType:    FTDbf,
			expectError: true,
		},
		{
			name:        "missing fpt file",
			baseFile:    filepath.Join(tempDir, "anotherfile"),
			fileType:    FTFpt,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FuzzyFindFile(tt.baseFile, tt.fileType, lister)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFuzzyFindFile_MultipleMatches(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "fuzzyfile_multi_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Create duplicate files (should cause error)
	testFiles := []string{
		"test.dbf",
		"Test.DBF",
	}

	for _, filename := range testFiles {
		fullPath := filepath.Join(tempDir, filename)
		f, err := os.Create(fullPath)
		if err != nil {
			t.Fatalf("failed to create test file %s: %s", filename, err)
		}
		f.Close()
	}

	// This should return an error because there are multiple matches
	lister := &testOpenerLister{}
	_, err = FuzzyFindFile(filepath.Join(tempDir, "test"), FTDbf, lister)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "multiple")
}

func TestFuzzyFindFile_WithExtensionInInput(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "fuzzyfile_ext_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file
	fullPath := filepath.Join(tempDir, "MyFile.DBF")
	f, err := os.Create(fullPath)
	if err != nil {
		t.Fatalf("failed to create test file: %s", err)
	}
	f.Close()

	// Test with extension in input (should be stripped)
	lister := &testOpenerLister{}
	result, err := FuzzyFindFile(filepath.Join(tempDir, "myfile.dbf"), FTDbf, lister)
	assert.NoError(t, err)
	assert.Equal(t, fullPath, result)
}
