package json

import (
	"bytes"
	"encoding/base64"
	"encoding/json"

	. "github.com/umk/go-dymessage"
	"github.com/umk/go-dymessage/internal/helpers"
)

type encoder struct {
	buf  bytes.Buffer
	json *json.Encoder
}

// Encode transforms the data from the dynamic entity to a buffer, containing
// the JSON. If the entity type doesn't correspond the data type of the message
// definition, the method will panic.
func Encode(e *Entity, pd *MessageDef) ([]byte, error) {
	helpers.DataTypesMustMatch(e, pd)
	var ec encoder
	ec.buf.Grow(1024)
	ec.json = json.NewEncoder(&ec.buf)
	if err := ec.encode(e, pd); err != nil {
		return nil, err
	}
	return ec.buf.Bytes(), nil
}

func (ec *encoder) encode(e *Entity, pd *MessageDef) (err error) {
	ec.buf.WriteRune('{')
	n := len(pd.Fields)
	for i, f := range pd.Fields {
		if err = ec.json.Encode(f.Name); err != nil {
			return
		}
		ec.buf.WriteRune(':')
		if f.Repeated {
			if f.DataType.IsRefType() {
				err = ec.encodeJsonRefs(e, pd, f)
			} else {
				err = ec.encodeJsonValues(e, f)
			}
		} else if f.DataType.IsRefType() {
			item := e.Entities[f.Offset]
			if item != nil {
				err = ec.encodeJsonRef(item, pd, f)
			} else {
				err = ec.json.Encode(nil)
			}
		} else {
			value := f.GetPrimitive(e)
			err = ec.encodeJsonValue(value, f)
		}
		if i != (n - 1) {
			ec.buf.WriteRune(',')
		}
	}
	ec.buf.WriteRune('}')
	return
}

func (ec *encoder) encodeJsonValue(value Primitive, f *MessageFieldDef) (err error) {
	switch f.DataType {
	case DtInt32:
		err = ec.json.Encode(value.ToInt32())
	case DtInt64:
		err = ec.json.Encode(value.ToInt64())
	case DtUint32:
		err = ec.json.Encode(value.ToUint32())
	case DtUint64:
		err = ec.json.Encode(value.ToUint64())
	case DtFloat32:
		err = ec.json.Encode(value.ToFloat32())
	case DtFloat64:
		err = ec.json.Encode(value.ToFloat64())
	case DtBool:
		err = ec.json.Encode(value.ToBool())
	default:
		panic(f.DataType)
	}
	return
}

func (ec *encoder) encodeJsonValues(e *Entity, f *MessageFieldDef) (err error) {
	ec.buf.WriteRune('[')
	data := e.Entities[f.Offset]
	if data != nil {
		n := len(data.Data) / f.DataType.GetWidthInBytes()
		for i := 0; i < n; i++ {
			value := f.GetPrimitiveAt(e, i)
			if err = ec.encodeJsonValue(value, f); err != nil {
				return
			}
			if i != (n - 1) {
				ec.buf.WriteRune(',')
			}
		}
	}
	ec.buf.WriteRune(']')
	return
}

func (ec *encoder) encodeJsonRef(e *Entity, pd *MessageDef, f *MessageFieldDef) (err error) {
	switch f.DataType {
	case DtBytes:
		str := base64.StdEncoding.EncodeToString(e.Data)
		err = ec.json.Encode(str)
	case DtString:
		err = ec.json.Encode(string(e.Data))
	case DtEntity:
		def := pd.Registry.GetMessageDef(f.DataType)
		return ec.encode(e, def)
	default:
		panic(f.DataType)
	}
	return
}

func (ec *encoder) encodeJsonRefs(e *Entity, pd *MessageDef, f *MessageFieldDef) (err error) {
	data := e.Entities[f.Offset]
	if data == nil {
		err = ec.json.Encode(nil)
	} else {
		ec.buf.WriteRune('[')
		n := len(data.Entities)
		for i, item := range data.Entities {
			if err = ec.encodeJsonRef(item, pd, f); err != nil {
				return
			}
			if i != (n - 1) {
				ec.buf.WriteRune(',')
			}
		}
		ec.buf.WriteRune(']')
	}
	return
}
