package dymessage

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

type (
	JsonEncoder struct {
		// Indicates whether encoder should produce human-readable output.
		Ident bool
		// Indicates whether the unknown fields must be silently skipped.
		Relaxed bool
	}

	// A map between the names of JSON fields and its values.
	jsonFields map[string]interface{}
)

// -----------------------------------------------------------------------------
// Encoding

func (s *JsonEncoder) Encode(e *Entity, pd *ProtoDef) ([]byte, error) {
	fields := s.getJsonFields(e, pd)
	if s.Ident {
		return json.MarshalIndent(fields, "", "\t")
	} else {
		return json.Marshal(fields)
	}
}

func (s *JsonEncoder) getJsonFields(e *Entity, pd *ProtoDef) jsonFields {
	fields := make(jsonFields)
	for _, f := range pd.Fields {
		if f.Repeated {
			var values []interface{}
			if f.DataType.isRef() {
				values = s.encodeJsonRefs(e, pd, f)
			} else {
				values = s.encodeJsonValues(e, f)
			}
			fields[f.Name] = values
		} else if f.DataType.isRef() {
			item := e.Entities[f.index]
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

func (*JsonEncoder) encodeJsonValue(
	value Primitive, f *ProtoField) interface{} {
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

func (s *JsonEncoder) encodeJsonValues(
	e *Entity, f *ProtoField) (result []interface{}) {
	data := e.Entities[f.index]
	if data != nil {
		n := len(data.Data) / f.DataType.getSize()
		for i := 0; i < n; i++ {
			value, _ := f.GetValueAt(e, i)
			result = append(result, s.encodeJsonValue(value, f))
		}
	}
	return
}

func (s *JsonEncoder) encodeJsonRef(
	e *Entity, pd *ProtoDef, f *ProtoField) interface{} {
	switch f.DataType {
	case DtBytes:
		return base64.StdEncoding.EncodeToString(e.Data)
	case DtString:
		return string(e.Data)
	default:
		def := pd.repo.defs[f.DataType&^DtEntity]
		return s.getJsonFields(e, def)
	}
}

func (s *JsonEncoder) encodeJsonRefs(
	e *Entity, pd *ProtoDef, f *ProtoField) (result []interface{}) {
	data := e.Entities[f.index]
	if data == nil {
		return nil
	}
	for _, item := range data.Entities {
		value := s.encodeJsonRef(item, pd, f)
		result = append(result, value)
	}
	return
}

// -----------------------------------------------------------------------------
// Decoding

func (s *JsonEncoder) Decode(b []byte, pd *ProtoDef) (*Entity, error) {
	var data jsonFields
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	e := pd.NewEntity()
	if err := s.setJsonFields(e, pd, data); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *JsonEncoder) setJsonFields(
	e *Entity, pd *ProtoDef, data jsonFields) error {
	count := 0
	for _, f := range pd.Fields {
		value, ok := data[f.Name]
		if !ok {
			continue
		}
		count++
		if value != nil {
			var err error
			if f.DataType.isRef() {
				err = s.decodeJsonRef(e, pd, f, value)
			} else {
				err = s.decodeJsonValue(e, f, value)
			}
			if err != nil {
				return err
			}
		}
	}
	if !s.Relaxed && count < len(data) {
		return BadMessageErr
	}
	return nil
}

func (s *JsonEncoder) decodeJsonValue(
	e *Entity, f *ProtoField, value interface{}) error {
	if f.Repeated {
		if items, ok := value.([]interface{}); ok {
			f.Reserve(e, len(items))
			for i, item := range items {
				if value, err := s.decodePrimitive(item, f); err == nil {
					f.SetValueAt(e, i, value)
				} else {
					return err
				}
			}
		} else {
			return fmt.Errorf("expected array, but %q provided", value)
		}
	} else {
		if value, err := s.decodePrimitive(value, f); err == nil {
			f.SetValue(e, value)
		} else {
			return err
		}
	}
	return nil
}

func (s *JsonEncoder) decodeJsonRef(
	e *Entity, pd *ProtoDef, f *ProtoField, value interface{}) error {
	if f.Repeated {
		if items, ok := value.([]interface{}); ok {
			f.Reserve(e, len(items))
			for i, item := range items {
				if value, err := s.decodeEntity(item, pd, f); err == nil {
					f.SetEntityAt(e, i, value)
				} else {
					return err
				}
			}
		} else {
			return fmt.Errorf("expected array, but %q provided", value)
		}
	} else {
		if value, err := s.decodeEntity(value, pd, f); err == nil {
			f.SetEntity(e, value)
		} else {
			return err
		}
	}
	return nil
}

func (s *JsonEncoder) decodeEntity(
	value interface{}, pd *ProtoDef, f *ProtoField) (*Reference, error) {
	switch value := value.(type) {
	case string:
		switch f.DataType {
		case DtBytes:
			if b, err := base64.StdEncoding.DecodeString(value); err == nil {
				return FromBytes(b, false), nil
			} else {
				return nil, err
			}
		case DtString:
			return FromString(value), nil
		}
	case map[string]interface{}:
		def := pd.repo.defs[f.DataType&^DtEntity]
		e := def.NewEntity()
		if err := s.setJsonFields(e, pd, value); err != nil {
			return nil, err
		}
		return FromEntity(e), nil
	}
	err := fmt.Errorf("unexpected value %q of %v", value, reflect.TypeOf(value))
	return nil, err
}

func (*JsonEncoder) decodePrimitive(
	value interface{}, f *ProtoField) (Primitive, error) {
	switch value := value.(type) {
	case json.Number:
		str := value.String()
		switch f.DataType {
		case DtInt32:
			value, err := strconv.ParseInt(str, 10, 32)
			return FromInt32(int32(value)), err
		case DtInt64:
			value, err := strconv.ParseInt(str, 10, 64)
			return FromInt64(value), err
		case DtUint32:
			value, err := strconv.ParseUint(str, 10, 32)
			return FromUint32(uint32(value)), err
		case DtUint64:
			value, err := strconv.ParseUint(str, 10, 64)
			return FromUint64(value), err
		case DtFloat32:
			value, err := strconv.ParseFloat(str, 32)
			return FromFloat32(float32(value)), err
		case DtFloat64:
			value, err := strconv.ParseFloat(str, 64)
			return FromFloat64(value), err
		}
	case bool:
		if f.DataType == DtBool {
			return FromBool(value), nil
		}
	}
	err := fmt.Errorf("unexpected value %q of %v", value, reflect.TypeOf(value))
	return primitiveDefault, err
}
