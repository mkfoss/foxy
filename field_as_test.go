package foxy

import (
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestField_AsString(t *testing.T) {
	// Test with names.dbf for character fields
	dbf := &Dbf{}
	err := dbf.Open(path.Join(TesrDataDir(t), "names.dbf"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer func() { _ = dbf.Close() }()

	// Test character field
	fld := dbf.Field(1) // name field
	val, err := fld.AsString(true, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	assert.Contains(t, val, "Anony Mouse")

	// Test integer field
	fld2 := dbf.Field(0) // id field
	val2, err := fld2.AsString(false, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	assert.Contains(t, val2, "1")

	// Test with floats.dbf for numeric fields
	dbf2 := &Dbf{}
	err = dbf2.Open(path.Join(TesrDataDir(t), "floats.dbf"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer func() { _ = dbf2.Close() }()

	// Test numeric field
	fld3 := dbf2.Field(0) // numeric field
	val3, err := fld3.AsString(false, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	assert.Contains(t, val3, "123")
}

func TestField_AsFloat(t *testing.T) {
	dbf := &Dbf{}
	err := dbf.Open(path.Join(TesrDataDir(t), "floats.dbf"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer func() { _ = dbf.Close() }()

	tests := []struct {
		name     string
		field    int
		expected float64
	}{
		{"numeric field", 0, 123.0},
		{"float field", 1, 123.0},
		{"double field", 2, 123.0},
		{"currency field", 3, 123.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fld := dbf.Field(tt.field)
			val, err := fld.AsFloat()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			assert.InDelta(t, tt.expected, val, 0.0001)
		})
	}

	// Test with names.dbf for integer to float conversion
	dbf2 := &Dbf{}
	err = dbf2.Open(path.Join(TesrDataDir(t), "names.dbf"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer func() { _ = dbf2.Close() }()

	fld := dbf2.Field(0) // integer field
	val, err := fld.AsFloat()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	assert.InDelta(t, 1.0, val, 0.0001)
}

func TestField_AsInteger(t *testing.T) {
	dbf := &Dbf{}
	err := dbf.Open(path.Join(TesrDataDir(t), "names.dbf"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer func() { _ = dbf.Close() }()

	tests := []struct {
		name     string
		field    int
		expected int
	}{
		{"integer field", 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fld := dbf.Field(tt.field)
			val, err := fld.AsInteger()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			assert.Equal(t, tt.expected, val)
		})
	}

	// Test float to integer conversion
	dbf2 := &Dbf{}
	err = dbf2.Open(path.Join(TesrDataDir(t), "floats.dbf"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer func() { _ = dbf2.Close() }()

	fld := dbf2.Field(2) // numeric field with 123.456
	val, err := fld.AsInteger()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	assert.Equal(t, 123, val)
}

func TestField_AsTime(t *testing.T) {
	dbf := &Dbf{}
	err := dbf.Open(path.Join(TesrDataDir(t), "dates.dbf"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer func() { _ = dbf.Close() }()

	tests := []struct {
		name     string
		field    int
		expected time.Time
	}{
		{"date field", 0, time.Date(2025, 11, 10, 0, 0, 0, 0, time.UTC)},
		{"datetime field", 1, time.Date(2025, 12, 11, 11, 12, 13, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fld := dbf.Field(tt.field)
			val, err := fld.AsTime()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			assert.Equal(t, tt.expected.Year(), val.Year())
			assert.Equal(t, tt.expected.Month(), val.Month())
			assert.Equal(t, tt.expected.Day(), val.Day())
		})
	}
}

func TestField_AsLogical(t *testing.T) {
	dbf := &Dbf{}
	err := dbf.Open(path.Join(TesrDataDir(t), "logicalmem.dbf"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer func() { _ = dbf.Close() }()

	tests := []struct {
		name     string
		expected bool
	}{
		{"true logical", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fld := dbf.Field(0) // logical field
			val, err := fld.AsLogical()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			assert.Equal(t, tt.expected, val)
		})
	}

	// Test second record (false)
	err = dbf.Next()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	fld := dbf.Field(0)
	val, err := fld.AsLogical()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	assert.Equal(t, false, val)
}

func TestField_MustAsString(t *testing.T) {
	dbf := &Dbf{}
	err := dbf.Open(path.Join(TesrDataDir(t), "names.dbf"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer func() { _ = dbf.Close() }()

	fld := dbf.Field(1) // name field
	val := fld.MustAsString(true, false, false)
	assert.Contains(t, val, "Anony Mouse")
}

func TestField_MustAsFloat(t *testing.T) {
	dbf := &Dbf{}
	err := dbf.Open(path.Join(TesrDataDir(t), "floats.dbf"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer func() { _ = dbf.Close() }()

	fld := dbf.Field(2) // double field
	val := fld.MustAsFloat()
	assert.InDelta(t, 123.0, val, 0.0001)
}

func TestField_MustAsInteger(t *testing.T) {
	dbf := &Dbf{}
	err := dbf.Open(path.Join(TesrDataDir(t), "names.dbf"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer func() { _ = dbf.Close() }()

	fld := dbf.Field(0) // id field
	val := fld.MustAsInteger()
	assert.Equal(t, 1, val)
}

func TestField_MustAsTime(t *testing.T) {
	dbf := &Dbf{}
	err := dbf.Open(path.Join(TesrDataDir(t), "dates.dbf"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer func() { _ = dbf.Close() }()

	fld := dbf.Field(0) // date field
	val := fld.MustAsTime()
	assert.Equal(t, 2025, val.Year())
	assert.Equal(t, time.November, val.Month())
	assert.Equal(t, 10, val.Day())
}

func TestField_MustAsLogical(t *testing.T) {
	dbf := &Dbf{}
	err := dbf.Open(path.Join(TesrDataDir(t), "logicalmem.dbf"))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	defer func() { _ = dbf.Close() }()

	fld := dbf.Field(0) // logical field
	val := fld.MustAsLogical()
	assert.Equal(t, true, val)
}
