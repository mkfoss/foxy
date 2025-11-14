package foxy

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCdxNavigator_Basic(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	// Open DBF with CDX navigator
	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	
	// Set CDX navigator
	nav := NewCdxNavigator("CATEGORYID")
	err = dbf.SetNavigator(nav)
	assert.NoError(t, err)
	defer dbf.Close()

	// Verify the DBF is active
	assert.True(t, dbf.Active())

	// Test First
	err = dbf.First()
	assert.NoError(t, err)
	assert.Equal(t, 1, dbf.Position())
	assert.True(t, dbf.Navigator.IsFirst())

	// Test Next
	err = dbf.Next()
	assert.NoError(t, err)
	assert.Equal(t, 2, dbf.Position())
	assert.False(t, dbf.Navigator.IsFirst())

	// Test Last
	err = dbf.Last()
	assert.NoError(t, err)
	lastPos := dbf.Position()
	assert.Greater(t, lastPos, 2)
	assert.True(t, dbf.Navigator.IsLast())

	// Test Previous
	err = dbf.Previous()
	assert.NoError(t, err)
	assert.Equal(t, lastPos-1, dbf.Position())
	assert.False(t, dbf.Navigator.IsLast())

	// Test Goto
	err = dbf.Goto(5)
	assert.NoError(t, err)
	assert.Equal(t, 5, dbf.Position())

	// Verify we can read fields at this position
	if dbf.Fields.Count() > 0 {
		field := dbf.Fields.Field(0)
		_, err = field.Value()
		assert.NoError(t, err)
	}
}

func TestCdxNavigator_BoundaryConditions(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	
	nav := NewCdxNavigator("CATEGORYID")
	err = dbf.SetNavigator(nav)
	assert.NoError(t, err)
	defer dbf.Close()

	// Go to first
	err = dbf.First()
	assert.NoError(t, err)
	assert.True(t, dbf.Navigator.IsFirst())

	// Try Previous at BOF
	err = dbf.Previous()
	assert.Error(t, err)

	// Go to last
	err = dbf.Last()
	assert.NoError(t, err)
	assert.True(t, dbf.Navigator.IsLast())

	// Try Next at EOF
	err = dbf.Next()
	assert.Error(t, err)
}

func TestCdxNavigator_InvalidTag(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	defer dbf.Close()
	
	nav := NewCdxNavigator("NONEXISTENT")
	err = dbf.SetNavigator(nav)
	assert.Error(t, err)
}

func TestCdxNavigator_SequentialIteration(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	
	nav := NewCdxNavigator("CATEGORYID")
	err = dbf.SetNavigator(nav)
	assert.NoError(t, err)
	defer dbf.Close()

	// Iterate through all records
	err = dbf.First()
	assert.NoError(t, err)

	positions := []int{dbf.Position()}
	for {
		err = dbf.Next()
		if err != nil {
			break
		}
		positions = append(positions, dbf.Position())
	}

	// Verify positions are sequential
	for i := 0; i < len(positions); i++ {
		assert.Equal(t, i+1, positions[i])
	}

	t.Logf("Iterated through %d records", len(positions))

	// Verify we can go back
	err = dbf.Previous()
	assert.NoError(t, err)
	assert.Equal(t, len(positions)-1, dbf.Position())
}

func TestCdxNavigator_MultipleNavigators(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	// Test switching navigators
	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	defer dbf.Close()

	// Start with default navigator
	err = dbf.First()
	assert.NoError(t, err)
	defaultFirstPos := dbf.Position()

	// Switch to CDX navigator
	nav := NewCdxNavigator("CATEGORYID")
	err = dbf.SetNavigator(nav)
	assert.NoError(t, err)

	// First position in index order might be different
	cdxFirstPos := dbf.Position()
	t.Logf("Default first position: %d, CDX first position: %d", defaultFirstPos, cdxFirstPos)

	// Navigate a few records
	err = dbf.Next()
	assert.NoError(t, err)
	err = dbf.Next()
	assert.NoError(t, err)

	// Switch back to default navigator
	err = dbf.SetNavigator(&DefaulNavigator{})
	assert.NoError(t, err)

	// Should be at first position again with default navigator
	assert.Equal(t, 1, dbf.Position())
}

func TestCdxNavigator_ReadFieldsInIndexOrder(t *testing.T) {
	dbfPath := path.Join(TesrDataDir(t), "category.dbf")

	dbf := &Dbf{}
	err := dbf.Open(dbfPath)
	assert.NoError(t, err)
	
	nav := NewCdxNavigator("CATEGORYID")
	err = dbf.SetNavigator(nav)
	assert.NoError(t, err)
	defer dbf.Close()

	// Read first few records in index order
	err = dbf.First()
	assert.NoError(t, err)

	recordsRead := 0
	maxRecords := 5

	for recordsRead < maxRecords {
		// Try to read CATEGORYID field
		field := dbf.Fields.FieldByName("CATEGORYID")
		if field != nil {
			value, err := field.Value()
			assert.NoError(t, err)
			t.Logf("Position %d: CATEGORYID = %v", dbf.Position(), value)
		}

		err = dbf.Next()
		if err != nil {
			break
		}
		recordsRead++
	}

	assert.Greater(t, recordsRead, 0, "Should have read at least one record")
}
