package foxy

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Navigation(t *testing.T) {
	dbf := &Dbf{}

	assert.ErrorContains(t, dbf.SetNavigator(&DefaulNavigator{}), "set navigator: dbf is inactive, could not perform operation")

	if err := dbf.Open(path.Join(DataDir(t), "basicnav.dbf")); err != nil {
		t.Fatal("Unexpected error:", err)
	}

	if dbf.Navigator == nil {
		_ = dbf.Close()
		t.Fatal("Navigator is nil")
	}

	assert.Equal(t, 10, dbf.RecordCount())
	assert.Equal(t, 1, dbf.Position())
	assert.True(t, dbf.IsFirst())
	assert.ErrorContains(t, dbf.Previous(), "seek previous: bof")

	for i := range 9 {
		err := dbf.Next()
		assert.NoError(t, err)
		assert.Equal(t, i+2, dbf.Position())
	}

	assert.True(t, dbf.IsLast())
	assert.ErrorContains(t, dbf.Next(), "seek next: eof")

	assert.NoError(t, dbf.First())
	assert.True(t, dbf.IsFirst())
	assert.Equal(t, 1, dbf.Position())
	assert.NoError(t, dbf.Last())
	assert.True(t, dbf.IsLast())
	assert.Equal(t, 10, dbf.Position())
	assert.NoError(t, dbf.Goto(5))
	assert.Equal(t, []byte{0x20, 0x05, 0x00, 0x00, 0x00}, dbf.Record.Data())
	assert.Equal(t, 5, dbf.Position())

	err := dbf.Close()

	assert.Nil(t, dbf.Navigator)

	if err != nil {
		t.Fatal("Unexpected error:", err)
	}
}
