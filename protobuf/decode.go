package protobuf

import (
	"errors"
	"fmt"

	. "github.com/umk/go-dymessage"
	"github.com/umk/go-dymessage/internal/helpers"
	. "github.com/umk/go-dymessage/protobuf/internal/impl"
)

// Decode transforms the protocol buffers representation of the message to a
// dynamic entity against the provided message definition.
func (ec *Encoder) Decode(b []byte, pd *MessageDef) (*Entity, error) {
	return ec.DecodeInto(b, pd, pd.NewEntity())
}

// DecodeInto transforms the protocol buffers representation of the message to
// specified dynamic entity against the provided message definition. The
// returned entity is the one that has been provided as an input parameter e,
// but now populated with the data.
//
// If the entity type doesn't correspond the data type of the message
// definition, the method will panic.
func (ec *Encoder) DecodeInto(b []byte, pd *MessageDef, e *Entity) (*Entity, error) {
	helpers.DataTypesMustMatch(e, pd)
	defer ec.pushBuf(b)()
	// If entity data is not empty, resetting it to default just in case if
	// some of the fields are not populated.
	if len(e.Data) > 0 {
		for i := range e.Data {
			e.Data[i] = 0
		}
	}
	fseq, fields := 0, pd.Fields
	for !ec.buf.Eob() {
		t, err := ec.buf.DecodeVarint()
		if err != nil {
			return nil, err
		}
		wire, tag := t&7, t>>3
		f, ok := (*MessageFieldDef)(nil), false
		// Advancing fseq until it points the field with the current
		// tag, or goes outside of the fields slice length. While
		// enumerating, the nested entities get prepared for reuse.
		for fseq < len(fields) {
			fcur := fields[fseq]
			fseq++
			if fcur.Repeated || fcur.DataType == DtBytes || fcur.DataType == DtString {
				if ch := e.Entities[fcur.Offset]; ch != nil {
					ch.Reset()
				}
			}
			if fcur.Tag == tag {
				f = fcur
				goto FoundField
			}
			if (fcur.DataType & DtEntity) != 0 {
				// In case if the entity won't be provided at all.
				e.Entities[fcur.Offset] = nil
			}
		}
		// If the field count not be found, trying to find by looking
		// through all the collection of entity fields.
		f, ok = pd.TryGetField(tag)
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
	FoundField:
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
		return ec.decodeValuePacked(e, f, value)
	}

	// Getting an entity, which can be reused. This assumes that the nested
	// entities have already been prepared for this by shrinking the size of
	// collections to zero.
	var entity *Entity
	if f.Repeated {
		data := e.Entities[f.Offset]
		if data == nil {
			data = new(Entity)
			e.Entities[f.Offset] = data
		}
		// If the field is repeated and represented by a reference type,
		// which is true in this scope, a place for the new item of
		// collection is reserved before the item is retrieved in order
		// to make possible to reuse an existing item.
		n := len(data.Entities)
		if n < cap(data.Entities) {
			data.Entities = data.Entities[:n+1]
			entity = data.Entities[n]
		} else {
			data.Entities = append(data.Entities, nil)
		}
	} else {
		// For non-repeated fields trying to reuse the entity, which
		// represents the nested entity, byte array or string.
		entity = e.Entities[f.Offset]
	}
	// Populating the nested entity with the data from the buffer.
	if (f.DataType & DtEntity) != 0 {
		def := pd.Registry.GetMessageDef(f.DataType)
		if entity == nil {
			entity = def.NewEntity()
		}
		if entity, err = ec.DecodeInto(value, def, entity); err != nil {
			return err
		}
	} else {
		if entity == nil {
			entity = &Entity{}
		}
		// If capacity allows the data block of the value is reused in
		// order to store the binary data. Otherwise a new block is
		// created, and existing one is abandoned.
		n := len(value)
		if n <= cap(entity.Data) {
			entity.Data = entity.Data[0:n]
		} else {
			entity.Data = make([]byte, len(value))
		}
		copy(entity.Data, value)
	}
	// Updating the entity with a value built from the buffer.
	if f.Repeated {
		// The repeated fields already got the last item of the entities
		// slice reserved for the new one, so just assigning it.
		data := e.Entities[f.Offset]
		n := len(data.Entities)
		data.Entities[n-1] = entity
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
		value, err = ec.decodeValueByKind(f, extension.integerKind)
	} else {
		value, err = ec.decodeValueDefault(f)
	}
	if err == nil {
		if f.Repeated {
			entity := e.Entities[f.Offset]
			if entity == nil {
				entity = &Entity{}
				e.Entities[f.Offset] = entity
			}
			n := f.Reserve(e, 1)
			f.SetPrimitiveAt(e, n, Primitive(value))
		} else {
			f.SetPrimitive(e, Primitive(value))
		}
	}
	return err
}

func (ec *Encoder) decodeValuePacked(e *Entity, f *MessageFieldDef, value []byte) (err error) {
	defer ec.pushBuf(value)()
	extension, ok := tryGetExtension(f)
	for !ec.buf.Eob() {
		var i uint64
		if ok && extension.integerKind != ikDefault {
			i, err = ec.decodeValueByKind(f, extension.integerKind)
		} else {
			i, err = ec.decodeValueDefault(f)
		}
		if err != nil {
			return
		}
		prev := f.Reserve(e, 1)
		f.SetPrimitiveAt(e, prev, Primitive(i))
	}
	return
}

func (ec *Encoder) decodeValueDefault(f *MessageFieldDef) (value uint64, err error) {
	switch f.DataType {
	case DtInt32, DtUint32, DtFloat32:
		value, err = ec.buf.DecodeFixed32()
	case DtInt64, DtUint64, DtFloat64:
		value, err = ec.buf.DecodeFixed64()
	case DtBool:
		value, err = ec.buf.DecodeVarint()
	default:
		panic(fmt.Sprintf("unsupported decoding data type %d", f.DataType))
	}
	return
}

func (ec *Encoder) decodeValueByKind(f *MessageFieldDef, ik integerKind) (uint64, error) {
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
