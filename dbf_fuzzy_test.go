package foxy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDbf_FuzzyFileSearch(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "dbf_fuzzy_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Copy the logicalmem files with mixed case
	// Read original files
	dbfData, err := os.ReadFile(filepath.Join(TesrDataDir(t), "logicalmem.dbf"))
	if err != nil {
		t.Fatalf("failed to read logicalmem.dbf: %s", err)
	}
	fptData, err := os.ReadFile(filepath.Join(TesrDataDir(t), "logicalmem.fpt"))
	if err != nil {
		t.Fatalf("failed to read logicalmem.fpt: %s", err)
	}

	// Write files with mixed case names
	mixedCaseDbf := filepath.Join(tempDir, "TestFile.DBF")
	mixedCaseFpt := filepath.Join(tempDir, "testFile.Fpt")

	if err := os.WriteFile(mixedCaseDbf, dbfData, 0644); err != nil {
		t.Fatalf("failed to write mixed case dbf: %s", err)
	}
	if err := os.WriteFile(mixedCaseFpt, fptData, 0644); err != nil {
		t.Fatalf("failed to write mixed case fpt: %s", err)
	}

	// Test 1: Open with fuzzy search enabled (should work regardless of case in input)
	dbf := &Dbf{UseFuzzyFileSearch: true}
	baseFile := filepath.Join(tempDir, "testfile") // all lowercase
	err = dbf.Open(baseFile)
	if err != nil {
		t.Fatalf("unexpected error with fuzzy search: %s", err)
	}

	// Verify it opened correctly
	assert.True(t, dbf.Active())
	assert.Equal(t, 2, dbf.RecordCount())
	assert.Equal(t, 2, dbf.Fields.Count())

	// Read memo field to ensure FPT file is found
	memoFld := dbf.Field(1)
	val, err := memoFld.Value()
	if err != nil {
		t.Fatalf("unexpected error reading memo: %s", err)
	}
	assert.NotNil(t, val)

	err = dbf.Close()
	if err != nil {
		t.Fatalf("unexpected error closing: %s", err)
	}

	// Test 2: Try with different case in input
	dbf2 := &Dbf{UseFuzzyFileSearch: true}
	baseFile2 := filepath.Join(tempDir, "TESTFILE") // all uppercase
	err = dbf2.Open(baseFile2)
	if err != nil {
		t.Fatalf("unexpected error with fuzzy search uppercase: %s", err)
	}
	assert.True(t, dbf2.Active())
	err = dbf2.Close()
	if err != nil {
		t.Fatalf("unexpected error closing: %s", err)
	}

	// Test 3: Without fuzzy search (should fail with wrong case)
	dbf3 := &Dbf{UseFuzzyFileSearch: false}
	baseFile3 := filepath.Join(tempDir, "testfile")
	err = dbf3.Open(baseFile3 + ".dbf")
	assert.Error(t, err) // Should fail because file is actually TestFile.DBF
}

func TestDbf_FuzzyFileSearch_NotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dbf_fuzzy_notfound_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with non-existent file
	dbf := &Dbf{UseFuzzyFileSearch: true}
	err = dbf.Open(filepath.Join(tempDir, "nonexistent"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fuzzy search")
}

func TestDbf_FuzzyFileSearch_Multiple(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "dbf_fuzzy_multi_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %s", err)
	}
	defer os.RemoveAll(tempDir)

	// Create duplicate files with different cases
	dbfData, err := os.ReadFile(filepath.Join(TesrDataDir(t), "names.dbf"))
	if err != nil {
		t.Fatalf("failed to read names.dbf: %s", err)
	}

	// Write files with conflicting names
	if err := os.WriteFile(filepath.Join(tempDir, "test.dbf"), dbfData, 0644); err != nil {
		t.Fatalf("failed to write test.dbf: %s", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "Test.DBF"), dbfData, 0644); err != nil {
		t.Fatalf("failed to write Test.DBF: %s", err)
	}

	// Test with multiple matches (should fail)
	dbf := &Dbf{UseFuzzyFileSearch: true}
	err = dbf.Open(filepath.Join(tempDir, "test"))
	assert.Error(t, err)
	// Error should indicate fuzzy search failure
	assert.Contains(t, err.Error(), "fuzzy search")
}
