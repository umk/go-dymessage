package json

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	. "github.com/umk/go-dymessage"
)

func (s *Encoder) Encode(e *Entity, pd *MessageDef) ([]byte, error) {
	fields := s.getJsonFields(e, pd)
	if s.Ident {
		return json.MarshalIndent(fields, "", "\t")
	} else {
		return json.Marshal(fields)
	}
}

func (s *Encoder) getJsonFields(e *Entity, pd *MessageDef) fields {
	fields := make(fields)
	for _, f := range pd.Fields {
		if f.Repeated {
			var values []interface{}
			if f.DataType.IsRefType() {
				values = s.encodeJsonRefs(e, pd, f)
			} else {
				values = s.encodeJsonValues(e, f)
			}
			fields[f.Name] = values
		} else if f.DataType.IsRefType() {
			item := e.Entities[f.Offset]
			if item != nil {
				fields[f.Name] = s.encodeJsonRef(item, pd, f)
			} else {
				fields[f.Name] = nil
			}
		} else {
			value := f.GetValue(e)
			fields[f.Name] = s.encodeJsonValue(value, f)
		}
	}
	return fields
}

func (*Encoder) encodeJsonValue(value Primitive, f *MessageFieldDef) interface{} {
	var number interface{}
	switch f.DataType {
	case DtInt32:
		number = value.ToInt32()
	case DtInt64:
		number = value.ToInt64()
	case DtUint32:
		number = value.ToUint32()
	case DtUint64:
		number = value.ToUint64()
	case DtFloat32:
		number = value.ToFloat32()
	case DtFloat64:
		number = value.ToFloat64()
	case DtBool:
		return value.ToBool()
	default:
		panic(f.DataType)
	}
	return json.Number(fmt.Sprint(number))
}

func (s *Encoder) encodeJsonValues(e *Entity, f *MessageFieldDef) (result []interface{}) {
	data := e.Entities[f.Offset]
	if data != nil {
		n := len(data.Data) / f.DataType.GetWidthInBytes()
		for i := 0; i < n; i++ {
			value, _ := f.GetValueAt(e, i)
			result = append(result, s.encodeJsonValue(value, f))
		}
	}
	return
}

func (s *Encoder) encodeJsonRef(e *Entity, pd *MessageDef, f *MessageFieldDef) interface{} {
	switch f.DataType {
	case DtBytes:
		return base64.StdEncoding.EncodeToString(e.Data)
	case DtString:
		return string(e.Data)
	default:
		def := pd.Registry.Defs[f.DataType&^DtEntity]
		return s.getJsonFields(e, def)
	}
}

func (s *Encoder) encodeJsonRefs(
	e *Entity, pd *MessageDef, f *MessageFieldDef) (result []interface{}) {
	data := e.Entities[f.Offset]
	if data == nil {
		return nil
	}
	for _, item := range data.Entities {
		value := s.encodeJsonRef(item, pd, f)
		result = append(result, value)
	}
	return
}
