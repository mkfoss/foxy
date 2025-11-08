package core

import (
	"fmt"
	"testing"

	"github.com/mkfoss/foxy/internal/mockfiler"
	"github.com/stretchr/testify/assert"
)

func Test_HeaderDef(t *testing.T) {

	type testcase struct {
		id       int
		input    []byte
		expected *HeaderDef
		error    string
	}

	testcases := []testcase{
		{1,
			[]byte{0x30, 0x19, 0x0B, 0x08, 0x03, 0x00, 0x00, 0x00, 0x68, 0x01, 0x15, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x49, 0x44,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			&HeaderDef{
				LastupdateYear:  25,
				LastupdateMonth: 11,
				LastupdateDay:   8,
				RecordCount:     3,
				RecordOffset:    360,
				RecordSize:      21,
				Reserved1:       [16]byte{},
				TableFlags:      0,
				CodePage:        0,
				Reserved2:       [2]byte{},
			},
			"",
		},
		{2,
			[]byte{0x30, 0x19, 0x0B, 0x08, 0x03, 0x00, 0x00, 0x00},
			&HeaderDef{},
			"unexpected EOF",
		},
		{3,
			[]byte{0x30, 0x19, 0x0B, 0x08, 0x03, 0x00, 0x00, 0x00},
			&HeaderDef{},
			"unexpected EOF",
		},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("HeaderDef #%.2d", tc.id), func(t *testing.T) {
			fl := mockfiler.NewMockFiler()
			fl.Data = tc.input
			defer func() { _ = fl.Close() }()

			got, err := ReadDbfHeaderDef(fl)
			if tc.error != "" {
				if err == nil {
					t.Errorf("ReadDbfHeaderDef() expected error but got nil")
				}
				assert.ErrorContains(t, err, tc.error)
			} else {
				if err != nil {
					t.Fatal("unexpected error", err)
				}
				assert.Equal(t, tc.expected, got)
			}
		})
	}

}
