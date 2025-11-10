package foxy

import (
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/mkfoss/foxy/pkg/core"
	"github.com/stretchr/testify/assert"
)

func Test_DbfOpenHeader(t *testing.T) {

	type testcase struct {
		id         int
		inputfile  string
		inputbytes []byte
		expected   map[string]any
		err        string
	}

	testcases := []testcase{
		{1, "names.dbf", nil, map[string]any{
			"count": 3, "lastup": "2025-11-08", "recoff": 360, "recsz": 21, "cp": core.Codepage(0x00), "idx": false, "fpt": false,
		}, ""},
		{2, "names.dbf", []byte{0x30, 0x00}, map[string]any{
			"count": 0, "lastup": time.Time{}, "recoff": 0, "recsz": 0, "cp": core.Codepage(0x00), "idx": false, "fpt": false,
		}, "open with opener: read header failed"},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("DbfOpenHeader #%.2d", tc.id), func(t *testing.T) {

			dbf := &Dbf{}
			var err error
			if tc.inputbytes != nil {
				err = dbf.OpenWithOpener(tc.inputfile, &BytesOpener{
					data: tc.inputbytes,
				})
			} else {
				err = dbf.Open(path.Join(DataDir(t), tc.inputfile))
			}
			if tc.err != "" {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				assert.ErrorContains(t, err, tc.err)
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				lu, err := time.Parse("2006-01-02", tc.expected["lastup"].(string))
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				assert.True(t, dbf.Active())
				assert.NotNil(t, dbf.opener)
				assert.NotNil(t, dbf.fl)
				assert.Equal(t, tc.expected["count"], dbf.RecordCount())
				assert.Equal(t, lu, dbf.LastUpdated())
				assert.Equal(t, tc.expected["recoff"], dbf.RecordOffset())
				assert.Equal(t, tc.expected["recsz"], dbf.RecordSize())
				assert.Equal(t, tc.expected["cp"], dbf.CodePage())
				assert.Equal(t, tc.expected["idx"], dbf.HasIndex())
				assert.Equal(t, tc.expected["fpt"], dbf.HasFpt())

				err = dbf.Close()
				assert.NoError(t, err)

				assert.False(t, dbf.Active())
				assert.Equal(t, 0, dbf.RecordCount())
				assert.True(t, dbf.lastupdated.IsZero())
				assert.Equal(t, 0, dbf.RecordOffset())
				assert.Equal(t, 0, dbf.RecordSize())
				assert.Equal(t, core.Codepage(0), dbf.CodePage())
				assert.Equal(t, false, dbf.HasIndex())
				assert.Equal(t, false, dbf.HasFpt())
				assert.Nil(t, dbf.opener)
				assert.Nil(t, dbf.fl)
				assert.Equal(t, "", dbf.Filename())
			}
		})
	}
}
