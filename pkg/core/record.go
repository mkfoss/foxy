package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

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

func (rec *Record) ReadCurrency(start int) (float64, error) {

	var i int64
	n, err := binary.Decode(rec.data[start:start+8], binary.LittleEndian, &i)
	if err != nil {
		return 0, err
	}
	if n != 8 {
		return 0, fmt.Errorf("invalid currency length: %d", n)
	}

	return float64(i) / 10000.00, nil
}

func (rec *Record) ReadNumeric(start, length, decimals int) (float64, error) {
	// N values are stored as string values, if no decimals return as int64, if decimals treat as float64
	trimmed := bytes.Trim(rec.data, " ")
	if len(trimmed) == 0 {
		return 0.0, nil
	}
	return strconv.ParseFloat(string(trimmed), 64)
}

func (rec *Record) ReadFloat(start, length, decimals int) (float64, error) {
	return rec.ReadNumeric(start, length, decimals)
}
