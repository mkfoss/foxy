package core

import (
	"encoding/binary"
	"io"
)

type Record []byte

func NewRecord(size int) Record {
	return Record(make([]byte, size))
}

func (rec Record) Deleted() bool {
	return rec[0] == 0x2A
}

func (rec Record) Read(rdr io.Reader) error {
	return binary.Read(rdr, binary.LittleEndian, &rec)
}
