package protobuf

import (
	"errors"
	"fmt"

	. "github.com/umk/go-dymessage"
	. "github.com/umk/go-dymessage/protobuf/internal/impl"
)

func (ec *Encoder) Decode(b []byte, pd *MessageDef) (*Entity, error) {
	return ec.DecodeInto(b, pd, pd.NewEntity())
}

func (ec *Encoder) DecodeInto(b []byte, pd *MessageDef, e *Entity) (*Entity, error) {
	ec.buf.SetBuf(b)
	defer ec.buf.Reset()
	for !ec.buf.Eob() {
		t, err := ec.buf.DecodeVarint()
		if err != nil {
			return nil, err
		}
		wire, tag := t&7, t>>3
		f, ok := pd.Fields[tag]
		if !ok {
			if !ec.IgnoreUnknown {
				message := fmt.Sprintf("Unexpected tag %d in the message", tag)
				return nil, errors.New(message)
			}
			if err = ec.skipValue(wire); err != nil {
				return nil, err
			}
			continue
		}
		if wire == WireBytes {
			err = ec.decodeRef(e, pd, f)
		} else {
			err = ec.decodeValue(e, f)
		}
		if err != nil {
			return nil, err
		}
	}
	return e, nil
}

func (ec *Encoder) decodeRef(e *Entity, pd *MessageDef, f *MessageFieldDef) error {
	value, err := ec.buf.DecodeRawBytes(false)
	if err != nil {
		return err
	}
	if !f.DataType.IsRefType() {
		return decodePacked(e, f, value)
	}
	var entity *Entity
	if (f.DataType & DtEntity) != 0 {
		def := pd.Registry.GetMessageDef(f.DataType)
		another := ec.clone()
		entity, err = another.Decode(value, def)
		if err != nil {
			return err
		}
	} else {
		entity = &Entity{
			Data: make([]byte, len(value)),
		}
		copy(entity.Data, value)
	}
	if f.Repeated {
		data := e.Entities[f.Offset]
		if data == nil {
			data = new(Entity)
			e.Entities[f.Offset] = data
		}
		data.Entities = append(data.Entities, entity)
	} else {
		e.Entities[f.Offset] = entity
	}
	return nil
}

func (ec *Encoder) skipValue(wire uint64) (err error) {
	switch wire {
	case WireVarint:
		_, err = ec.buf.DecodeVarint()
	case WireFixed32:
		_, err = ec.buf.DecodeFixed32()
	case WireFixed64:
		_, err = ec.buf.DecodeFixed64()
	case WireBytes:
		_, err = ec.buf.DecodeRawBytes(false)
	default:
		message := fmt.Sprintf("The wire format %d is not supported.", wire)
		err = errors.New(message)
	}
	return
}

func (ec *Encoder) decodeValue(e *Entity, f *MessageFieldDef) (err error) {
	var value uint64
	extension, ok := tryGetExtension(f)
	if ok && extension.integerKind != ikDefault {
		ik := extension.integerKind
		value, err = ec.decodeValueByKind(f, ik)
	} else {
		switch f.DataType {
		case DtInt32:
		case DtUint32:
		case DtFloat32:
			value, err = ec.buf.DecodeFixed32()
		case DtInt64:
		case DtUint64:
		case DtFloat64:
			value, err = ec.buf.DecodeFixed64()
		case DtBool:
			value, err = ec.buf.DecodeVarint()
		default:
			panic(fmt.Sprintf("unsupported decoding data type %d", f.DataType))
		}
	}
	if err == nil {
		if f.Repeated {
			entity := e.Entities[f.Offset]
			if entity == nil {
				entity = &Entity{}
				e.Entities[f.Offset] = entity
			}
			n := f.Reserve(entity, 1)
			f.SetPrimitiveAt(entity, n, Primitive(value))
		} else {
			f.SetPrimitive(e, Primitive(value))
		}
	}
	return err
}

func (ec *Encoder) decodeValueByKind(
	f *MessageFieldDef, ik integerKind) (uint64, error) {
	switch ik {
	case ikVarint:
		return ec.buf.DecodeVarint()
	case ikZigZag:
		switch f.DataType {
		case DtInt32:
			return ec.buf.DecodeZigzag32()
		case DtInt64:
			return ec.buf.DecodeZigzag64()
		default:
			panic(fmt.Sprintf("ZigZag encoding is applied to invalid data type %d", f.DataType))
		}
	default:
		panic(fmt.Sprintf("unsupported value of integer kind %d", ik))
	}
}

func decodePacked(e *Entity, f *MessageFieldDef, value []byte) error {
	buf := NewBuffer(value)
	var err error
	for !buf.Eob() {
		var i uint64
		switch f.DataType {
		case DtInt32, DtInt64, DtUint32, DtUint64, DtBool:
			i, err = buf.DecodeVarint()
		case DtFloat32:
			i, err = buf.DecodeFixed32()
		case DtFloat64:
			i, err = buf.DecodeFixed64()
		default:
			panic("unexpected data type")
		}
		if err != nil {
			return err
		}
		n := f.Reserve(e, 1)
		f.SetPrimitiveAt(e, n, Primitive(i))
	}
	return nil
}
