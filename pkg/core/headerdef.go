package core

import (
	"encoding/binary"
	"io"
)

type HeaderDef struct {
	LastupdateYear  uint8
	LastupdateMonth uint8
	LastupdateDay   uint8
	RecordCount     uint32
	RecordOffset    uint16
	RecordSize      uint16
	Reserved1       [16]byte
	TableFlags      uint8
	CodePage        uint8
	Reserved2       [2]byte
}

func ReadDbfHeaderDef(rdr io.ReadSeeker) (*HeaderDef, error) {

	var ret HeaderDef

	if _, err := rdr.Seek(1, io.SeekStart); err != nil {
		return nil, err
	}

	if err := binary.Read(rdr, binary.LittleEndian, &ret); err != nil {
		return nil, err
	}

	//todo: sanity check here, or in component?

	return &ret, nil
}
