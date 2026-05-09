package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_RecordReadstring(t *testing.T) {
	type testcase struct {
		id       int
		input    []byte
		cp       Codepage
		start    int
		length   int
		trim     bool
		decode   bool
		expected string
		err      string
	}

	testcases := []*testcase{
		{1, []byte{0x20, 0x01, 0x00, 0x00, 0x00, 0x41, 0x6E, 0x6F, 0x6E, 0x79, 0x20, 0x4D, 0x6F, 0x75, 0x73, 0x65, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x02, 0x00, 0x00, 0x00},
			0x00, 5, 16, false, false, "Anony Mouse     ", ""},
		{2, []byte{0x20, 0x01, 0x00, 0x00, 0x00, 0x41, 0x6E, 0x6F, 0x6E, 0x79, 0x20, 0x4D, 0x6F, 0x75, 0x73, 0x65, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x02, 0x00, 0x00, 0x00},
			0x00, 5, 16, true, false, "Anony Mouse", ""},
		{3, []byte{0x20, 0x01, 0x00, 0x00, 0x00, 0x87, 0x41, 0x6E, 0x6F, 0x6E, 0x79, 0x20, 0x4D, 0x6F, 0x75, 0x73, 0x8A, 0x20, 0x20, 0x20, 0x20, 0x20, 0x02, 0x00, 0x00, 0x00},
			0x03, 5, 16, true, true, "‡Anony MousŠ", ""},
		{4, []byte{},
			0x00, 5, 16, true, true, "", "out of range"},
		{5, []byte{0x20, 0x01, 0x00, 0x00, 0x00, 0x87, 0x41, 0x6E, 0x6F, 0x6E, 0x79, 0x20, 0x4D, 0x6F, 0x75, 0x73, 0x8A, 0x20, 0x20, 0x20, 0x20, 0x20, 0x02, 0x00, 0x00, 0x00},
			0x9A, 5, 16, true, true, "", "unsupported codepage 0x9A"},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("%.2d", tc.id), func(t *testing.T) {
			rec := NewRecord(len(tc.input), tc.cp)
			err := rec.LoadData(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatalf("unexpected load error: %s", err)
			}
			var got string
			got, err = rec.ReadString(tc.start, tc.length, tc.trim, tc.decode, false)
			if tc.err != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.ErrorContains(t, err, tc.err)
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				assert.Equal(t, tc.expected, got)
			}
		})
	}
}

func Test_RecordReadCurrency(t *testing.T) {
	type testcase struct {
		id       int
		input    []byte
		expected float64
		err      string
	}

	testcases := []*testcase{
		{1, []byte{0x20, 0x80, 0xD6, 0x12, 0x00, 0x00, 0x00, 0x00, 0x00}, 123.456, ""},
		{2, []byte{0x20, 0x04, 0x77, 0xDA, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}, -245.99, ""},
		{3, []byte{0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0, ""},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("ReadCurrency %.2d", tc.id), func(t *testing.T) {
			rec := NewRecord(9, 0x03)
			err := rec.LoadData(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatalf("unexpected load error: %s", err)
			}
			got, err := rec.ReadCurrency(1)
			if tc.err != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.ErrorContains(t, err, tc.err)
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				assert.Equal(t, tc.expected, got)
			}
		})
	}
}

func TestRecord_ReadNumeric(t *testing.T) {
	type testcase struct {
		id       int
		input    []byte
		expected float64
		err      string
	}

	testcases := []*testcase{
		{1, []byte{0x20, 0x31, 0x32, 0x33, 0x2E, 0x31, 0x35, 0x36, 0x20}, 123.156, ""},
		{2, []byte{0x20, 0x20, 0x20, 0x20, 0x30, 0x2E, 0x30, 0x30, 0x30}, 0, ""},
		{3, []byte{0x20, 0x2D, 0x31, 0x32, 0x33, 0x2E, 0x34, 0x35, 0x36}, -123.456, ""},
		{4, []byte{0x20, 0x20, 0x31, 0x32, 0x33, 0x2E, 0x34, 0x35, 0x37}, 123.457, ""},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("ReadNumeric %d", tc.id), func(t *testing.T) {
			rec := NewRecord(9, 0x03)
			err := rec.LoadData(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatalf("unexpected load error: %s", err)
			}
			got, err := rec.ReadNumeric(1, 8)
			if tc.err != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.ErrorContains(t, err, tc.err)
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				assert.Equal(t, tc.expected, got)
			}
		})
	}
}

