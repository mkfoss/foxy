package core

import (
	"bytes"
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
		{3, []byte{0x20, 0x01, 0x00, 0x00, 0x00, 0x87, 0x41, 0x6E, 0x6F, 0x6E, 0x79, 0x20, 0x4D, 0x6F, 0x75, 0x73, 0x8A, 0x20, 0x20, 0x20, 0x20, 0x20, 0x02, 0x00, 0x00, 0x00},
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
			got, err = rec.ReadString(tc.start, tc.length, tc.trim, tc.decode)
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
		start    int
		length   int
		dec      int
		expected float64
		err      string
	}

	testcases := []*testcase{
		{1, []byte{0x20, 0x31, 0x32, 0x33, 0x2E, 0x31, 0x35, 0x36, 0x20}, 1, 8, 3, 123.156, ""},
		{2, []byte{0x20, 0x20, 0x20, 0x20, 0x30, 0x2E, 0x30, 0x30, 0x30}, 1, 8, 3, 0, ""},
		{3, []byte{0x20, 0x2D, 0x31, 0x32, 0x33, 0x2E, 0x34, 0x35, 0x36}, 1, 8, 3, -123.456, ""},
		{4, []byte{0x20, 0x20, 0x31, 0x32, 0x33, 0x2E, 0x34, 0x35, 0x37}, 1, 8, 3, 123.457, ""},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("ReadNumeric %d", tc.id), func(t *testing.T) {
			rec := NewRecord(9, 0x03)
			err := rec.LoadData(bytes.NewReader(tc.input))
			if err != nil {
				t.Fatalf("unexpected load error: %s", err)
			}
			got, err := rec.ReadNumeric(1, tc.length, tc.dec)
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
		start    int
		expected time.Time
		err      string
	}

	testcases := []*testcase{
		{1, []byte(" 20250606"), 1, time.Date(2025, time.Month(6), 6, 0, 0, 0, 0, time.UTC), ""},
		{2, []byte(" 20251306"), 1, time.Time{}, "month out of range"},
		{3, []byte(" 20251232"), 1, time.Time{}, "day out of range"},
		{4, []byte("         "), 1, time.Time{}, "cannot parse"},
		{5, []byte("\x77ABC         "), 1, time.Time{}, "cannot parse"},
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
