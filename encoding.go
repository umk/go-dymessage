package dymessage

import (
	"errors"

	"github.com/umk/protobuf/proto"
)

type Encoder struct {
	// Indicates whether the unknown fields must be silently skipped.
	Relaxed bool
	buf     proto.Buffer
}

var BadMessageErr = errors.New("bad message")
var RepeatedNullErr = errors.New("repeated field has null item")

// -----------------------------------------------------------------------------
// Encoding

func (s *Encoder) Encode(e *Entity, pd *ProtoDef) ([]byte, error) {
	s.buf.Reset()
	for _, f := range pd.Fields {
		var err error
		if f.Repeated {
			if f.DataType.isRef() {
				err = s.encodeRefs(e, pd, f)
			} else {
				err = s.encodeValues(e, f)
			}
		} else if f.DataType.isRef() {
			item := e.Entities[f.index]
			if item != nil {
				err = s.encodeRef(item, pd, f)
			}
		} else {
			value := f.GetValue(e)
			err = s.encodeValue(uint64(value), f)
		}
		if err != nil {
			s.buf.Reset()
			return nil, err
		}
	}
	return s.buf.Bytes(), nil
}

func (s *Encoder) encodeValue(value uint64, f *ProtoField) (err error) {
	switch f.DataType {
	case DtFloat32:
		if err = s.encodeTag(f.Tag, proto.WireFixed32); err == nil {
			err = s.buf.EncodeFixed32(value)
		}
	case DtFloat64:
		if err = s.encodeTag(f.Tag, proto.WireFixed64); err == nil {
			err = s.buf.EncodeFixed64(value)
		}
	default:
		if err = s.encodeTag(f.Tag, proto.WireVarint); err == nil {
			err = s.buf.EncodeVarint(value)
		}
	}
	return
}

func (s *Encoder) encodeValues(e *Entity, f *ProtoField) (err error) {
	data := e.Entities[f.index]
	if data != nil {
		n := len(data.Data) / f.DataType.getSize()
		for i := 0; i < n && err == nil; i++ {
			var value Primitive
			if value, err = f.GetValueAt(e, i); err == nil {
				err = s.encodeValue(uint64(value), f)
			}
		}
	}
	return
}

func (s *Encoder) encodeRef(
	e *Entity, pd *ProtoDef, f *ProtoField) (err error) {
	var bytes []byte
	if f.DataType == DtBytes || f.DataType == DtString {
		bytes = e.Data
	} else {
		def := pd.repo.defs[f.DataType&^DtEntity]
		enc := Encoder{}
		if bytes, err = enc.Encode(e, def); err != nil {
			return
		}
	}
	if err = s.encodeTag(f.Tag, proto.WireBytes); err == nil {
		err = s.buf.EncodeRawBytes(bytes)
	}
	return
}

func (s *Encoder) encodeRefs(e *Entity, pd *ProtoDef, f *ProtoField) error {
	data := e.Entities[f.index]
	if data == nil {
		return nil
	}
	for _, item := range data.Entities {
		if item == nil {
			return RepeatedNullErr
		}
		if err := s.encodeRef(item, pd, f); err != nil {
			return err
		}
	}
	return nil
}

func (s *Encoder) encodeTag(tag, wire uint64) error {
	return s.buf.EncodeVarint(uint64((tag << 3) | wire))
}

// -----------------------------------------------------------------------------
// Decoding

func (s *Encoder) Decode(b []byte, pd *ProtoDef) (*Entity, error) {
	s.buf.SetBuf(b)
	e := pd.NewEntity()
	for !s.buf.Eob() {
		t, err := s.buf.DecodeVarint()
		if err != nil {
			return nil, err
		}
		wire, tag := t&7, t>>3
		f, ok := pd.Fields[tag]
		if !ok && !s.Relaxed {
			return nil, BadMessageErr
		}
		if wire == proto.WireBytes {
			err = s.decodeRef(e, pd, f)
		} else {
			err = s.decodeValue(e, wire, f)
		}
		if err != nil {
			return nil, err
		}
	}
	return e, nil
}

func (s *Encoder) decodeRef(e *Entity, pd *ProtoDef, f *ProtoField) error {
	value, err := s.buf.DecodeRawBytes(false)
	if err != nil {
		return err
	}
	if f == nil {
		// The field has not been found, but if we've managed to reach this point,
		// it doesn't matter, so returning without an error.
		return nil
	}
	if !f.DataType.isRef() {
		return decodePacked(e, f, value)
	}
	var entity *Entity
	if (f.DataType & DtEntity) != 0 {
		def := pd.repo.defs[f.DataType&^DtEntity]
		enc := Encoder{
			Relaxed: s.Relaxed,
		}
		entity, err = enc.Decode(value, def)
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
		data := e.Entities[f.index]
		if data == nil {
			data = new(Entity)
			e.Entities[f.index] = data
		}
		data.Entities = append(data.Entities, entity)
	} else {
		e.Entities[f.index] = entity
	}
	return nil
}

func (s *Encoder) decodeValue(e *Entity, wire uint64, f *ProtoField) error {
	var value uint64
	var err error
	switch wire {
	case proto.WireVarint:
		value, err = s.buf.DecodeVarint()
	case proto.WireFixed32:
		value, err = s.buf.DecodeFixed32()
	case proto.WireFixed64:
		value, err = s.buf.DecodeFixed64()
	default:
		return BadMessageErr
	}
	if f == nil {
		// The field has not been found, but if we've managed to reach this point,
		// it doesn't matter, so returning without an error.
		return nil
	}
	if err == nil {
		if f.Repeated {
			entity := e.Entities[f.index]
			if entity == nil {
				entity = &Entity{}
				e.Entities[f.index] = entity
			}
			n := f.Reserve(entity, 1)
			err = f.SetValueAt(entity, n, Primitive(value))
			if err != nil {
				panic(err)
			}
		} else {
			f.SetValue(e, Primitive(value))
		}
	}
	return err
}

func decodePacked(e *Entity, f *ProtoField, value []byte) error {
	buf := proto.NewBuffer(value)
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
		f.SetValueAt(e, n, Primitive(i))
	}
	return nil
}
