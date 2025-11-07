package core

import (
	"fmt"
	"testing"

	"github.com/mkfoss/foxy/internal/mockfiler"
	"github.com/stretchr/testify/assert"
)

func Test_ReadDbftype(t *testing.T) {
	type testcase struct {
		id       int
		input    []byte
		expected Dbftype
	}

	testcases := []*testcase{
		{1, []byte{0x30}, DbfVisualFoxpro},
		{2, []byte{0x00}, DbfUnknown},
		{3, []byte{0x30, 0x46, 0xaa}, DbfVisualFoxpro},
		{4, []byte{0x12, 0x11, 0x65}, DbfUnknown},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("ReadDbftype %.2d", tc.id), func(t *testing.T) {

			fl := mockfiler.NewMockFiler()
			fl.Data = tc.input

			dt := DbftypeFromMagicByte(tc.input[0])
			assert.Equal(t, tc.expected, dt)

			dr, err := ReadDbftype(fl)
			if err != nil {
				t.Fatal("unexpected error: ", err)
			}
			assert.Equal(t, tc.expected, dr)
		})
	}
}
