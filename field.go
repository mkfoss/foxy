package foxy

import "github.com/mkfoss/foxy/pkg/core"

// Field represents a single field (column) definition in a DBF file.
type Field struct {
	dbf      *Dbf
	name     string
	datatype core.DataType
	index    int
	offset   int
	size     int
	decimals int
	binary   bool
	nullable bool
}

// Name returns the name of the field (lowercase).
func (fld *Field) Name() string {
	return fld.name
}

// DataType returns the core data type of the field.
func (fld *Field) DataType() core.DataType {
	return fld.datatype
}

// Index returns the 0-based index of the field in the table.
func (fld *Field) Index() int {
	return fld.index
}

// Size returns the length of the field in bytes.
func (fld *Field) Size() int {
	return fld.size
}

// Decimals returns the number of decimal places for numeric/floating fields.
func (fld *Field) Decimals() int {
	return fld.decimals
}

// Binary returns true if the field contains binary data.
func (fld *Field) Binary() bool {
	return fld.binary
}

// Nullable returns true if the field can contain null values.
func (fld *Field) Nullable() bool {
	return fld.nullable
}

// Offset returns the byte offset of the field within a record.
func (fld *Field) Offset() int {
	return fld.offset
}
