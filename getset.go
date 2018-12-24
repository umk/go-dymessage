package dymessage

import (
	"errors"
	"fmt"

	"github.com/umk/go-memutil"
)

var IndexOutOfRangeErr = errors.New("index is out of range")

// The default value of a primitive.
var PrimitiveDefault = Primitive(0)

func (f *MessageFieldDef) GetValue(e *Entity) Primitive {
	data, _ := f.getData(e, f.Offset)
	return Primitive(data)
}

func (f *MessageFieldDef) GetValueAt(e *Entity, n int) (Primitive, error) {
	data := e.Entities[f.Offset]
	if data == nil {
		return PrimitiveDefault, IndexOutOfRangeErr
	} else {
		return f.getData(data, f.DataType.GetSizeInBytes()*n)
	}
}

func (f *MessageFieldDef) SetValue(e *Entity, value Primitive) {
	if err := f.setData(e, f.Offset, value); err != nil {
		panic(err)
	}
}

func (f *MessageFieldDef) SetValueAt(e *Entity, n int, value Primitive) error {
	data := e.Entities[f.Offset]
	if data == nil {
		return IndexOutOfRangeErr
	} else {
		return f.setData(data, f.DataType.GetSizeInBytes()*n, value)
	}
}

func (f *MessageFieldDef) GetEntity(e *Entity) *Reference {
	return (*Reference)(e.Entities[f.Offset])
}

func (f *MessageFieldDef) GetEntityAt(e *Entity, n int) (*Reference, error) {
	data := e.Entities[f.Offset]
	if data == nil || len(data.Entities) <= n {
		return nil, IndexOutOfRangeErr
	} else {
		return (*Reference)(data.Entities[n]), nil
	}
}

func (f *MessageFieldDef) SetEntity(e *Entity, value *Reference) {
	e.Entities[f.Offset] = (*Entity)(value)
}

func (f *MessageFieldDef) SetEntityAt(e *Entity, n int, value *Reference) error {
	data := e.Entities[f.Offset]
	if data == nil || len(data.Entities) <= n {
		return IndexOutOfRangeErr
	} else {
		data.Entities[n] = (*Entity)(value)
		return nil
	}
}

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
		sz := f.DataType.GetSizeInBytes()
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
		return len(data.Data) / f.DataType.GetSizeInBytes()
	}
}

func (f *MessageFieldDef) getData(e *Entity, off int) (Primitive, error) {
	sz := f.DataType.GetSizeInBytes()
	if off+sz > len(e.Data) {
		return 0, IndexOutOfRangeErr
	}
	var value uint64
	switch sz {
	case typeSize8:
		value = uint64(e.Data[off])
	case typeSize32:
		v32 := memutil.GetByteOrder().Uint32(e.Data[off : off+4])
		value = uint64(v32)
	case typeSize64:
		value = memutil.GetByteOrder().Uint64(e.Data[off : off+8])
	default:
		panic(fmt.Sprintf("unexpected size of the field: %v", sz))
	}
	return Primitive(value), nil
}

func (f *MessageFieldDef) setData(e *Entity, off int, value Primitive) error {
	sz := f.DataType.GetSizeInBytes()
	if off+sz > len(e.Data) {
		return IndexOutOfRangeErr
	}
	switch sz {
	case typeSize8:
		e.Data[off] = byte(value)
	case typeSize32:
		memutil.GetByteOrder().PutUint32(e.Data[off:off+4], uint32(value))
	case typeSize64:
		memutil.GetByteOrder().PutUint64(e.Data[off:off+8], uint64(value))
	default:
		panic(fmt.Sprintf("unexpected size of the field: %v", sz))
	}
	return nil
}
