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
	"github.com/umk/go-dymessage/internal/helpers"
)

type decoder struct {
	json *json.Decoder
}

// DecodeNew transforms the JSON representation of the message to a dynamic
// entity against the provided message definition.
func DecodeNew(b []byte, pd *MessageDef) (*Entity, error) {
	return Decode(b, pd, pd.NewEntity())
}

// Decode transforms the JSON representation of the message to specified
// dynamic entity against the provided message definition. The returned entity
// is the one that has been provided as an input parameter e, but now populated
// with the data.
//
// If the entity type doesn't correspond the data type of the message
// definition, the method will panic.
func Decode(b []byte, pd *MessageDef, e *Entity) (*Entity, error) {
	buf := bytes.NewBuffer(b)
	dc := decoder{
		json: json.NewDecoder(buf),
	}
	dc.json.UseNumber()
	err := dc.decode(pd, e)
	return e, err
}

func (dc *decoder) decode(pd *MessageDef, e *Entity) error {
	if err := dc.acceptToken('{'); err != nil {
		return err
	}
	for dc.json.More() {
		t, err := dc.json.Token()
		if err != nil {
			return err
		}
		name := t.(string)
		f, ok := pd.TryGetFieldByName(name)
		if !ok {
			continue
		}
		if f.Repeated {
			if err := dc.acceptToken('['); err != nil {
				return err
			}
			if f.DataType.IsRefType() {
				err = dc.decodeJsonRefs()
			} else {
				err = dc.decodeJsonValues()
			}
			if err := dc.acceptToken(']'); err != nil {
				return err
			}
		} else {

		}
	}
	if err := dc.acceptToken('}'); err != nil {
		return err
	}
	return nil
}

func (dc *decoder) decodeJsonValue(f *MessageFieldDef) (pr Primitive, err error) {
	switch f.DataType {
	case DtBool:
		if b, err := dc.acceptBoolean(); err != nil {
			return pr, err
		} else {
			return FromBool(b), nil
		}
	case DtInt32, DtInt64, DtUint32, DtUint64:
		if n, err := dc.acceptNumber(); err != nil {
			return pr, err
		} else if i, err := n.Int64(); err != nil  {
			return pr, err
		} else {
			return FromInt64(i), nil
		}
	case DtFloat32, DtFloat64:
		if n, err := dc.acceptNumber(); err != nil {
			return pr, err
		} else if f, err := n.Float64(); err != nil {
			return pr, err
		} else {
			return FromFloat64(f), nil
		}
	default:
		panic(f.DataType)
	}
	return
}

//func (dc *decoder) decodeJsonRefs() error {
//
//}
//
//func (dc *decoder) decodeJsonValues() error {
//
//}

//func (dc *decoder) decodeJsonRefs(e *Entity, pd *MessageDef, f *MessageFieldDef) error {
//	if f.Repeated {
//		if items, ok := value.([]interface{}); ok {
//			f.Reserve(e, len(items))
//			for i, item := range items {
//				if value, err := s.decodeEntity(item, pd, f); err == nil {
//					f.SetReferenceAt(e, i, value)
//				} else {
//					return err
//				}
//			}
//		} else {
//			return fmt.Errorf("expected array, but %q provided", value)
//		}
//	} else {
//		if value, err := s.decodeEntity(value, pd, f); err == nil {
//			f.SetReference(e, value)
//		} else {
//			return err
//		}
//	}
//	return nil
//}
//
//func (dc *decoder) decodeJsonValue(e *Entity, f *MessageFieldDef) error {
//	if f.Repeated {
//		if items, ok := value.([]interface{}); ok {
//			f.Reserve(e, len(items))
//			for i, item := range items {
//				if value, err := s.decodePrimitive(item, f); err == nil {
//					f.SetPrimitiveAt(e, i, value)
//				} else {
//					return err
//				}
//			}
//		} else {
//			return fmt.Errorf("expected array, but %q provided", value)
//		}
//	} else {
//		if value, err := s.decodePrimitive(value, f); err == nil {
//			f.SetPrimitive(e, value)
//		} else {
//			return err
//		}
//	}
//	return nil
//}

// -----------------------------------------------------------------------------
// Helper methods

func (dc *decoder) acceptToken(token rune) error {
	if tok, err := dc.json.Token(); err != nil {
		return err
	} else if tok != token {
		return fmt.Errorf("dymessage: expected token %q, but found %q", token, tok)
	}
	return nil
}

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

func (s *Encoder) Decode(b []byte, pd *MessageDef, e *Entity) (*Entity, error) {
	helpers.DataTypesMustMatch(e, pd)
	var data map[string]interface{}
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	if err := s.setJsonFields(e, pd, data); err != nil {
		return nil, err
	}
	return e, nil
}

