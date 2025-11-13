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
			return nil, NewErrorf("could not read value for field %s", fld.name).SetWrapped(err).SetContext("read int value")
		}
		return val, nil
	case core.DTNumeric, core.DTFloat:
		val, err := fld.dbf.Record.ReadNumeric(fld.offset, fld.size, fld.decimals)
		if err != nil {
			return nil, NewErrorf("could not read value for field %s", fld.name).SetWrapped(err).SetContext("read ascii float like value")
		}
		return val, nil
	case core.DTDouble:
		val, err := fld.dbf.Record.ReadDouble(fld.offset)
		if err != nil {
			return nil, NewErrorf("could not read value for field %s", fld.name).SetWrapped(err).SetContext("read float value")
		}
		return val, nil
	case core.DTCurrency:
		val, err := fld.dbf.Record.ReadCurrency(fld.offset)
		if err != nil {
			return nil, NewErrorf("could not read value for field %s", fld.name).SetWrapped(err).SetContext("read currency value")
		}
		return val, nil
	case core.DTDate:
		val, err := fld.dbf.Record.ReadDate(fld.offset)
		if err != nil {
			return nil, NewErrorf("could not read value for field %s", fld.name).SetWrapped(err).SetContext("read date value")
		}
		return val, nil
	case core.DTDateTime:
		val, err := fld.dbf.Record.ReadDateTime(fld.offset)
		if err != nil {
			return nil, NewErrorf("could not read value for field %s", fld.name).SetWrapped(err).SetContext("read datetimevalue")
		}
		return val, nil
	case core.DTLogical:
		val, err := fld.dbf.Record.ReadLogical(fld.offset)
		if err != nil {
			return nil, NewErrorf("could not read value for field %s", fld.name).SetWrapped(err).SetContext("read logical value")
		}
		return val, nil
	case core.DTMemo:
		// Lazy load FPT file if needed
		if err := fld.dbf.ensureFptLoaded(); err != nil {
			return nil, NewErrorf("could not load FPT file for field %s", fld.name).SetWrapped(err).SetContext("read memo value")
		}
		data, isText, err := fld.dbf.Record.ReadMemo(fld.offset, fld.dbf.fptBlockSize, fld.dbf.fpt)
		if err != nil {
			return nil, NewErrorf("could not read value for field %s", fld.name).SetWrapped(err).SetContext("read memo value")
		}
		// Return as string if text, otherwise return as bytes
		if isText {
			return string(data), nil
		}
		return data, nil
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
