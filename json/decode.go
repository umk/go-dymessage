package json

import (
	"encoding/base64"
	"errors"
	"strconv"

	. "github.com/umk/go-dymessage"
)

type decoder struct {
	lx lexer
}

// DecodeNew transforms the JSON representation of the message to dynamic entity
// against the provided message definition.
func DecodeNew(b []byte, pd *MessageDef) (e *Entity, err error) {
	var dc decoder
	dc.lx.reader.reset(b)
	dc.lx.next()
	if e, err = dc.decode(pd); err == nil {
		if !dc.lx.eof() {
			message := dc.createErrorMessage(tkEof)
			err = errors.New(message)
		}
	}
	return
}

func (dc *decoder) decode(pd *MessageDef) (r *Entity, err error) {
	if err = dc.accept(tkCrBrOpen); err != nil {
		return
	}
	r = pd.NewEntity()
	if dc.tryAccept(tkCrBrClose) {
		return
	}
	for {
		if err = dc.decodeProperty(r, pd); err != nil {
			return
		}
		if !dc.tryAccept(tkComma) {
			break
		}
	}
	err = dc.accept(tkCrBrClose)
	return
}

func (dc *decoder) decodeProperty(r *Entity, pd *MessageDef) (err error) {
	var name string
	if name, err = dc.acceptValue(tkString); err != nil {
		return
	}
	if err = dc.accept(tkColon); err != nil {
		return
	}
	f, ok := pd.TryGetFieldByName(name)
	if !ok {
		err = dc.ignoreValue()
		return
	}
	if f.Repeated {
		if dc.tryAccept(tkNull) {
			// Do nothing but leave default value in the entity field.
		} else if err = dc.decodeRepeated(r, pd, f); err != nil {
			return
		}
	} else {
		if err = dc.decodeSingle(r, pd, f); err != nil {
			return
		}
	}
	return
}

func (dc *decoder) decodeSingle(
	r *Entity, pd *MessageDef, f *MessageFieldDef) (err error) {
	if f.DataType.IsRefType() {
		var ref Reference
		if ref, err = dc.decodeJsonRef(pd, f); err != nil {
			return
		}
		f.SetReference(r, ref)
	} else {
		var p Primitive
		if p, err = dc.decodeJsonValue(f); err != nil {
			return
		}
		f.SetPrimitive(r, p)
	}
	return
}

func (dc *decoder) decodeRepeated(
	r *Entity, pd *MessageDef, f *MessageFieldDef) (err error) {
	if err = dc.accept(tkSqBrOpen); err != nil {
		return
	}
	if dc.tryAccept(tkSqBrClose) {
		return
	}
	for {
		n := f.Reserve(r, 1)
		if f.DataType.IsRefType() {
			var ref Reference
			if ref, err = dc.decodeJsonRef(pd, f); err != nil {
				return
			}
			f.SetReferenceAt(r, n, ref)
		} else {
			var p Primitive
			if p, err = dc.decodeJsonValue(f); err != nil {
				return
			}
			f.SetPrimitiveAt(r, n, p)
		}
		if !dc.tryAccept(tkComma) {
			break
		}
	}
	err = dc.accept(tkSqBrClose)
	return
}

func (dc *decoder) decodeJsonValue(f *MessageFieldDef) (pr Primitive, err error) {
	if f.DataType == DtBool {
		if b, err := dc.acceptBool(); err != nil {
			return pr, err
		} else {
			return FromBool(b), nil
		}
	} else if n, err := dc.acceptValue(tkNumber); err != nil {
		return pr, err
	} else {
		switch f.DataType {
		case DtInt32:
			var value int64
			if value, err = strconv.ParseInt(n, 10, 32); err == nil {
				return FromInt32(int32(value)), nil
			}
		case DtInt64:
			var value int64
			if value, err = strconv.ParseInt(n, 10, 64); err == nil {
				return FromInt64(value), nil
			}
		case DtUint32:
			var value uint64
			if value, err = strconv.ParseUint(n, 10, 32); err == nil {
				return FromUint32(uint32(value)), nil
			}
		case DtUint64:
			var value uint64
			if value, err = strconv.ParseUint(n, 10, 64); err == nil {
				return FromUint64(value), nil
			}
		case DtFloat32:
			var value float64
			if value, err = strconv.ParseFloat(n, 32); err == nil {
				return FromFloat32(float32(value)), nil
			}
		case DtFloat64:
			var value float64
			if value, err = strconv.ParseFloat(n, 64); err == nil {
				return FromFloat64(value), nil
			}
		default:
			panic(f.DataType)
		}
		return pr, err
	}
}

func (dc *decoder) decodeJsonRef(
	pd *MessageDef, f *MessageFieldDef) (ref Reference, err error) {
	if dc.tryAccept(tkNull) {
		return ref, nil
	}
	switch {
	case f.DataType == DtString:
		var str string
		if str, err = dc.acceptValue(tkString); err == nil {
			return FromString(str), nil
		}
	case f.DataType == DtBytes:
		var str string
		if str, err = dc.acceptValue(tkString); err != nil {
			return
		}
		var b []byte
		if b, err = base64.StdEncoding.DecodeString(str); err == nil {
			return FromBytes(b, false), nil
		}
	case f.DataType.IsEntity():
		def := pd.Registry.GetMessageDef(f.DataType)
		var nested *Entity
		if nested, err = dc.decode(def); err != nil {
			return
		}
		return FromEntity(nested), nil
	}
	return
}

// -----------------------------------------------------------------------------
// Ignore methods

// ignoreValue skips the value the parser has stepped on.
func (dc *decoder) ignoreValue() (err error) {
	switch {
	case dc.tryAcceptAny(tkNumber, tkString, tkNull, tkTrue, tkFalse):
		// Do nothing.
	case dc.probably(tkCrBrOpen):
		err = dc.ignoreObject()
	case dc.probably(tkSqBrOpen):
		err = dc.ignoreArray()
	default:
		message := dc.createErrorMessage()
		err = errors.New(message)
	}
	return
}

func (dc *decoder) ignoreObject() (err error) {
	if err = dc.accept(tkCrBrOpen); err != nil {
		return
	}
	if dc.tryAccept(tkCrBrClose) {
		return
	}
	for {
		if err = dc.acceptSeq(tkString, tkColon); err != nil {
			return
		}
		if err = dc.ignoreValue(); err != nil {
			return
		}
		if !dc.tryAccept(tkComma) {
			break
		}
	}
	err = dc.accept(tkCrBrClose)
	return
}

func (dc *decoder) ignoreArray() (err error) {
	if err = dc.accept(tkSqBrOpen); err != nil {
		return
	}
	if dc.tryAccept(tkSqBrClose) {
		return
	}
	for {
		if err = dc.ignoreValue(); err != nil {
			return
		}
		if !dc.tryAccept(tkComma) {
			break
		}
	}
	err = dc.accept(tkSqBrClose)
	return
}
