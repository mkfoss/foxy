package foxy

import (
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mkfoss/foxy/pkg/core"
	"github.com/stretchr/testify/assert"
)

func Test_ReadValue(t *testing.T) {

	type testcase struct {
		id       int
		input    string
		expected string
		err      string
	}

	testcases := []testcase{
		{1, "names.dbf", "names.csv", ""},
		{2, "floats.dbf", "floats.csv", ""},
		{3, "dates.dbf", "dates.csv", ""},
		{4, "logicalmem.dbf", "logicalmem.csv", ""},
	}

	for _, tc := range testcases {
		t.Run(fmt.Sprintf("ReadValue #%.2d", tc.id), func(t *testing.T) {

			dbf := &Dbf{}
			err := dbf.Open(path.Join(TesrDataDir(t), tc.input))
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			defer func() { _ = dbf.Close() }()

			f, err := os.OpenFile(path.Join(TesrDataDir(t), tc.expected), os.O_RDONLY, 0)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			defer func() { _ = f.Close() }()

			csvf := csv.NewReader(f)

			crecs, err := csvf.ReadAll()
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			for _, crec := range crecs {
				for i, cfld := range crec {

					dfld := dbf.Field(i)
					switch dfld.datatype {
					case core.DTInteger:
						cval, err := strconv.Atoi(cfld)
						if err != nil {
							t.Fatalf("unexpected error: %s", err)
						}
						assert.Equal(t, int32(cval), dfld.MustValue())
						break
					case core.DTCharacter:
						assert.Equal(t, cfld, strings.Trim(dfld.MustValue().(string), " ")) //because .Value() returns untrimmed, undecoded string, we trim it here
						break
					case core.DTNumeric, core.DTFloat, core.DTDouble, core.DTCurrency:
						val, err := strconv.ParseFloat(cfld, 64)
						if err != nil {
							t.Fatalf("unexpected error: %s", err)
						}
						dval, err := dfld.Value()
						if err != nil {
							t.Fatalf("unexpected error: %s", err)
						}
						assert.Equal(t, val, dval)
						break
					case core.DTDate:
						val, err := time.Parse("20060102", cfld)
						if err != nil {
							t.Fatalf("unexpected error: %s", err)
						}
						assert.Equal(t, dfld.MustValue(), val)
					case core.DTDateTime:
						val, err := time.Parse("2006010215:04:05", cfld)
						if err != nil {
							t.Fatalf("unexpected error: %s", err)
						}
						assert.Equal(t, dfld.MustValue(), val)
					case core.DTLogical:
						var expected bool
						if cfld == "T" || cfld == "t" || cfld == "Y" || cfld == "y" {
							expected = true
						} else {
							expected = false
						}
						assert.Equal(t, expected, dfld.MustValue())
					case core.DTMemo:
						// Memo fields return the text content
						// Normalize line endings (CSV reader normalizes \r\n to \n)
						memoVal := dfld.MustValue().(string)
						memoVal = strings.ReplaceAll(memoVal, "\r\n", "\n")
						assert.Equal(t, cfld, memoVal)
					default:
						t.Fatalf("unexpected datatype: %d", dfld.datatype)
					}
				}
				if !dbf.IsLast() {
					err = dbf.Next()
					if err != nil {
						t.Fatalf("unexpected error: %s", err)
					}
				}
			}
		})
	}
}
