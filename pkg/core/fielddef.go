package core

import (
	"encoding/binary"
	"io"
)

type FieldDef struct {
	FieldName [11]byte
	FieldType byte
	Offset    uint32
	Length    uint8
	Decimals  uint8
	Flags     uint8
	Reserved  [13]byte
}

// ReadFieldDef reads in field definition data.  The func assumes that the rdr has been correctly positionod by the calling
// func
func ReadFieldDef(rdr io.Reader) (*FieldDef, error) {

	val := FieldDef{}
	if err := binary.Read(rdr, binary.LittleEndian, &val); err != nil {
		return nil, err
	}
	return &val, nil
}
