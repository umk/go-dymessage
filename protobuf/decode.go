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
			continue
		}
		if wire == WireBytes {
			err = ec.decodeRef(e, pd, f)
		} else {
			err = ec.decodeValue(e, wire, f)
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
	if f == nil {
		// The field has not been found, but if we've managed to reach this point,
		// it doesn't matter, so returning without an error.
		return nil
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

func (ec *Encoder) decodeValue(e *Entity, wire uint64, f *MessageFieldDef) error {
	var value uint64
	var err error
	switch wire {
	case WireVarint:
		value, err = ec.buf.DecodeVarint()
	case WireFixed32:
		value, err = ec.buf.DecodeFixed32()
	case WireFixed64:
		value, err = ec.buf.DecodeFixed64()
	default:
		message := fmt.Sprintf("The wire format %d is not supported.", wire)
		return errors.New(message)
	}
	if f == nil {
		// The field has not been found, but if we've managed to reach
		// this point, it doesn't matter, so returning without an error.
		return nil
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
			if err != nil {
				panic(err)
			}
		} else {
			f.SetPrimitive(e, Primitive(value))
		}
	}
	return err
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
