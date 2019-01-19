package json

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"encoding/base64"
	. "github.com/umk/go-dymessage"
	//"github.com/umk/go-dymessage/internal/helpers"
)

type decoder struct {
	//json *json.Decoder
	lx lexer
	returned bool
}

func (dc *decoder) putBack() {
	if dc.returned {
		panic("ret")
	}
	dc.returned = true
}

func (dc *decoder) token() *decoder {
	if dc.returned {
		dc.returned = false
		return dc
	}
	_ = dc.lx.next()
	return dc
}

func (dc *decoder) prob(t token) bool {
	return dc.lx.reader.err == nil && dc.lx.tok.kind == t
}

func (dc *decoder) tok(t token) error {
	if dc.lx.tok.kind != t {
		return fmt.Errorf("dymessage: expected %v, but found %v", t, dc.lx.tok.kind)
	}
	return nil
}

func (dc *decoder) null() bool {
	return dc.lx.reader.err == nil && dc.lx.tok.kind == tokNull
}

func (dc *decoder) boolean() (bool, error) {
	if dc.lx.reader.err != nil {
		return false, dc.lx.reader.err
	}
	if dc.lx.tok.kind != tokBool {
		return false, errors.New("b")
	}
	return dc.lx.tok.bool, nil
}

func (dc *decoder) string() (string, error) {
	if dc.lx.reader.err != nil {
		return "", dc.lx.reader.err
	}
	if dc.lx.tok.kind != tokString {
		return "", fmt.Errorf("dymessage: expected string, but found %v", dc.lx.tok.kind)
	}
	return dc.lx.tok.string, nil
}

func (dc *decoder) number() (string, error) {
	if dc.lx.reader.err != nil {
		return "", dc.lx.reader.err
	}
	if dc.lx.tok.kind != tokNumber {
		return "", errors.New("d")
	}
	return dc.lx.tok.number, nil
}

//func (dc *decoder) token() (tok token) {
//	tok.tok, tok.err = dc.json.Token()
//	return tok
//}

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
func Decode(b []byte, pd *MessageDef) (e *Entity, err error) {
	buf := bytes.NewBuffer(b)
	dc := decoder{ }
	dc.lx.reader.reset(buf)
	//dc.json.UseNumber()
	if err = dc.token().tok(tokCrBrOpen); err != nil {
		return
	}
	if e, err = dc.decode(pd); err != nil {
		return
	}
	//if err = dc.token().tok(tokCrBrClose); err != nil {
	//	return
	//}
	return
}

func (dc *decoder) decode(pd *MessageDef) (r *Entity, err error) {
	r = pd.NewEntity()
	if dc.token().prob(tokCrBrClose) {
		return
	}
	dc.putBack()
	for {
		var name string
		if name, err = dc.token().string(); err != nil {
			return
		}
		if err = dc.token().tok(tokColon); err != nil {
			return
		}
		f, ok := pd.TryGetFieldByName(name)
		if !ok {
			continue
		}
		if f.Repeated {
			if tok := dc.token(); !tok.null() {
				if err = tok.tok(tokSqBrOpen); err != nil {
					return
				}
				if !tok.token().prob(tokSqBrClose) {
					dc.putBack()
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
						if !dc.token().prob(tokComma) {
							dc.putBack()
							break
						}
					}
					if err = dc.token().tok(tokSqBrClose); err != nil {
						return
					}
				}
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
		if !dc.token().prob(tokComma) {
			dc.putBack()
			break
		}
	}
	if err = dc.token().tok(tokCrBrClose); err != nil {
		return
	}
	return
}

func (dc *decoder) decodeJsonValue(f *MessageFieldDef) (pr Primitive, err error) {
	tok := dc.token()
	if f.DataType == DtBool {
		if b, err := tok.boolean(); err != nil {
			return pr, err
		} else {
			return FromBool(b), nil
		}
	} else if n, err := tok.number(); err != nil {
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
	tok := dc.token()
	if tok.lx.reader.err != nil {
		return ref, tok.lx.reader.err
	}
	if tok.null() {
		return ref, nil
	}
	switch {
	case f.DataType == DtString:
		var str string
		if str, err = tok.string(); err == nil {
			return FromString(str), nil
		}
	case f.DataType == DtBytes:
		var str string
		if str, err = tok.string(); err != nil {
			return
		}
		var b []byte
		if b, err = base64.StdEncoding.DecodeString(str); err == nil {
			return FromBytes(b, false), nil
		}
	case f.DataType.IsEntity():
		if err = tok.tok(tokCrBrOpen); err != nil {
			return
		}
		def := pd.Registry.GetMessageDef(f.DataType)
		var nested *Entity
		if nested, err = dc.decode(def); err != nil {
			return
		}
		//if err = dc.token().tok(tokCrBrClose); err != nil {
		//	return
		//}
		return FromEntity(nested), nil
	}
	return
}
