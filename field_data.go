package foxy

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mkfoss/foxy/pkg/core"
)

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
		data, _, err := fld.dbf.Record.ReadMemo(fld.offset, fld.dbf.fptBlockSize, fld.dbf.fpt)
		if err != nil {
			return nil, NewErrorf("could not read value for field %s", fld.name).SetWrapped(err).SetContext("read memo value")
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

func (fld *Field) AsString(trim, decode, sanitize bool) (string, error) {
	switch fld.datatype {
	case core.DTCharacter:
		val, err := fld.dbf.Record.ReadString(fld.offset, fld.size, trim, decode)
		if sanitize {
			rp := strings.NewReplacer("\t", "\\t", "\n", "\\n", "\r", "\\r")
			// todo: worth placing the above somewhere and not recreating every call
			val = rp.Replace(val)
		}
		if err != nil {
			return "", NewFieldError(fld, "call to read string failed").SetWrapped(err).SetContext("as string")
		}
		return val, nil
	case core.DTCurrency, core.DTDouble, core.DTFloat, core.DTNumeric:
		val, err := fld.Value()
		if err != nil {
			return "", NewFieldError(fld, "call to value failed").SetWrapped(err).SetContext("as string")
		}
		return fmt.Sprintf("%f", val), nil
	case core.DTDate:
		val, err := fld.dbf.Record.ReadDate(fld.offset)
		if err != nil {
			return "", NewFieldError(fld, "call to read date failed").SetWrapped(err).SetContext("as string")
		}
		return val.Format("2006-01-02"), nil
	case core.DTDateTime:
		val, err := fld.dbf.Record.ReadDateTime(fld.offset)
		if err != nil {
			return "", NewFieldError(fld, "call to read date time failed").SetWrapped(err).SetContext("as string")
		}
		return val.Format("2006-01-02T15:04:05"), nil
	case core.DTInteger:
		val, err := fld.Value()
		if err != nil {
			return "", NewFieldError(fld, "call to read int failed").SetWrapped(err).SetContext("as string")
		}
		return strconv.Itoa(val.(int)), nil
	case core.DTLogical:
		val, err := fld.dbf.Record.ReadLogical(fld.offset)
		if err != nil {
			return "", NewFieldError(fld, "call to read logical failed").SetWrapped(err).SetContext("as string")
		}
		if val {
			return "T", nil
		} else {
			return "F", nil
		}
	default:
		return "", NewFieldErrorf(fld, "unknown data type %d", fld.datatype).SetContext("as string")
	}
}

func (fld *Field) MustAsString(trim, decode, sanitize bool) string {
	str, err := fld.AsString(trim, decode, sanitize)
	if err != nil {
		panic(err)
	}
	return str
}
