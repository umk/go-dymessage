package json

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/umk/go-dymessage"
)

func (s *Encoder) Encode(e *dymessage.Entity, pd *dymessage.MessageDef) ([]byte, error) {
	fields := s.getJsonFields(e, pd)
	if s.Ident {
		return json.MarshalIndent(fields, "", "\t")
	} else {
		return json.Marshal(fields)
	}
}

func (s *Encoder) getJsonFields(e *dymessage.Entity, pd *dymessage.MessageDef) fields {
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

func (*Encoder) encodeJsonValue(
	value dymessage.Primitive, f *dymessage.MessageFieldDef) interface{} {
	var number interface{}
	switch f.DataType {
	case dymessage.DtInt32:
		number = value.ToInt32()
	case dymessage.DtInt64:
		number = value.ToInt64()
	case dymessage.DtUint32:
		number = value.ToUint32()
	case dymessage.DtUint64:
		number = value.ToUint64()
	case dymessage.DtFloat32:
		number = value.ToFloat32()
	case dymessage.DtFloat64:
		number = value.ToFloat64()
	case dymessage.DtBool:
		return value.ToBool()
	default:
		panic(f.DataType)
	}
	return json.Number(fmt.Sprint(number))
}

func (s *Encoder) encodeJsonValues(
	e *dymessage.Entity, f *dymessage.MessageFieldDef) (result []interface{}) {
	data := e.Entities[f.Offset]
	if data != nil {
		n := len(data.Data) / f.DataType.GetSizeInBytes()
		for i := 0; i < n; i++ {
			value, _ := f.GetValueAt(e, i)
			result = append(result, s.encodeJsonValue(value, f))
		}
	}
	return
}

func (s *Encoder) encodeJsonRef(
	e *dymessage.Entity, pd *dymessage.MessageDef, f *dymessage.MessageFieldDef) interface{} {
	switch f.DataType {
	case dymessage.DtBytes:
		return base64.StdEncoding.EncodeToString(e.Data)
	case dymessage.DtString:
		return string(e.Data)
	default:
		def := pd.Registry.Defs[f.DataType&^dymessage.DtEntity]
		return s.getJsonFields(e, def)
	}
}

func (s *Encoder) encodeJsonRefs(
	e *dymessage.Entity, pd *dymessage.MessageDef, f *dymessage.MessageFieldDef) (result []interface{}) {
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
