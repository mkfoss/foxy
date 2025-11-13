package foxy

import "github.com/mkfoss/foxy/pkg/core"

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

func (fld *Field) Name() string {
	return fld.name
}

func (fld *Field) DataType() core.DataType {
	return fld.datatype
}

func (fld *Field) Index() int {
	return fld.index
}

func (fld *Field) Size() int {
	return fld.size
}

func (fld *Field) Decimals() int {
	return fld.decimals
}

func (fld *Field) Binary() bool {
	return fld.binary
}

func (fld *Field) Nullable() bool {
	return fld.nullable
}

func (fld *Field) Offset() int {
	return fld.offset
}
