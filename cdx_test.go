package foxy

import (
	"path"
	"testing"

	"github.com/mkfoss/foxy/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestCdxOpen(t *testing.T) {
	cdxPath := path.Join(TesrDataDir(t), "category.cdx")
	
	cdxFile, err := (&OsOpener{}).OpenFile(cdxPath, 0, 0600)
	assert.NoError(t, err)
	defer cdxFile.Close()

	cdx, err := core.OpenCdx(cdxFile)
	assert.NoError(t, err)
	assert.NotNil(t, cdx)

	assert.Greater(t, len(cdx.Tags), 0, "should have at least one tag")

	// Print tags for debugging
	t.Logf("Found %d tags:", len(cdx.Tags))
	for _, tag := range cdx.Tags {
		t.Logf("  Tag: %s, Expression: %s, KeyLen: %d, RootBlock: %d",
			tag.Name, tag.Expression, tag.KeyLen, tag.RootBlock)
	}
}

func TestCdxFindTag(t *testing.T) {
	cdxPath := path.Join(TesrDataDir(t), "category.cdx")
	
	cdxFile, err := (&OsOpener{}).OpenFile(cdxPath, 0, 0600)
	assert.NoError(t, err)
	defer cdxFile.Close()

	cdx, err := core.OpenCdx(cdxFile)
	assert.NoError(t, err)

	// Try to find a tag by name (case-insensitive)
	tag := cdx.FindTag("CATEGORYID")
	if tag != nil {
		t.Logf("Found tag CATEGORYID: %s", tag.Expression)
		assert.Equal(t, "CATEGORYID", tag.Name)
	} else {
		t.Log("CATEGORYID tag not found, checking what we have:")
		for _, tag := range cdx.Tags {
			t.Logf("  - %s: %s", tag.Name, tag.Expression)
		}
		t.Fatal("CATEGORYID tag should exist")
	}

	// Test case-insensitive search
	tagLower := cdx.FindTag("categoryid")
	assert.NotNil(t, tagLower, "should find tag with lowercase name")
	assert.Equal(t, tag.Name, tagLower.Name)
}

func TestCdxSeek(t *testing.T) {
	cdxPath := path.Join(TesrDataDir(t), "category.cdx")
	
	cdxFile, err := (&OsOpener{}).OpenFile(cdxPath, 0, 0600)
	assert.NoError(t, err)
	defer cdxFile.Close()

	cdx, err := core.OpenCdx(cdxFile)
	assert.NoError(t, err)

	// Find the CATEGORYID tag
	tag := cdx.FindTag("CATEGORYID")
	if tag == nil {
		t.Skip("CATEGORYID tag not found")
	}

	// Test seeking for a key
	recNo, result := tag.Seek(cdx, "00101")
	t.Logf("Seek for '00101': recNo=%d, result=%v", recNo, result)
	
	// We expect either Success or After (positioned at or after the key)
	assert.True(t, result == core.SeekSuccess || result == core.SeekAfter, 
		"Seek should return Success or After, got %v", result)
	assert.Greater(t, recNo, int32(0), "Should return a valid record number")
}

func TestCdxSeekMultipleKeys(t *testing.T) {
	cdxPath := path.Join(TesrDataDir(t), "category.cdx")
	
	cdxFile, err := (&OsOpener{}).OpenFile(cdxPath, 0, 0600)
	assert.NoError(t, err)
	defer cdxFile.Close()

	cdx, err := core.OpenCdx(cdxFile)
	assert.NoError(t, err)

	tag := cdx.FindTag("CATEGORYID")
	if tag == nil {
		t.Skip("CATEGORYID tag not found")
	}

	// Test seeking for different keys
	testKeys := []string{"00101", "00102", "00103"}
	
	for _, key := range testKeys {
		recNo, result := tag.Seek(cdx, key)
		t.Logf("Seek for '%s': recNo=%d, result=%v", key, recNo, result)
		
		if result == core.SeekError {
			t.Errorf("Seek for '%s' returned error", key)
		}
	}
}
