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
		{3, "fourfields.dbf", nil, map[string]any{
			"count": 2, "lastup": "2025-11-11", "recoff": 424, "recsz": 33, "cp": core.Codepage(0x03), "idx": false, "fpt": false,
		}, ""},
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

func Test_Fields(t *testing.T) {

	type testcase struct {
		id         int
		inputfile  string
		inputbytes []byte
		expected   []*Field
		err        string
	}

	testcases := []testcase{
		{1, "names.dbf", nil, []*Field{
			{nil, "id", DTInteger, 0, 4, 0, true, false},
			{nil, "name", DTCharacter, 1, 16, 0, false, false},
		}, ""},
		{2, "fourfields.dbf", nil, []*Field{
			{nil, "int", DTInteger, 0, 4, 0, true, false},
			{nil, "char", DTCharacter, 1, 10, 0, false, false},
			{nil, "num", DTNumeric, 2, 10, 4, false, false},
			{nil, "float", DTCurrency, 3, 8, 4, true, false},
		}, ""},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("Fields #%.2d", tc.id), func(t *testing.T) {
			dbf := &Dbf{}
			assert.Nil(t, dbf.Record)
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

				if dbf.Fields == nil {
					t.Fatal("dbf.Fields should not be nil")
				}

				if dbf.Record == nil {
					t.Fatal("dbf.Record should not be nil")
				}

				assert.Equal(t, len(dbf.Record.data), dbf.RecordSize())

				assert.Equal(t, len(tc.expected), dbf.Fields.Count())
				for i, fld := range tc.expected {
					fld.dbf = dbf
					assert.Equal(t, fld, dbf.Fields.Field(i))
					assert.Equal(t, fld, dbf.Fields.FieldByName(fld.Name()))
				}

				err := dbf.Close()
				if err != nil {
					t.Fatalf("unexpected error closing dbf: %s", err)
				}

				assert.Nil(t, dbf.Fields)
				assert.Nil(t, dbf.Record)
			}
		})
	}

}
