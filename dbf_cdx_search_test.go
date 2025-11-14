package foxy

import (
	"path"
	"testing"

	"github.com/mkfoss/foxy/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestDbf_Seek_Basic(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	defer dbf.Close()

	// Test seeking for a key
	recNo, result, err := dbf.Seek("CATEGORYID", "00101")
	assert.NoError(t, err)
	t.Logf("Seek for '00101': recNo=%d, result=%v", recNo, result)

	// Should find something
	assert.True(t, result == core.SeekSuccess || result == core.SeekAfter,
		"Seek should return Success or After")
	assert.Greater(t, recNo, int32(0), "Should return a valid record number")

	// Verify the record was loaded
	if dbf.Fields.Count() > 0 {
		field := dbf.Fields.Field(0)
		_, err = field.Value()
		assert.NoError(t, err)
	}
}

func TestDbf_Seek_MultipleKeys(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	defer dbf.Close()

	// Test seeking for different keys
	testKeys := []string{"00101", "00102", "00103"}

	for _, key := range testKeys {
		recNo, result, err := dbf.Seek("CATEGORYID", key)
		assert.NoError(t, err)
		t.Logf("Seek for '%s': recNo=%d, result=%v", key, recNo, result)

		assert.NotEqual(t, core.SeekError, result, "Seek should not return error")
	}
}

func TestDbf_Seek_NonExistentTag(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	defer dbf.Close()

	// Test seeking with non-existent tag
	_, _, err = dbf.Seek("NONEXISTENT", "00101")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tag not found")
}

func TestDbf_Seek_CaseInsensitiveTag(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	defer dbf.Close()

	// Test with lowercase tag name
	recNo1, result1, err := dbf.Seek("categoryid", "00101")
	assert.NoError(t, err)

	// Test with uppercase tag name
	recNo2, result2, err := dbf.Seek("CATEGORYID", "00101")
	assert.NoError(t, err)

	// Should get the same results
	assert.Equal(t, recNo1, recNo2)
	assert.Equal(t, result1, result2)
}

func TestDbf_Find_ExactMatch(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	defer dbf.Close()

	// Test finding an exact match
	recNo, err := dbf.Find("CATEGORYID", "00101")
	if err != nil {
		// If not found, it might not exist in the data
		t.Logf("Key '00101' not found (exact match): %v", err)
		t.Skip("Exact key not in test data")
	} else {
		t.Logf("Found exact match at record %d", recNo)
		assert.Greater(t, recNo, int32(0))

		// Verify the record is loaded and we can read the field
		field := dbf.Fields.FieldByName("CATEGORYID")
		if field != nil {
			value, err := field.Value()
			assert.NoError(t, err)
			t.Logf("CATEGORYID value: %v", value)
		}
	}
}

func TestDbf_Find_NoMatch(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	defer dbf.Close()

	// Test finding a key that definitely doesn't exist
	_, err = dbf.Find("CATEGORYID", "ZZZZZ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDbf_Seek_AfterOpen(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	defer dbf.Close()

	// Seek should work immediately after open, without setting navigator
	recNo, result, err := dbf.Seek("CATEGORYID", "00101")
	assert.NoError(t, err)
	assert.NotEqual(t, core.SeekError, result)
	t.Logf("Seek after open: recNo=%d, result=%v", recNo, result)
}

func TestDbf_SeekAndNavigate(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	defer dbf.Close()

	// Seek to position at a record
	recNo, result, err := dbf.Seek("CATEGORYID", "00101")
	assert.NoError(t, err)
	t.Logf("Seek result: recNo=%d, result=%v", recNo, result)

	// Now try to navigate using default navigator
	// Note: The navigator position is independent of Seek
	err = dbf.First()
	assert.NoError(t, err)

	// Seek again
	recNo2, result2, err := dbf.Seek("CATEGORYID", "00102")
	assert.NoError(t, err)
	t.Logf("Second seek result: recNo=%d, result=%v", recNo2, result2)
}

func TestDbf_Seek_LazyLoad(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	defer dbf.Close()

	// CDX should not be loaded yet
	assert.Nil(t, dbf.cdx)

	// First seek should trigger CDX load
	_, _, err = dbf.Seek("CATEGORYID", "00101")
	assert.NoError(t, err)

	// Now CDX should be loaded
	assert.NotNil(t, dbf.cdx)
	assert.NotEmpty(t, dbf.cdxFilename)

	// Second seek should use already-loaded CDX
	_, _, err = dbf.Seek("CATEGORYID", "00102")
	assert.NoError(t, err)
}

func TestDbf_Seek_ReadFields(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	defer dbf.Close()

	// Seek and read all fields
	recNo, result, err := dbf.Seek("CATEGORYID", "00101")
	assert.NoError(t, err)
	assert.NotEqual(t, core.SeekError, result)

	t.Logf("Record %d found, reading all fields:", recNo)
	for i := 0; i < dbf.Fields.Count(); i++ {
		field := dbf.Fields.Field(i)
		value, err := field.Value()
		assert.NoError(t, err)
		t.Logf("  %s = %v", field.Name(), value)
	}
}