func TestRecord_ReadDate(t *testing.T) {
	type testcase struct {
		id       int
		input    []byte
		expected time.Time
		err      string
	}

	testcases := []*testcase{
		{1, []byte(" 20250606"), time.Date(2025, time.Month(6), 6, 0, 0, 0, 0, time.UTC), ""},
		{2, []byte(" 20251306"), time.Time{}, "month out of range"},
		{3, []byte(" 20251232"), time.Time{}, "day out of range"},
		{4, []byte("         "), time.Time{}, "cannot parse"},
		{5, []byte("\x77ABC         "), time.Time{}, "cannot parse"},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("ReadDate %d", tc.id), func(t *testing.T) {
			rec := NewRecord(9, 0x03)
			err := rec.LoadData(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatalf("unexpected load error: %s", err)
			}
			got, err := rec.ReadDate(1)
			if tc.err != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.ErrorContains(t, err, tc.err)
				assert.True(t, got.IsZero())
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				assert.Equal(t, tc.expected, got)
			}
		})
	}
}

func TestRecord_ReadDateTime(t *testing.T) {
	type testcase struct {
		id       int
		input    []byte
		expected time.Time
		err      string
	}

	testcases := []*testcase{
		{1, []byte{0x20, 0x3E, 0x8D, 0x25, 0x00, 0x00, 0x00, 0x00, 0x00},
			time.Date(2025, time.Month(11), 10, 0, 0, 0, 0, time.UTC), ""},
		{2, []byte{0x20, 0x77, 0x07, 0x25, 0x00, 0x20, 0x08, 0x4D, 0x03},
			time.Date(1932, time.Month(2), 5, 15, 23, 0, 0, time.UTC), ""},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("ReadDateTime %d", tc.id), func(t *testing.T) {
			rec := NewRecord(9, 0x03)
			err := rec.LoadData(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatalf("unexpected load error: %s", err)
			}
			got, err := rec.ReadDateTime(1)
			if tc.err != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.ErrorContains(t, err, tc.err)
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				assert.Equal(t, tc.expected, got)
			}
		})
	}
}

func TestRecord_ReadInteger(t *testing.T) {
	type testcase struct {
		id       int
		input    []byte
		expected int32
		err      string
	}

	testcases := []*testcase{
		{1, []byte{0x20, 0x01, 0x00, 0x00, 0x00}, 1, ""},
		{2, []byte{0x20, 0xFF, 0xFF, 0xFF, 0xFF}, -1, ""},
		{3, []byte{0x20, 0x00, 0x00, 0x00, 0x00}, 0, ""},
		{4, []byte{0x20, 0x39, 0x30, 0x00, 0x00}, 12345, ""},
		{5, []byte{0x20, 0xC7, 0xCF, 0xFF, 0xFF}, -12345, ""},
		{6, []byte{0x20}, 0, "out of range"},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("ReadInteger %d", tc.id), func(t *testing.T) {
			rec := NewRecord(len(tc.input), 0x03)
			err := rec.LoadData(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatalf("unexpected load error: %s", err)
			}
			got, err := rec.ReadInteger(1)
			if tc.err != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.ErrorContains(t, err, tc.err)
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				assert.Equal(t, tc.expected, got)
			}
		})
	}
}

func TestRecord_ReadDouble(t *testing.T) {
	type testcase struct {
		id       int
		input    []byte
		expected float64
		err      string
	}

	testcases := []*testcase{
		{1, []byte{0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0.0, ""},
		{2, []byte{0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0x3F}, 1.0, ""},
		{3, []byte{0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xF0, 0xBF}, -1.0, ""},
		{4, []byte{0x20, 0x77, 0xBE, 0x9F, 0x1A, 0x2F, 0xDD, 0x5E, 0x40}, 123.456, ""},
		{5, []byte{0x20, 0x77, 0xBE, 0x9F, 0x1A, 0x2F, 0xDD, 0x5E, 0xC0}, -123.456, ""},
		{6, []byte{0x20}, 0, "out of range"},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("ReadDouble %d", tc.id), func(t *testing.T) {
			rec := NewRecord(len(tc.input), 0x03)
			err := rec.LoadData(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatalf("unexpected load error: %s", err)
			}
			got, err := rec.ReadDouble(1)
			if tc.err != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.ErrorContains(t, err, tc.err)
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				assert.Equal(t, tc.expected, got)
			}
		})
	}
}

