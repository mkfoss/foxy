package core

import "bytes"

type DataType byte

const (
	DTUnknown DataType = iota
	DTCharacter
	DTCurrency
	DTNumeric
	DTFloat
	DTDate
	DTDateTime
	DTDouble
	DTInteger
	DTLogical
	DTMemo
)

var supportedfieldtypes = []byte("CYNFDTBILM")

func (f DataType) String() string {

	if f < 1 || int(f) > len(supportedfieldtypes) {
		return "unknown"
	}
	return string(supportedfieldtypes[int(f)-1])
}

func DataTypeFromByte(ftchar byte) DataType {

	idx := bytes.IndexByte([]byte(supportedfieldtypes), bytes.ToUpper([]byte{ftchar})[0])
	if idx == -1 {
		return DTUnknown
	}
	return DataType(idx + 1)
}
