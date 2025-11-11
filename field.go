package foxy

type Field struct {
	name     string
	datatype DataType
	index    int
	size     int
	decimals int
	binary   bool
	nullable bool
}

func (fld *Field) Name() string {
	return fld.name
}

func (fld *Field) DataType() DataType {
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
