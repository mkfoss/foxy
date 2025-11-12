package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"golang.org/x/text/encoding"
)

type Record struct {
	data     []byte
	codepage Codepage
	decoder  *encoding.Decoder
}

func NewRecord(size int, cp Codepage) *Record {
	return &Record{data: make([]byte, size), codepage: cp}
}

func (rec *Record) Deleted() bool {
	return rec.data[0] == 0x2A
}

func (rec *Record) LoadData(rdr io.Reader) error {
	return binary.Read(rdr, binary.LittleEndian, &rec.data)
}

func (rec *Record) ClearData() {
	rec.data = make([]byte, 0)
}

func (rec *Record) Size() int {
	return len(rec.data)
}

func (rec *Record) Data() []byte {
	return rec.data
}

func (rec *Record) ReadString(start, length int, trim, decode bool) (string, error) {
	if start+length >= len(rec.data) {
		return "", fmt.Errorf("out of range")
	}
	bts := rec.data[start : start+length]
	if trim {
		bts = bytes.TrimRight(bts, "\x20\x00")
	}
	if decode {
		if rec.codepage != 0x00 {
			if rec.decoder == nil {
				rec.decoder = CodepageDecoder(rec.codepage)
			}
			//if decoder is still nil, means there is no valid encoder
			if rec.decoder == nil {
				return "", fmt.Errorf("unsupported codepage 0x%X", byte(rec.codepage))
			}

			var err error
			bts, err = rec.decoder.Bytes(bts)
			if err != nil {
				return "", err
			}
		}
	}

	return string(bts), nil
}
