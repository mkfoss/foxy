package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/mkfoss/foxy/pkg/julian"
	"golang.org/x/text/encoding"
)

type Record struct {
	data             []byte
	codepage         Codepage
	decoder          *encoding.Decoder
	sanitizereplacer *strings.Replacer
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

func (rec *Record) ReadString(start, length int, trim, decode, sanitize bool) (string, error) {
	if err := rec.checkInRange(start, length); err != nil {
		return "", err
	}
	bts := rec.data[start : start+length]

	str, err := rec.ProcessStringBytes(bts, trim, decode, sanitize)
	if err != nil {
		return "", err
	}
	return str, nil
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

func (rec *Record) ReadNumeric(start, length int) (float64, error) {
	if err := rec.checkInRange(start, length); err != nil {
		return 0, err
	}
	trimmed := bytes.Trim(rec.data[start:start+length], " ")
	if len(trimmed) == 0 {
		return 0.0, nil
	}
	return strconv.ParseFloat(string(trimmed), 64)
}

func (rec *Record) ReadFloat(start, length, decimals int) (float64, error) {
	return rec.ReadNumeric(start, length)
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
	mSec -= (nSec * 1000)
	// create time using ymd and nanosecond timestamp
	return time.Date(y, time.Month(m), d, 0, 0, int(nSec), int(mSec)*int(time.Millisecond), time.UTC), nil
}

func (rec *Record) ReadInteger(start int) (int32, error) {
	if err := rec.checkInRange(start, 4); err != nil {
		return 0, err
	}

	var i int32
	n, err := binary.Decode(rec.data[start:start+4], binary.LittleEndian, &i)
	if err != nil {
		return 0, err
	}
	if n != 4 {
		return 0, fmt.Errorf("invalid integer length: %d", n)
	}

	return i, nil
}

func (rec *Record) ReadDouble(start int) (float64, error) {
	if err := rec.checkInRange(start, 8); err != nil {
		return 0, err
	}

	var d float64
	n, err := binary.Decode(rec.data[start:start+8], binary.LittleEndian, &d)
	if err != nil {
		return 0, err
	}
	if n != 8 {
		return 0, fmt.Errorf("invalid double length: %d", n)
	}

	return d, nil
}

func (rec *Record) ReadLogical(start int) (bool, error) {
	if err := rec.checkInRange(start, 1); err != nil {
		return false, err
	}

	b := rec.data[start]
	switch b {
	case 'T', 't', 'Y', 'y':
		return true, nil
	case 'F', 'f', 'N', 'n', '?', ' ':
		return false, nil
	default:
		return false, fmt.Errorf("invalid logical value: 0x%X", b)
	}
}

func (rec *Record) ReadMemo(start int, blockSize uint16, fpt io.ReadSeeker) ([]byte, bool, error) {
	if err := rec.checkInRange(start, 4); err != nil {
		return nil, false, err
	}

	// Read the block number (little-endian in the DBF field)
	block := binary.LittleEndian.Uint32(rec.data[start : start+4])

	// Block 0 means empty memo
	if block == 0 {
		return nil, false, nil
	}

	// Seek to the block position in the FPT file
	// Position = blocknumber * blocksize
	if _, err := fpt.Seek(int64(blockSize)*int64(block), io.SeekStart); err != nil {
		return nil, false, fmt.Errorf("failed to seek in FPT file: %w", err)
	}

	// Read the memo block header (8 bytes)
	// First 4 bytes: signature (big-endian) - 1 = text, 0 = binary
	// Next 4 bytes: length (big-endian)
	hbuf := make([]byte, 8)
	if _, err := io.ReadFull(fpt, hbuf); err != nil {
		return nil, false, fmt.Errorf("failed to read FPT block header: %w", err)
	}

	sign := binary.BigEndian.Uint32(hbuf[:4])
	leng := binary.BigEndian.Uint32(hbuf[4:])

	if leng == 0 {
		// No data according to block header
		return []byte{}, sign == 1, nil
	}

	// Read the actual memo data
	buf := make([]byte, leng)
	if _, err := io.ReadFull(fpt, buf); err != nil {
		return buf, false, fmt.Errorf("failed to read FPT block data: %w", err)
	}

	return buf, sign == 1, nil
}

func (rec *Record) ProcessStringBytes(strbytes []byte, trim, decode, sanitize bool) (string, error) {
	if trim {
		strbytes = bytes.Trim(strbytes, "\x20\x00")
	}
	if decode {
		if rec.codepage != 0x00 {
			if rec.decoder == nil {
				rec.decoder = CodepageDecoder(rec.codepage)
			}
			// if decoder is still nil, means there is no valid encoder
			if rec.decoder == nil {
				return "", fmt.Errorf("unsupported codepage 0x%X", byte(rec.codepage))
			}

			var err error
			strbytes, err = rec.decoder.Bytes(strbytes)
			if err != nil {
				return "", err
			}
		}
	}
	str := string(strbytes)
	if sanitize {
		if rec.sanitizereplacer == nil {
			rec.sanitizereplacer = strings.NewReplacer("\t", "\\t", "\n", "\\n", "\r", "\\r")
		}
		str = rec.sanitizereplacer.Replace(str)
	}
	return str, nil
}
