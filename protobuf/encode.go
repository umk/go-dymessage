package protobuf

import (
	. "github.com/umk/go-dymessage"
	. "github.com/umk/go-dymessage/protobuf/internal/impl"
)

func (s *Encoder) Encode(e *Entity, pd *MessageDef) ([]byte, error) {
	s.buf.Reset()
	for _, f := range pd.Fields {
		var err error
		if f.Repeated {
			if f.DataType.IsRefType() {
				err = s.encodeRefs(e, pd, f)
			} else {
				err = s.encodeValues(e, f)
			}
		} else if f.DataType.IsRefType() {
			item := e.Entities[f.Offset]
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

func (s *Encoder) encodeValue(value uint64, f *MessageFieldDef) (err error) {
	switch f.DataType {
	case DtFloat32:
		if err = s.encodeTag(f.Tag, WireFixed32); err == nil {
			err = s.buf.EncodeFixed32(value)
		}
	case DtFloat64:
		if err = s.encodeTag(f.Tag, WireFixed64); err == nil {
			err = s.buf.EncodeFixed64(value)
		}
	default:
		if err = s.encodeTag(f.Tag, WireVarint); err == nil {
			err = s.buf.EncodeVarint(value)
		}
	}
	return
}

func (s *Encoder) encodeValues(e *Entity, f *MessageFieldDef) (err error) {
	data := e.Entities[f.Offset]
	if data != nil {
		n := len(data.Data) / f.DataType.GetSizeInBytes()
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
	e *Entity, pd *MessageDef, f *MessageFieldDef) (err error) {
	var bytes []byte
	if f.DataType == DtBytes || f.DataType == DtString {
		bytes = e.Data
	} else {
		def := pd.Registry.Defs[f.DataType&^DtEntity]
		enc := Encoder{}
		if bytes, err = enc.Encode(e, def); err != nil {
			return
		}
	}
	if err = s.encodeTag(f.Tag, WireBytes); err == nil {
		err = s.buf.EncodeRawBytes(bytes)
	}
	return
}

func (s *Encoder) encodeRefs(e *Entity, pd *MessageDef, f *MessageFieldDef) error {
	data := e.Entities[f.Offset]
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