func (s *Encoder) setJsonFields(
    e *Entity, pd *MessageDef, data map[string]interface{}) error {
	for _, f := range pd.Fields {
		value, ok := data[f.Name]
		if !ok {
			if !s.IgnoreUnknown {
				message := fmt.Sprintf("unknown field name %q in the message", f.Name)
				return errors.New(message)
			}
			continue
		}
		if value != nil {
			var err error

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
		def := pd.Registry.GetMessageDef(f.DataType)
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
		str := string(value)
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


//// DecodeNew transforms the JSON representation of the message to a dynamic entity
//// against the provided message definition.
//func (s *Encoder) DecodeNew(b []byte, pd *MessageDef) (*Entity, error) {
//	return s.Decode(b, pd, pd.NewEntity())
//}
//
//// Decode transforms the JSON representation of the message to specified
//// dynamic entity against the provided message definition. The returned entity
//// is the one that has been provided as an input parameter e, but now populated
//// with the data.
////
//// If the entity type doesn't correspond the data type of the message
//// definition, the method will panic.
//func (s *Encoder) Decode(b []byte, pd *MessageDef, e *Entity) (*Entity, error) {
//	helpers.DataTypesMustMatch(e, pd)
//	var data map[string]interface{}
//	decoder := json.NewDecoder(bytes.NewReader(b))
//	decoder.UseNumber()
//	if err := decoder.Decode(&data); err != nil {
//		return nil, err
//	}
//	if err := s.setJsonFields(e, pd, data); err != nil {
//		return nil, err
//	}
//	return e, nil
//}
//
//func (s *Encoder) setJsonFields(
//	e *Entity, pd *MessageDef, data map[string]interface{}) error {
//	for _, f := range pd.Fields {
//		value, ok := data[f.Name]
//		if !ok {
//			if !s.IgnoreUnknown {
//				message := fmt.Sprintf("unknown field name %q in the message", f.Name)
//				return errors.New(message)
//			}
//			continue
//		}
//		if value != nil {
//			var err error
//			if f.DataType.IsRefType() {
//				err = s.decodeJsonRefs(e, pd, f, value)
//			} else {
//				err = s.decodeJsonValue(e, f, value)
//			}
//			if err != nil {
//				return err
//			}
//		}
//	}
//	return nil
//}
//
//func (s *Encoder) decodeJsonValue(e *Entity, f *MessageFieldDef, value interface{}) error {
//	if f.Repeated {
//		if items, ok := value.([]interface{}); ok {
//			f.Reserve(e, len(items))
//			for i, item := range items {
//				if value, err := s.decodePrimitive(item, f); err == nil {
//					f.SetPrimitiveAt(e, i, value)
//				} else {
//					return err
//				}
//			}
//		} else {
//			return fmt.Errorf("expected array, but %q provided", value)
//		}
//	} else {
//		if value, err := s.decodePrimitive(value, f); err == nil {
//			f.SetPrimitive(e, value)
//		} else {
//			return err
//		}
//	}
//	return nil
//}
//
//func (s *Encoder) decodeJsonRefs(
//	e *Entity, pd *MessageDef, f *MessageFieldDef, value interface{}) error {
//	if f.Repeated {
//		if items, ok := value.([]interface{}); ok {
//			f.Reserve(e, len(items))
//			for i, item := range items {
//				if value, err := s.decodeEntity(item, pd, f); err == nil {
//					f.SetReferenceAt(e, i, value)
//				} else {
//					return err
//				}
//			}
//		} else {
//			return fmt.Errorf("expected array, but %q provided", value)
//		}
//	} else {
//		if value, err := s.decodeEntity(value, pd, f); err == nil {
//			f.SetReference(e, value)
//		} else {
//			return err
//		}
//	}
//	return nil
//}
//
//func (s *Encoder) decodeEntity(
//	value interface{}, pd *MessageDef, f *MessageFieldDef) (Reference, error) {
//	switch value := value.(type) {
//	case string:
//		switch f.DataType {
//		case DtBytes:
//			if b, err := base64.StdEncoding.DecodeString(value); err == nil {
//				return FromBytes(b, false), nil
//			} else {
//				return GetDefaultReference(), err
//			}
//		case DtString:
//			return FromString(value), nil
//		}
//	case map[string]interface{}:
//		def := pd.Registry.GetMessageDef(f.DataType)
//		e := def.NewEntity()
//		if err := s.setJsonFields(e, pd, value); err != nil {
//			return GetDefaultReference(), err
//		}
//		return FromEntity(e), nil
//	}
//	err := fmt.Errorf("unexpected value %q of %v", value, reflect.TypeOf(value))
//	return GetDefaultReference(), err
//}
//
//func (*Encoder) decodePrimitive(value interface{}, f *MessageFieldDef) (Primitive, error) {
//	switch value := value.(type) {
//	case json.Number:
//		str := string(value)
//		switch f.DataType {
//		case DtInt32:
//			value, err := strconv.ParseInt(str, 10, 32)
//			return FromInt32(int32(value)), err
//		case DtInt64:
//			value, err := strconv.ParseInt(str, 10, 64)
//			return FromInt64(value), err
//		case DtUint32:
//			value, err := strconv.ParseUint(str, 10, 32)
//			return FromUint32(uint32(value)), err
//		case DtUint64:
//			value, err := strconv.ParseUint(str, 10, 64)
//			return FromUint64(value), err
//		case DtFloat32:
//			value, err := strconv.ParseFloat(str, 32)
//			return FromFloat32(float32(value)), err
//		case DtFloat64:
//			value, err := strconv.ParseFloat(str, 64)
//			return FromFloat64(value), err
//		}
//	case bool:
//		if f.DataType == DtBool {
//			return FromBool(value), nil
//		}
//	}
//	err := fmt.Errorf("unexpected value %q of %v", value, reflect.TypeOf(value))
//	return GetDefaultPrimitive(), err
//}