func TestRecord_ReadLogical(t *testing.T) {
	type testcase struct {
		id       int
		input    []byte
		expected bool
		err      string
	}

	testcases := []*testcase{
		{1, []byte{0x20, 0x54}, true, ""},
		{2, []byte{0x20, 0x74}, true, ""},
		{3, []byte{0x20, 0x59}, true, ""},
		{4, []byte{0x20, 0x79}, true, ""},
		{5, []byte{0x20, 0x46}, false, ""},
		{6, []byte{0x20, 0x66}, false, ""},
		{7, []byte{0x20, 0x4E}, false, ""},
		{8, []byte{0x20, 0x6E}, false, ""},
		{9, []byte{0x20, 0x3F}, false, ""},
		{10, []byte{0x20, 0x20}, false, ""},
		{11, []byte{0x20, 0x58}, false, "invalid logical value"},
		{12, []byte{0x20}, false, "out of range"},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("ReadLogical %d", tc.id), func(t *testing.T) {
			rec := NewRecord(len(tc.input), 0x03)
			err := rec.LoadData(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatalf("unexpected load error: %s", err)
			}
			got, err := rec.ReadLogical(1)
			if tc.err != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.ErrorContains(t, err, tc.err)
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				assert.Equal(t, tc.expected, got)
			}
		})
	}
}

func TestRecord_ReadMemo(t *testing.T) {
	type testcase struct {
		id           int
		input        []byte
		blockSize    uint16
		fptData      []byte
		expectedData []byte
		expectedText bool
		err          string
	}

	// Helper to create FPT block data
	createFptBlock := func(blockNum uint32, signature uint32, data []byte) []byte {
		blockSize := uint16(64) // Standard block size

		// Calculate required size: enough blocks to hold header + data
		neededBytes := int(blockNum)*int(blockSize) + 8 + len(data)
		numBlocks := (neededBytes + int(blockSize) - 1) / int(blockSize)
		fptData := make([]byte, numBlocks*int(blockSize))

		// Write block header at position blockNum * blockSize
		offset := int(blockNum) * int(blockSize)
		binary.BigEndian.PutUint32(fptData[offset:offset+4], signature)
		binary.BigEndian.PutUint32(fptData[offset+4:offset+8], uint32(len(data)))

		// Write data after header
		copy(fptData[offset+8:], data)
		return fptData
	}

	testcases := []*testcase{
		// Empty memo (block 0)
		{1, []byte{0x20, 0x00, 0x00, 0x00, 0x00}, 64, []byte{}, nil, false, ""},

		// Text memo at block 1
		{2, []byte{0x20, 0x01, 0x00, 0x00, 0x00}, 64,
			createFptBlock(1, 1, []byte("Hello, World!")),
			[]byte("Hello, World!"), true, ""},

		// Binary memo at block 2
		{3, []byte{0x20, 0x02, 0x00, 0x00, 0x00}, 64,
			createFptBlock(2, 0, []byte{0x01, 0x02, 0x03, 0x04}),
			[]byte{0x01, 0x02, 0x03, 0x04}, false, ""},

		// Text memo with larger content at block 5
		{4, []byte{0x20, 0x05, 0x00, 0x00, 0x00}, 64,
			createFptBlock(5, 1, []byte("This is a longer memo field with more text content that spans multiple lines.\nLine 2\nLine 3")),
			[]byte("This is a longer memo field with more text content that spans multiple lines.\nLine 2\nLine 3"), true, ""},

		// Empty data but non-zero block
		{5, []byte{0x20, 0x03, 0x00, 0x00, 0x00}, 64,
			createFptBlock(3, 1, []byte{}),
			[]byte{}, true, ""},

		// Out of range read
		{6, []byte{0x20}, 64, []byte{}, nil, false, "out of range"},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("ReadMemo %d", tc.id), func(t *testing.T) {
			rec := NewRecord(len(tc.input), 0x03)
			err := rec.LoadData(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatalf("unexpected load error: %s", err)
			}

			gotData, gotText, err := rec.ReadMemo(1, tc.blockSize, bytes.NewReader(tc.fptData))
			if tc.err != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				assert.ErrorContains(t, err, tc.err)
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				assert.Equal(t, tc.expectedData, gotData)
				assert.Equal(t, tc.expectedText, gotText)
			}
		})
	}
}
