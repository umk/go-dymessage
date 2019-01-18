package json

import (
	"bytes"
	"fmt"
	"strconv"

	"encoding/base64"
	"encoding/json"

	. "github.com/umk/go-dymessage"
	//"github.com/umk/go-dymessage/internal/helpers"
)

type decoder struct {
	json *json.Decoder
}

//// DecodeNew transforms the JSON representation of the message to a dynamic
//// entity against the provided message definition.
//func DecodeNew(b []byte, pd *MessageDef) (*Entity, error) {
//	return Decode(b, pd, pd.NewEntity())
//}

// Decode transforms the JSON representation of the message to specified
// dynamic entity against the provided message definition. The returned entity
// is the one that has been provided as an input parameter e, but now populated
// with the data.
//
// If the entity type doesn't correspond the data type of the message
// definition, the method will panic.
func Decode(b []byte, pd *MessageDef) (*Entity, error) {
	buf := bytes.NewBuffer(b)
	dc := decoder{
		json: json.NewDecoder(buf),
	}
	dc.json.UseNumber()
	e, err := dc.decode(pd)
	return e, err
}

func (dc *decoder) decode(pd *MessageDef) (r *Entity, err error) {
	var tok interface{}
	if tok, err = dc.acceptAnyOf('{', nil); err != nil {
		return
	} else if tok == nil {
		return
	}
	r = pd.NewEntity()
	for dc.json.More() {
		var t interface{}
		t, err = dc.json.Token()
		if err != nil {
			return nil, err
		}
		name := t.(string)
		f, ok := pd.TryGetFieldByName(name)
		if !ok {
			continue
		}
		if f.Repeated {
			if err = dc.acceptToken('['); err != nil {
				return nil, err
			}
			for dc.json.More() {
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
			}
			if err := dc.acceptToken(']'); err != nil {
				return nil, err
			}
		} else {
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
		}
	}
	if err = dc.acceptToken('}'); err != nil {
		return
	}
	return
}

func (dc *decoder) decodeJsonValue(f *MessageFieldDef) (pr Primitive, err error) {
	if f.DataType == DtBool {
		if b, err := dc.acceptBoolean(); err != nil {
			return pr, err
		} else {
			return FromBool(b), nil
		}
	} else if n, err := dc.acceptNumber(); err != nil {
		return pr, err
	} else {
		switch f.DataType {
		case DtInt32:
			var value int64
			if value, err = strconv.ParseInt(n.String(), 10, 32); err == nil {
				return FromInt32(int32(value)), nil
			}
		case DtInt64:
			var value int64
			if value, err = strconv.ParseInt(n.String(), 10, 64); err == nil {
				return FromInt64(value), nil
			}
		case DtUint32:
			var value uint64
			if value, err = strconv.ParseUint(n.String(), 10, 32); err == nil {
				return FromUint32(uint32(value)), nil
			}
		case DtUint64:
			var value uint64
			if value, err = strconv.ParseUint(n.String(), 10, 64); err == nil {
				return FromUint64(value), nil
			}
		case DtFloat32:
			var value float64
			if value, err = strconv.ParseFloat(n.String(), 32); err == nil {
				return FromFloat32(float32(value)), nil
			}
		case DtFloat64:
			var value float64
			if value, err = strconv.ParseFloat(n.String(), 64); err == nil {
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
	switch {
	case f.DataType == DtString:
		var str string
		if str, err = dc.acceptString(); err == nil {
			return FromString(str), nil
		}
	case f.DataType == DtBytes:
		var str string
		if str, err = dc.acceptString(); err == nil {
			var b []byte
			if b, err = base64.StdEncoding.DecodeString(str); err == nil {
				return FromBytes(b, false), nil
			}
		}
	case f.DataType.IsEntity():
		def := pd.Registry.GetMessageDef(f.DataType)
		var nested *Entity
		if nested, err = dc.decode(def); err == nil {
			return FromEntity(nested), nil
		}
	}
	return
}

// -----------------------------------------------------------------------------
// Generic accept methods

func (dc *decoder) acceptToken(r rune) (err error) {
	_, err = dc.acceptAnyOf(r)
	return
}

func (dc *decoder) acceptAnyOf(tokens ...interface{}) (r interface{}, err error) {
	var tok interface{}
	if tok, err = dc.json.Token(); err == nil {
		for _, r = range tokens {
			if tok == r {
				return
			}
		}
		err = fmt.Errorf("dymessage: unexpected token %q", tok)
	}
	return
}

// -----------------------------------------------------------------------------
// Specific accept methods

func (dc *decoder) acceptNumber() (json.Number, error) {
	if tok, err := dc.json.Token(); err != nil {
		return json.Number(0), err
	} else if n, ok := tok.(json.Number); !ok {
		err := fmt.Errorf("dymessage: expected a number, but found %q", tok)
		return n, err
	} else {
		return n, nil
	}
}

func (dc *decoder) acceptBoolean() (bool, error) {
	if tok, err := dc.json.Token(); err != nil {
		return false, err
	} else if b, ok := tok.(bool); !ok {
		err := fmt.Errorf("dymessage: expected a boolean, but found %q", tok)
		return b, err
	} else {
		return b, nil
	}
}

func (dc *decoder) acceptString() (string, error) {
	if tok, err := dc.json.Token(); err != nil {
		return "", err
	} else if str, ok := tok.(string); !ok {
		err := fmt.Errorf("dymessage: expected a string, but found %q", tok)
		return str, err
	} else {
		return str, nil
	}
}
