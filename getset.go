package dymessage

import (
	"errors"
	"fmt"

	"github.com/umk/go-memutil"
)

var ErrIndexOutOfRange = errors.New("index is out of range")

func (f *MessageFieldDef) GetPrimitive(e *Entity) Primitive {
	data, _ := f.getPrimitive(e, f.Offset)
	return Primitive(data)
}

func (f *MessageFieldDef) GetPrimitiveAt(e *Entity, n int) (Primitive, error) {
	data := e.Entities[f.Offset]
	if data == nil {
		return GetDefaultPrimitive(), ErrIndexOutOfRange
	} else {
		return f.getPrimitive(data, f.DataType.GetWidthInBytes()*n)
	}
}

func (f *MessageFieldDef) SetPrimitive(e *Entity, value Primitive) {
	if err := f.setPrimitive(e, f.Offset, value); err != nil {
		panic(err)
	}
}

func (f *MessageFieldDef) SetPrimitiveAt(e *Entity, n int, value Primitive) error {
	data := e.Entities[f.Offset]
	if data == nil {
		return ErrIndexOutOfRange
	} else {
		return f.setPrimitive(data, f.DataType.GetWidthInBytes()*n, value)
	}
}

func (f *MessageFieldDef) GetReference(e *Entity) Reference {
	return Reference{e.Entities[f.Offset]}
}

func (f *MessageFieldDef) GetReferenceAt(e *Entity, n int) (Reference, error) {
	data := e.Entities[f.Offset]
	if data == nil || len(data.Entities) <= n {
		return GetDefaultReference(), ErrIndexOutOfRange
	} else {
		return Reference{data.Entities[n]}, nil
	}
}

func (f *MessageFieldDef) SetReference(e *Entity, value Reference) {
	e.Entities[f.Offset] = value.Entity
}

func (f *MessageFieldDef) SetReferenceAt(e *Entity, n int, value Reference) error {
	data := e.Entities[f.Offset]
	if data == nil || len(data.Entities) <= n {
		return ErrIndexOutOfRange
	} else {
		data.Entities[n] = value.Entity
		return nil
	}
}

// Reserve reserves a room for specified number of items for the repeated
// message field and returns the number of items that have been allocated in the
// collection before a place for the new ones has been reserved.
func (f *MessageFieldDef) Reserve(e *Entity, count int) int {
	data := e.Entities[f.Offset]
	if data == nil {
		data = new(Entity)
		e.Entities[f.Offset] = data
	}
	if f.DataType.IsRefType() {
		n := len(data.Entities)
		data.Entities = append(data.Entities, make([]*Entity, count)...)
		return n
	} else {
		sz := f.DataType.GetWidthInBytes()
		n := len(data.Data) / sz
		data.Data = append(data.Data, make([]byte, count*sz)...)
		return n
	}
}

func (f *MessageFieldDef) Len(e *Entity) int {
	data := e.Entities[f.Offset]
	if data == nil {
		return 0
	}
	if f.DataType.IsRefType() {
		return len(data.Entities)
	} else {
		return len(data.Data) / f.DataType.GetWidthInBytes()
	}
}

func (f *MessageFieldDef) getPrimitive(e *Entity, off int) (Primitive, error) {
	sz := f.DataType.GetWidthInBytes()
	if off+sz > len(e.Data) {
		return 0, ErrIndexOutOfRange
	}
	var value uint64
	switch sz {
	case TypeWidth8:
		value = uint64(e.Data[off])
	case TypeWidth32:
		v32 := memutil.GetByteOrder().Uint32(e.Data[off : off+4])
		value = uint64(v32)
	case TypeWidth64:
		value = memutil.GetByteOrder().Uint64(e.Data[off : off+8])
	default:
		panic(fmt.Sprintf("unexpected size of the field: %v", sz))
	}
	return Primitive(value), nil
}

func (f *MessageFieldDef) setPrimitive(e *Entity, off int, value Primitive) error {
	sz := f.DataType.GetWidthInBytes()
	if off+sz > len(e.Data) {
		return ErrIndexOutOfRange
	}
	switch sz {
	case TypeWidth8:
		e.Data[off] = byte(value)
	case TypeWidth32:
		memutil.GetByteOrder().PutUint32(e.Data[off:off+4], uint32(value))
	case TypeWidth64:
		memutil.GetByteOrder().PutUint64(e.Data[off:off+8], uint64(value))
	default:
		panic(fmt.Sprintf("unexpected size of the field: %v", sz))
	}
	return nil
}
