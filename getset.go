package dymessage

import (
	"errors"
	"fmt"

	"github.com/umk/go-memutil"
)

var IndexOutOfRangeErr = errors.New("index is out of range")

var primitiveDefault = Primitive(0)

func (f *ProtoField) GetValue(e *Entity) Primitive {
	data, _ := f.getData(e, f.index)
	return Primitive(data)
}

func (f *ProtoField) GetValueAt(e *Entity, n int) (Primitive, error) {
	data := e.Entities[f.index]
	if data == nil {
		return primitiveDefault, IndexOutOfRangeErr
	} else {
		return f.getData(data, f.DataType.getSize()*n)
	}
}

func (f *ProtoField) SetValue(e *Entity, value Primitive) {
	f.setData(e, f.index, value)
}

func (f *ProtoField) SetValueAt(e *Entity, n int, value Primitive) error {
	data := e.Entities[f.index]
	if data == nil {
		return IndexOutOfRangeErr
	} else {
		return f.setData(data, f.DataType.getSize()*n, value)
	}
}

func (f *ProtoField) GetEntity(e *Entity) *Reference {
	return (*Reference)(e.Entities[f.index])
}

func (f *ProtoField) GetEntityAt(e *Entity, n int) (*Reference, error) {
	data := e.Entities[f.index]
	if data == nil || len(data.Entities) <= n {
		return nil, IndexOutOfRangeErr
	} else {
		return (*Reference)(data.Entities[n]), nil
	}
}

func (f *ProtoField) SetEntity(e *Entity, value *Reference) {
	e.Entities[f.index] = (*Entity)(value)
}

func (f *ProtoField) SetEntityAt(e *Entity, n int, value *Reference) error {
	data := e.Entities[f.index]
	if data == nil || len(data.Entities) <= n {
		return IndexOutOfRangeErr
	} else {
		data.Entities[n] = (*Entity)(value)
		return nil
	}
}

func (f *ProtoField) Reserve(e *Entity, count int) int {
	data := e.Entities[f.index]
	if data == nil {
		data = new(Entity)
		e.Entities[f.index] = data
	}
	if f.DataType.isRef() {
		n := len(data.Entities)
		data.Entities = append(data.Entities, make([]*Entity, count)...)
		return n
	} else {
		sz := f.DataType.getSize()
		n := len(data.Data) / sz
		data.Data = append(data.Data, make([]byte, count*sz)...)
		return n
	}
}

func (f *ProtoField) Len(e *Entity) int {
	data := e.Entities[f.index]
	if data == nil {
		return 0
	}
	if f.DataType.isRef() {
		return len(data.Entities)
	} else {
		return len(data.Data) / f.DataType.getSize()
	}
}

func (f *ProtoField) getData(e *Entity, off int) (Primitive, error) {
	sz := f.DataType.getSize()
	if off+sz > len(e.Data) {
		return 0, IndexOutOfRangeErr
	}
	var value uint64
	switch sz {
	case bytes8:
		value = uint64(e.Data[off])
	case bytes32:
		v32 := memutil.GetByteOrder().Uint32(e.Data[off : off+4])
		value = uint64(v32)
	case bytes64:
		value = memutil.GetByteOrder().Uint64(e.Data[off : off+8])
	default:
		panic(fmt.Sprintf("unexpected size of the field: %v", sz))
	}
	return Primitive(value), nil
}

func (f *ProtoField) setData(e *Entity, off int, value Primitive) error {
	sz := f.DataType.getSize()
	if off+sz > len(e.Data) {
		return IndexOutOfRangeErr
	}
	switch sz {
	case bytes8:
		e.Data[off] = byte(value)
	case bytes32:
		memutil.GetByteOrder().PutUint32(e.Data[off:off+4], uint32(value))
	case bytes64:
		memutil.GetByteOrder().PutUint64(e.Data[off:off+8], uint64(value))
	default:
		panic(fmt.Sprintf("unexpected size of the field: %v", sz))
	}
	return nil
}
