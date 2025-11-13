package foxy

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"

	"github.com/mkfoss/foxy/pkg/core"
)

type Fields struct {
	fields   []*Field
	fieldmap map[string]int
}

// Read reads the fields from the reader
func (flds *Fields) Read(dbf *Dbf) error {

	_, err := dbf.fl.Seek(32, io.SeekStart)
	if err != nil {
		return NewError("failed to seek start of field definitions").SetWrapped(err).SetContext("dbf fields read")
	}
	expectedflds := (dbf.recordoffset - 296) / 32
	//todo: sanity check
	fields := make([]*Field, expectedflds)
	fmap := make(map[string]int)
	for i := range expectedflds {
		def, err := core.ReadFieldDef(dbf.fl)
		if err != nil {
			return NewErrorf("failed to read field definition %d", i).SetWrapped(err).SetContext("dbf fields read")
		}
		fld := &Field{}
		fld.dbf = dbf
		fld.index = i
		fld.datatype = core.DataTypeFromByte(def.FieldType)
		fld.name = strings.ToLower(string(bytes.TrimRight(def.FieldName[:], " \x00\t\r\n")))
		fld.offset = int(def.Offset)
		fld.size = int(def.Length)
		fld.decimals = int(def.Decimals)
		fld.nullable = def.Flags&0x02 == 0x02
		fld.binary = def.Flags&0x04 == 0x04
		fields[i] = fld
		fmap[fld.name] = i
	}

	var b byte
	err = binary.Read(dbf.fl, binary.LittleEndian, &b)
	if err != nil {
		return NewErrorf("failed to read end of fields marker").SetWrapped(err).SetContext("dbf fields read")
	}
	if b != 0x0D {
		return NewErrorf("invalid end of fields marker").SetContext("dbf fields read")
	}

	flds.fields = fields
	flds.fieldmap = fmap

	return nil
}

func (flds *Fields) Count() int {

	return len(flds.fields)
}

func (flds *Fields) Field(index int) *Field {

	if index < 0 && index >= len(flds.fields) {
		return nil
	}
	return flds.fields[index]
}

func (flds *Fields) FieldByName(name string) *Field {

	idx, ok := flds.fieldmap[strings.ToLower(name)]
	if !ok {
		return nil
	}
	return flds.Field(idx)
}
