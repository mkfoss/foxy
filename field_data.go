package foxy

import "github.com/mkfoss/foxy/pkg/core"

func (fld *Field) Value() (any, error) {

	switch fld.datatype {
	case core.DTCharacter:
		val, err := fld.dbf.Record.ReadString(fld.offset, fld.size, false, false)
		if err != nil {
			return nil, NewErrorf("could not read value for field %s", fld.name).SetWrapped(err).SetContext("read value")
		}
		return val, nil
	case core.DTInteger:
		val, err := fld.dbf.Record.ReadInteger(fld.offset)
		if err != nil {
			return nil, NewErrorf("could not read value for field %s", fld.name).SetWrapped(err).SetContext("read value")
		}
		return val, nil
	case core.DTNumeric, core.DTFloat:
		val, err := fld.dbf.Record.ReadNumeric(fld.offset, fld.size, fld.decimals)
		if err != nil {
			return nil, NewErrorf("could not read value for field %s", fld.name).SetWrapped(err).SetContext("read value")
		}
		return val, nil
	case core.DTDouble:
		val, err := fld.dbf.Record.ReadDouble(fld.offset)
		if err != nil {
			return nil, NewErrorf("could not read value for field %s", fld.name).SetWrapped(err).SetContext("read value")
		}
		return val, nil
	case core.DTCurrency:
		val, err := fld.dbf.Record.ReadCurrency(fld.offset)
		if err != nil {
			return nil, NewErrorf("could not read value for field %s", fld.name).SetWrapped(err).SetContext("read value")
		}
		return val, nil
	default:
		return nil, NewErrorf("unknown data type %d", fld.datatype)
	}
}

func (fld *Field) MustValue() any {

	val, err := fld.Value()
	if err != nil {
		panic(err)
	}
	return val
}
