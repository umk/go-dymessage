package helpers

import (
	. "github.com/umk/go-dymessage"
	"github.com/umk/go-dymessage/internal/impl"
)

const (
	TypeSize8  = 1
	TypeSize32 = 4
	TypeSize64 = 8
)

// IsRefType gets a value indicating whether the type represents a reference
// type rather than a primitive type.
func IsRefType(dt impl.DataType) bool {
	switch dt {
	case DtInt32, DtInt64, DtUint32, DtUint64, DtFloat32, DtFloat64, DtBool:
		return false
	case DtString, DtBytes:
		return true
	default:
		if (dt & DtEntity) != 0 {
			return true
		}
		panic(dt)
	}
}

// GetSizeInBytes returns the value indicating how many bytes of memory does
// the type require. This method is only valid for primitive types.
func GetSizeInBytes(dt impl.DataType) int {
	if IsRefType(dt) {
		panic("operation is valid only for primitive types")
	}
	switch dt {
	case DtInt32, DtUint32, DtFloat32:
		return TypeSize32
	case DtInt64, DtUint64, DtFloat64:
		return TypeSize64
	case DtBool:
		return TypeSize8
	default:
		panic(dt)
	}
}
