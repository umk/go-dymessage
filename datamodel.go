package dymessage

import (
	"math"

	"github.com/umk/go-dymessage/internal/impl"
)

type (
	Entity struct{ *impl.Entity }

	// A generic representation of the primitive values that provides
	// methods for converting the value to any native primitive type. The
	// provided methods do not keep track of the correct usage, meaning that
	// the user must correlate between the methods and the value types.
	Primitive uint64

	// A generic representation of the reference values that provides
	// methods for converting the value to any native reference type. The
	// provided methods do not keep track of the correct usage, meaning that
	// the user must correlate between the methods and the value types.
	Reference struct{ *impl.Entity }
)

// -----------------------------------------------------------------------------
// Type defaults

// GetDefaultPrimitive gets a default primitive value, which will evaluate to
// zero for all numeric types or false for boolean.
func GetDefaultPrimitive() Primitive { return Primitive(0) }

// GetDefaultReference gets a default reference value, which doesn't contain any
// data and will evaluate to nil for the reference native types or an empty
// string for string type.
func GetDefaultReference() Reference { return Reference{Entity: nil} }

// GetDefaultEntity gets a default entity value, which corresponds to nil.
func GetDefaultEntity() Entity { return Entity{Entity: nil} }

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

func FromEntity(value *Entity) Reference {
	return Reference{value.Entity}
}

func FromString(value string) Reference {
	return Reference{&impl.Entity{Data: ([]byte)(value)}}
}

func FromBytes(value []byte, clone bool) Reference {
	data := value
	if clone {
		data = make([]byte, len(value))
		copy(data, value)
	}
	return Reference{&impl.Entity{Data: data}}
}

func (r Reference) ToEntity() Entity {
	return Entity{r.Entity}
}

func (r Reference) ToString() string {
	return string(r.ToBytes())
}

func (r Reference) ToBytes() []byte {
	if r.Entity == nil {
		return nil
	}
	return r.Entity.Data
}
