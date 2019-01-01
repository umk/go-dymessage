package protobuf

import (
	"fmt"
	. "github.com/umk/go-dymessage"
	. "github.com/umk/go-dymessage/protobuf/internal/impl"
)

func (ec *Encoder) Encode(e *Entity, pd *MessageDef) ([]byte, error) {
	defer ec.buf.Reset()
	for _, f := range pd.Fields {
		var err error
		if f.Repeated {
			if f.DataType.IsRefType() {
				err = ec.encodeRefs(e, pd, f)
			} else {
				err = ec.encodeValues(e, f)
			}
		} else if f.DataType.IsRefType() {
			item := e.Entities[f.Offset]
			if item != nil {
				err = ec.encodeRef(item, pd, f)
			}
		} else {
			value := f.GetPrimitive(e)
			err = ec.encodeValue(uint64(value), f)
		}
		if err != nil {
			return nil, err
		}
	}
	return ec.buf.Bytes(), nil
}

func (ec *Encoder) encodeValue(value uint64, f *MessageFieldDef) (err error) {
	extension, ok := tryGetExtension(f)
	if ok && extension.integerKind != ikDefault {
		ik := extension.integerKind
		return ec.encodeValueByKind(value, f, ik)
	}
	switch f.DataType {
	case DtInt32:
	case DtUint32:
	case DtFloat32:
		if err = ec.encodeTag(f.Tag, WireFixed32); err == nil {
			err = ec.buf.EncodeFixed32(value)
		}
	case DtInt64:
	case DtUint64:
	case DtFloat64:
		if err = ec.encodeTag(f.Tag, WireFixed64); err == nil {
			err = ec.buf.EncodeFixed64(value)
		}
	case DtBool:
		if err = ec.encodeTag(f.Tag, WireVarint); err == nil {
			err = ec.buf.EncodeVarint(value)
		}
	default:
		panic(fmt.Sprintf("unsupported encoding data type %d", f.DataType))
	}
	return
}

func (ec *Encoder) encodeValueByKind(
	value uint64, f *MessageFieldDef, ik integerKind) (err error) {
	switch ik {
	case ikVarint:
		if err = ec.encodeTag(f.Tag, WireVarint); err == nil {
			err = ec.buf.EncodeVarint(value)
		}
	case ikZigZag:
		if err = ec.encodeTag(f.Tag, WireVarint); err == nil {
			switch f.DataType {
			case DtInt32:
				err = ec.buf.EncodeZigzag32(value)
			case DtInt64:
				err = ec.buf.EncodeZigzag64(value)
			default:
				panic(fmt.Sprintf("ZigZag encoding is applied to invalid data type %d", f.DataType))
			}
		}
	default:
		panic(fmt.Sprintf("unsupported value of integer kind %d", ik))
	}
	return
}

func (ec *Encoder) encodeValues(e *Entity, f *MessageFieldDef) (err error) {
	data := e.Entities[f.Offset]
	if data != nil {
		n := len(data.Data) / f.DataType.GetWidthInBytes()
		for i := 0; i < n && err == nil; i++ {
			value := f.GetPrimitiveAt(e, i)
			err = ec.encodeValue(uint64(value), f)
		}
	}
	return
}

func (ec *Encoder) encodeRef(
	e *Entity, pd *MessageDef, f *MessageFieldDef) (err error) {
	var bytes []byte
	if f.DataType == DtBytes || f.DataType == DtString {
		bytes = e.Data
	} else {
		def := pd.Registry.GetMessageDef(f.DataType)
		another := ec.clone()
		if bytes, err = another.Encode(e, def); err != nil {
			return
		}
	}
	if err = ec.encodeTag(f.Tag, WireBytes); err == nil {
		err = ec.buf.EncodeRawBytes(bytes)
	}
	return
}

func (ec *Encoder) encodeRefs(e *Entity, pd *MessageDef, f *MessageFieldDef) error {
	data := e.Entities[f.Offset]
	if data == nil {
		return nil
	}
	for _, item := range data.Entities {
		if item == nil {
			return ErrRepeatedNull
		}
		if err := ec.encodeRef(item, pd, f); err != nil {
			return err
		}
	}
	return nil
}

func (ec *Encoder) encodeTag(tag, wire uint64) error {
	return ec.buf.EncodeVarint(uint64((tag << 3) | wire))
}
