package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/mkfoss/foxy/pkg/julian"
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

func (rec *Record) checkInRange(start, length int) error {
	if start+length > len(rec.data) {
		return fmt.Errorf("out of range")
	}
	return nil
}

func (rec *Record) ReadString(start, length int, trim, decode bool) (string, error) {
	if err := rec.checkInRange(start, length); err != nil {
		return "", err
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
	if err := rec.checkInRange(start, 8); err != nil {
		return 0, err
	}

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
	if err := rec.checkInRange(start, length); err != nil {
		return 0, err
	}
	trimmed := bytes.Trim(rec.data, " ")
	if len(trimmed) == 0 {
		return 0.0, nil
	}
	return strconv.ParseFloat(string(trimmed), 64)
}

func (rec *Record) ReadFloat(start, length, decimals int) (float64, error) {
	return rec.ReadNumeric(start, length, decimals)
}

func (rec *Record) ReadDate(start int) (time.Time, error) {
	if err := rec.checkInRange(start, 8); err != nil {
		return time.Time{}, err
	}
	return time.Parse("20060102", string(rec.data[start:start+8]))
}

func (rec *Record) ReadDateTime(start int) (time.Time, error) {
	if err := rec.checkInRange(start, 8); err != nil {
		return time.Time{}, err
	}
	var julDat uint32
	var mSec uint32
	if _, err := binary.Decode(rec.data[start:start+4], binary.LittleEndian, &julDat); err != nil {
		return time.Time{}, err
	}
	if _, err := binary.Decode(rec.data[start+4:start+8], binary.LittleEndian, &mSec); err != nil {
		return time.Time{}, err
	}
	// determine year, month, day
	y, m, d := julian.YMD(julDat)
	if y < 0 || y > 9999 {
		//todo: some dbf files seem to contain invalid dates, not sure if we want treat this an error until I know what is going on
		return time.Time{}, fmt.Errorf("time data invalid")
	}
	// calculate whole seconds and use the remainder as nanosecond resolution
	nSec := mSec / 1000
	mSec = mSec - (nSec * 1000)
	// create time using ymd and nanosecond timestamp
	return time.Date(y, time.Month(m), d, 0, 0, int(nSec), int(mSec)*int(time.Millisecond), time.UTC), nil
}
