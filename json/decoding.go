package json

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/umk/go-dymessage"
)

func (s *Encoder) Decode(b []byte, pd *dymessage.MessageDef) (*dymessage.Entity, error) {
	var data fields
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

func (s *Encoder) setJsonFields(
	e *dymessage.Entity, pd *dymessage.MessageDef, data fields) error {
	count := 0
	for _, f := range pd.Fields {
		value, ok := data[f.Name]
		if !ok {
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
	if !s.Relaxed && count < len(data) {
		return errors.New("bad message")
	}
	return nil
}

func (s *Encoder) decodeJsonValue(
	e *dymessage.Entity, f *dymessage.MessageFieldDef, value interface{}) error {
	if f.Repeated {
		if items, ok := value.([]interface{}); ok {
			f.Reserve(e, len(items))
			for i, item := range items {
				if value, err := s.decodePrimitive(item, f); err == nil {
					if err := f.SetValueAt(e, i, value); err != nil {
						panic(err)
					}
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
	e *dymessage.Entity,
	pd *dymessage.MessageDef,
	f *dymessage.MessageFieldDef,
	value interface{}) error {
	if f.Repeated {
		if items, ok := value.([]interface{}); ok {
			f.Reserve(e, len(items))
			for i, item := range items {
				if value, err := s.decodeEntity(item, pd, f); err == nil {
					if err := f.SetEntityAt(e, i, value); err != nil {
						panic(err)
					}
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
	value interface{},
	pd *dymessage.MessageDef,
	f *dymessage.MessageFieldDef) (*dymessage.Reference, error) {
	switch value := value.(type) {
	case string:
		switch f.DataType {
		case dymessage.DtBytes:
			if b, err := base64.StdEncoding.DecodeString(value); err == nil {
				return dymessage.FromBytes(b, false), nil
			} else {
				return nil, err
			}
		case dymessage.DtString:
			return dymessage.FromString(value), nil
		}
	case map[string]interface{}:
		def := pd.Registry.Defs[f.DataType&^dymessage.DtEntity]
		e := def.NewEntity()
		if err := s.setJsonFields(e, pd, value); err != nil {
			return nil, err
		}
		return dymessage.FromEntity(e), nil
	}
	err := fmt.Errorf("unexpected value %q of %v", value, reflect.TypeOf(value))
	return nil, err
}

func (*Encoder) decodePrimitive(
	value interface{}, f *dymessage.MessageFieldDef) (dymessage.Primitive, error) {
	switch value := value.(type) {
	case json.Number:
		str := value.String()
		switch f.DataType {
		case dymessage.DtInt32:
			value, err := strconv.ParseInt(str, 10, 32)
			return dymessage.FromInt32(int32(value)), err
		case dymessage.DtInt64:
			value, err := strconv.ParseInt(str, 10, 64)
			return dymessage.FromInt64(value), err
		case dymessage.DtUint32:
			value, err := strconv.ParseUint(str, 10, 32)
			return dymessage.FromUint32(uint32(value)), err
		case dymessage.DtUint64:
			value, err := strconv.ParseUint(str, 10, 64)
			return dymessage.FromUint64(value), err
		case dymessage.DtFloat32:
			value, err := strconv.ParseFloat(str, 32)
			return dymessage.FromFloat32(float32(value)), err
		case dymessage.DtFloat64:
			value, err := strconv.ParseFloat(str, 64)
			return dymessage.FromFloat64(value), err
		}
	case bool:
		if f.DataType == dymessage.DtBool {
			return dymessage.FromBool(value), nil
		}
	}
	err := fmt.Errorf("unexpected value %q of %v", value, reflect.TypeOf(value))
	return dymessage.PrimitiveDefault, err
}
