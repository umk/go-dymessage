package json

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"encoding/base64"
	"encoding/json"

	. "github.com/umk/go-dymessage"
)

func (s *Encoder) Decode(b []byte, pd *MessageDef) (*Entity, error) {
	e := pd.NewEntity()
	return e, s.DecodeInto(b, pd, e)
}

func (s *Encoder) DecodeInto(b []byte, pd *MessageDef, e *Entity) error {
	var data fields
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	if err := decoder.Decode(&data); err != nil {
		return err
	}
	if err := s.setJsonFields(e, pd, data); err != nil {
		return err
	}
	return nil
}

func (s *Encoder) setJsonFields(e *Entity, pd *MessageDef, data fields) error {
	count := 0
	for _, f := range pd.Fields {
		value, ok := data[f.Name]
		if !ok {
			if !s.IgnoreUnknown {
				message := fmt.Sprintf("Unknown field name %s in the message", f.Name)
				return errors.New(message)
			}
			continue
		}
		count++
		if value != nil {
			var err error
			if f.DataType.IsRefType() {
				err = s.decodeJsonRef(e, pd, f, value)
			} else {
				err = s.decodeJsonValue(e, f, value)
			}
			if err != nil {
				return err
			}
		}
	}
	if s.RequireAll && count < len(data) {
		return errors.New("some of the fields are missing in the message")
	}
	return nil
}

func (s *Encoder) decodeJsonValue(e *Entity, f *MessageFieldDef, value interface{}) error {
	if f.Repeated {
		if items, ok := value.([]interface{}); ok {
			f.Reserve(e, len(items))
			for i, item := range items {
				if value, err := s.decodePrimitive(item, f); err == nil {
					_ = f.SetValueAt(e, i, value)
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

func (s *Encoder) decodeJsonRef(
	e *Entity, pd *MessageDef, f *MessageFieldDef, value interface{}) error {
	if f.Repeated {
		if items, ok := value.([]interface{}); ok {
			f.Reserve(e, len(items))
			for i, item := range items {
				if value, err := s.decodeEntity(item, pd, f); err == nil {
					_ = f.SetEntityAt(e, i, value)
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

func (s *Encoder) decodeEntity(
	value interface{}, pd *MessageDef, f *MessageFieldDef) (Reference, error) {
	switch value := value.(type) {
	case string:
		switch f.DataType {
		case DtBytes:
			if b, err := base64.StdEncoding.DecodeString(value); err == nil {
				return FromBytes(b, false), nil
			} else {
				return GetDefaultReference(), err
			}
		case DtString:
			return FromString(value), nil
		}
	case map[string]interface{}:
		def := pd.Registry.Defs[f.DataType&^DtEntity]
		e := def.NewEntity()
		if err := s.setJsonFields(e, pd, value); err != nil {
			return GetDefaultReference(), err
		}
		return FromEntity(e), nil
	}
	err := fmt.Errorf("unexpected value %q of %v", value, reflect.TypeOf(value))
	return GetDefaultReference(), err
}

func (*Encoder) decodePrimitive(value interface{}, f *MessageFieldDef) (Primitive, error) {
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
	return GetDefaultPrimitive(), err
}
