package dymessage

import (
	"math"
)

// A generic representation of the primitive values that provides methods for
// converting the value to any primitive type.
type Primitive uint64

// Attaches the conversion methods to an entity.
type Reference Entity

// -----------------------------------------------------------------------------
// Primitive value conversions

func FromInt32(value int32) Primitive   { return Primitive(value) }
func FromInt64(value int64) Primitive   { return Primitive(value) }
func FromUint32(value uint32) Primitive { return Primitive(value) }
func FromUint64(value uint64) Primitive { return Primitive(value) }

func FromFloat32(value float32) Primitive { return Primitive(math.Float32bits(value)) }
func FromFloat64(value float64) Primitive { return Primitive(math.Float64bits(value)) }

func FromBool(value bool) Primitive {
	if value {
		return 1
	} else {
		return 0
	}
}

func (p Primitive) ToInt32() int32   { return int32(p) }
func (p Primitive) ToInt64() int64   { return int64(p) }
func (p Primitive) ToUint32() uint32 { return uint32(p) }
func (p Primitive) ToUint64() uint64 { return uint64(p) }

func (p Primitive) ToFloat32() float32 { return math.Float32frombits(uint32(p)) }
func (p Primitive) ToFloat64() float64 { return math.Float64frombits(uint64(p)) }

func (p Primitive) ToBool() bool { return p != 0 }

// -----------------------------------------------------------------------------
// Reference value conversions

func FromEntity(value *Entity) *Reference {
	return (*Reference)(value)
}

func FromString(value string) *Reference {
	return (*Reference)(&Entity{Data: ([]byte)(value)})
}

func FromBytes(value []byte, clone bool) *Reference {
	data := value
	if clone {
		data = make([]byte, len(value))
		copy(data, value)
	}
	return (*Reference)(&Entity{Data: data})
}

func (r *Reference) ToEntity() *Entity { return (*Entity)(r) }
func (r *Reference) ToString() string  { return string(r.ToBytes()) }

func (r *Reference) ToBytes() []byte {
	if r == nil {
		return nil
	}
	return r.Data
}
