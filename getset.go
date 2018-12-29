package dymessage

import (
	"errors"
	"fmt"

	"github.com/umk/go-dymessage/internal/impl"
	"github.com/umk/go-dymessage/types"
	"github.com/umk/go-memutil"
)

var ErrIndexOutOfRange = errors.New("index is out of range")

func (f MessageFieldDef) GetValue(e Entity) Primitive {
	data, _ := f.getData(e.Entity, f.Offset)
	return Primitive(data)
}

func (f MessageFieldDef) GetValueAt(e Entity, n int) (Primitive, error) {
	data := e.Entities[f.Offset]
	if data == nil {
		return GetDefaultPrimitive(), ErrIndexOutOfRange
	} else {
		return f.getData(data, f.DataType.GetWidthInBytes()*n)
	}
}

func (f MessageFieldDef) SetValue(e Entity, value Primitive) {
	if err := f.setData(e.Entity, f.Offset, value); err != nil {
		panic(err)
	}
}

func (f MessageFieldDef) SetValueAt(e Entity, n int, value Primitive) error {
	data := e.Entities[f.Offset]
	if data == nil {
		return ErrIndexOutOfRange
	} else {
		return f.setData(data, f.DataType.GetWidthInBytes()*n, value)
	}
}

func (f MessageFieldDef) GetEntity(e Entity) Reference {
	return Reference{e.Entities[f.Offset]}
}

func (f MessageFieldDef) GetEntityAt(e Entity, n int) (Reference, error) {
	data := e.Entities[f.Offset]
	if data == nil || len(data.Entities) <= n {
		return GetDefaultReference(), ErrIndexOutOfRange
	} else {
		return Reference{data.Entities[n]}, nil
	}
}

func (f MessageFieldDef) SetEntity(e Entity, value Reference) {
	e.Entities[f.Offset] = value.Entity
}

func (f MessageFieldDef) SetEntityAt(e Entity, n int, value Reference) error {
	data := e.Entities[f.Offset]
	if data == nil || len(data.Entities) <= n {
		return ErrIndexOutOfRange
	} else {
		data.Entities[n] = value.Entity
		return nil
	}
}

func (f MessageFieldDef) Reserve(e Entity, count int) int {
	data := e.Entities[f.Offset]
	if data == nil {
		data = new(impl.Entity)
		e.Entities[f.Offset] = data
	}
	if f.DataType.IsRefType() {
		n := len(data.Entities)
		data.Entities = append(data.Entities, make([]*impl.Entity, count)...)
		return n
	} else {
		sz := f.DataType.GetWidthInBytes()
		n := len(data.Data) / sz
		data.Data = append(data.Data, make([]byte, count*sz)...)
		return n
	}
}

func (f MessageFieldDef) Len(e Entity) int {
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

func (f MessageFieldDef) getData(e *impl.Entity, off int) (Primitive, error) {
	sz := f.DataType.GetWidthInBytes()
	if off+sz > len(e.Data) {
		return 0, ErrIndexOutOfRange
	}
	var value uint64
	switch sz {
	case types.TypeWidth8:
		value = uint64(e.Data[off])
	case types.TypeWidth32:
		v32 := memutil.GetByteOrder().Uint32(e.Data[off : off+4])
		value = uint64(v32)
	case types.TypeWidth64:
		value = memutil.GetByteOrder().Uint64(e.Data[off : off+8])
	default:
		panic(fmt.Sprintf("unexpected size of the field: %v", sz))
	}
	return Primitive(value), nil
}

func (f MessageFieldDef) setData(e *impl.Entity, off int, value Primitive) error {
	sz := f.DataType.GetWidthInBytes()
	if off+sz > len(e.Data) {
		return ErrIndexOutOfRange
	}
	switch sz {
	case types.TypeWidth8:
		e.Data[off] = byte(value)
	case types.TypeWidth32:
		memutil.GetByteOrder().PutUint32(e.Data[off:off+4], uint32(value))
	case types.TypeWidth64:
		memutil.GetByteOrder().PutUint64(e.Data[off:off+8], uint64(value))
	default:
		panic(fmt.Sprintf("unexpected size of the field: %v", sz))
	}
	return nil
}
