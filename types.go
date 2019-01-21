package dymessage

// The data type of a dynamic message field.
type DataType uint32

const (
	DtNone DataType = iota
	DtInt32
	DtInt64
	DtUint32
	DtUint64
	DtFloat32
	DtFloat64
	DtBool
	DtString
	DtBytes
	DtEntity DataType = 1 << 31
)

const (
	TypeWidth8  = 1
	TypeWidth32 = 4
	TypeWidth64 = 8
)

// IsRefType gets a value indicating whether the type represents a reference
// type rather than a primitive type.
func (dt DataType) IsRefType() bool {
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

// IsEntity gets a value indicating whether the data type refers to an entity.
func (dt DataType) IsEntity() bool { return (dt & DtEntity) != 0 }

// GetWidthInBytes returns the value indicating how many bytes of memory does
// the type require. This method is only valid for primitive types.
func (dt DataType) GetWidthInBytes() int {
	if dt.IsRefType() {
		panic("operation is valid only for primitive types")
	}
	switch dt {
	case DtInt32, DtUint32, DtFloat32:
		return TypeWidth32
	case DtInt64, DtUint64, DtFloat64:
		return TypeWidth64
	case DtBool:
		return TypeWidth8
	default:
		panic(dt)
	}
}
