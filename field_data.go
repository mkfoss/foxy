package foxy

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mkfoss/foxy/pkg/core"
)

// Value returns the raw value of the field as an interface{}.
// The underlying type depends on the field's data type.
func (fld *Field) Value() (any, error) {

	switch fld.datatype {
	case core.DTCharacter:
		val, err := fld.dbf.Record.ReadString(fld.offset, fld.size, false, false, false)
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
		val, err := fld.dbf.Record.ReadNumeric(fld.offset, fld.size)
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

// MustValue returns the raw value of the field and panics if an error occurs.
func (fld *Field) MustValue() any {

	val, err := fld.Value()
	if err != nil {
		panic(err)
	}
	return val
}

// AsString returns the field value as a string.
// trim: remove trailing spaces.
// decode: decode character encoding.
// sanitize: remove non-printable characters.
func (fld *Field) AsString(trim, decode, sanitize bool) (string, error) {
	switch fld.datatype {
	case core.DTCharacter:
		val, err := fld.dbf.Record.ReadString(fld.offset, fld.size, trim, decode, sanitize)
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
		return strconv.Itoa(int(val.(int32))), nil
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
	case core.DTMemo:
		val, isbinary, err := fld.dbf.ReadMemo(fld.offset, fld.dbf.fptBlockSize, fld.dbf.fpt)
		if err != nil {
			return "", NewFieldError(fld, "call to read memo failed").SetWrapped(err).SetContext("as string")
		}
		if isbinary {
			return "", NewFieldError(fld, "cannot read binary memo as string").SetWrapped(err).SetContext("as string")
		}
		strval, err := fld.dbf.Record.ProcessStringBytes(val, trim, decode, sanitize)
		if err != nil {
			return "", NewFieldError(fld, "call to read process string bytes failed").SetWrapped(err).SetContext("as string")
		}
		return strval, nil
	default:
		return "", NewFieldErrorf(fld, "unsupported data type %d", fld.datatype).SetContext("as string")
	}
}

// MustAsString returns the field value as a string and panics if an error occurs.
func (fld *Field) MustAsString(trim, decode, sanitize bool) string {
	str, err := fld.AsString(trim, decode, sanitize)
	if err != nil {
		panic(err)
	}
	return str
}

// AsFloat returns the field value as a float64.
// Supports Numeric, Float, Currency, Double, Integer, and Character (if numeric) fields.
func (fld *Field) AsFloat() (float64, error) {
	switch fld.datatype {
	case core.DTCurrency, core.DTDouble, core.DTFloat, core.DTNumeric:
		val, err := fld.Value()
		if err != nil {
			return 0, NewFieldError(fld, "call to value failed").SetWrapped(err).SetContext("as float")
		}
		return val.(float64), nil
	case core.DTInteger:
		val, err := fld.Value()
		if err != nil {
			return 0, NewFieldError(fld, "call to value failed").SetWrapped(err).SetContext("as float")
		}
		return float64(val.(int32)), nil
	case core.DTCharacter:
		val, err := fld.dbf.Record.ReadString(fld.offset, fld.size, true, false, false)
		if err != nil {
			return 0, NewFieldError(fld, "call to read string failed").SetWrapped(err).SetContext("as float")
		}
		floatVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, NewFieldError(fld, "cannot parse string as float").SetWrapped(err).SetContext("as float")
		}
		return floatVal, nil
	default:
		return 0, NewFieldErrorf(fld, "unsupported data type %d", fld.datatype).SetContext("as float")
	}
}

// MustAsFloat returns the field value as a float64 and panics if an error occurs.
func (fld *Field) MustAsFloat() float64 {
	val, err := fld.AsFloat()
	if err != nil {
		panic(err)
	}
	return val
}

// AsTime returns the field value as a time.Time.
// Supports Date, DateTime, and Character (if formatted) fields.
func (fld *Field) AsTime() (time.Time, error) {
	switch fld.datatype {
	case core.DTDate:
		val, err := fld.dbf.Record.ReadDate(fld.offset)
		if err != nil {
			return time.Time{}, NewFieldError(fld, "call to read date failed").SetWrapped(err).SetContext("as time")
		}
		return val, nil
	case core.DTDateTime:
		val, err := fld.dbf.Record.ReadDateTime(fld.offset)
		if err != nil {
			return time.Time{}, NewFieldError(fld, "call to read date time failed").SetWrapped(err).SetContext("as time")
		}
		return val, nil
	case core.DTCharacter:
		val, err := fld.dbf.Record.ReadString(fld.offset, fld.size, true, false, false)
		if err != nil {
			return time.Time{}, NewFieldError(fld, "call to read string failed").SetWrapped(err).SetContext("as time")
		}
		// Try date format first
		timeVal, err := time.Parse("2006-01-02", val)
		if err != nil {
			// Try datetime format
			timeVal, err = time.Parse("2006-01-02T15:04:05", val)
			if err != nil {
				return time.Time{}, NewFieldError(fld, "cannot parse string as time").SetWrapped(err).SetContext("as time")
			}
		}
		return timeVal, nil
	default:
		return time.Time{}, NewFieldErrorf(fld, "unsupported data type %d", fld.datatype).SetContext("as time")
	}
}

// MustAsTime returns the field value as a time.Time and panics if an error occurs.
func (fld *Field) MustAsTime() time.Time {
	val, err := fld.AsTime()
	if err != nil {
		panic(err)
	}
	return val
}

// AsInteger returns the field value as an int.
// Supports Integer, Numeric, Float, Currency, Double, and Character (if numeric) fields.
func (fld *Field) AsInteger() (int, error) {
	switch fld.datatype {
	case core.DTInteger:
		val, err := fld.Value()
		if err != nil {
			return 0, NewFieldError(fld, "call to value failed").SetWrapped(err).SetContext("as integer")
		}
		return int(val.(int32)), nil
	case core.DTCurrency, core.DTDouble, core.DTFloat, core.DTNumeric:
		val, err := fld.Value()
		if err != nil {
			return 0, NewFieldError(fld, "call to value failed").SetWrapped(err).SetContext("as integer")
		}
		return int(val.(float64)), nil
	case core.DTCharacter:
		val, err := fld.dbf.Record.ReadString(fld.offset, fld.size, true, false, false)
		if err != nil {
			return 0, NewFieldError(fld, "call to read string failed").SetWrapped(err).SetContext("as integer")
		}
		intVal, err := strconv.Atoi(val)
		if err != nil {
			return 0, NewFieldError(fld, "cannot parse string as integer").SetWrapped(err).SetContext("as integer")
		}
		return intVal, nil
	default:
		return 0, NewFieldErrorf(fld, "unsupported data type %d", fld.datatype).SetContext("as integer")
	}
}

// MustAsInteger returns the field value as an int and panics if an error occurs.
func (fld *Field) MustAsInteger() int {
	val, err := fld.AsInteger()
	if err != nil {
		panic(err)
	}
	return val
}

// AsLogical returns the field value as a bool.
// Supports Logical and Character (T/F/Y/N) fields.
func (fld *Field) AsLogical() (bool, error) {
	switch fld.datatype {
	case core.DTLogical:
		val, err := fld.dbf.Record.ReadLogical(fld.offset)
		if err != nil {
			return false, NewFieldError(fld, "call to read logical failed").SetWrapped(err).SetContext("as logical")
		}
		return val, nil
	case core.DTCharacter:
		val, err := fld.dbf.Record.ReadString(fld.offset, fld.size, true, false, false)
		if err != nil {
			return false, NewFieldError(fld, "call to read string failed").SetWrapped(err).SetContext("as logical")
		}
		if len(val) == 0 {
			return false, nil
		}
		switch val[0] {
		case 'T', 't', 'Y', 'y':
			return true, nil
		case 'F', 'f', 'N', 'n', '?', ' ':
			return false, nil
		default:
			return false, NewFieldErrorf(fld, "invalid logical character: %c", val[0]).SetContext("as logical")
		}
	default:
		return false, NewFieldErrorf(fld, "unsupported data type %d", fld.datatype).SetContext("as logical")
	}
}

// MustAsLogical returns the field value as a bool and panics if an error occurs.
func (fld *Field) MustAsLogical() bool {
	val, err := fld.AsLogical()
	if err != nil {
		panic(err)
	}
	return val
}
