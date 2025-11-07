package core

import (
	"io"
)

type Dbftype uint8

const (
	DbfUnknown Dbftype = iota
	DbfVisualFoxpro
)

func (dbft Dbftype) String() string {
	switch dbft {
	case DbfVisualFoxpro:
		return "Visual Foxpro"
	default:
		return "Unknown"
	}
}

func DbftypeFromMagicByte(mb uint8) Dbftype {
	switch mb {
	case 0x30:
		return DbfVisualFoxpro
	default:
		return DbfUnknown
	}
}

func ReadDbftype(rdr io.ReadSeeker) (Dbftype, error) {

	if _, err := rdr.Seek(0, io.SeekStart); err != nil {
		return DbfUnknown, err
	}

	mb := make([]byte, 1)
	if _, err := rdr.Read(mb); err != nil {
		return DbfUnknown, err
	}

	return DbftypeFromMagicByte(mb[0]), nil
}
