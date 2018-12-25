package dymessage

type (
	// Represents a collection of message definitions. The messages
	// defined in the registry may refer only these messages, which are
	// also defined in the same registry.
	Registry struct {
		// A collection of message definitions at the positions by
		// which these definitions are referenced from other ones and
		// outside.
		Defs []*MessageDef
	}

	// Represents a definition of the message structure.
	MessageDef struct {
		Namespace string
		Name      string
		// A registry this definition belongs to.
		Registry *Registry
		// Number of bytes taken by primitive values. These doesn't
		// include the repeated values, which are represented by a
		// separate entity.
		dataLength int
		// Number of entities referenced by the root. The collections
		// of entities and repeated primitive values are represented
		// by a single entity.
		entitiesLength int
		// A collection of fields that belong to the message.
		Fields map[uint64]*MessageFieldDef
	}

	// Represents a single field of a message.
	MessageFieldDef struct {
		Name     string
		DataType DataType
		Tag      uint64
		Repeated bool
		// Offset of the field in the array of bytes if the field is of
		// a primitive type and not repeated. Elsewhere, an index in the
		// array of entities.
		Offset int
	}

	// The message field data type.
	DataType uint32
)

// -----------------------------------------------------------------------------
// Data types

const (
	DtInt32 DataType = iota + 1
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
	typeSize8  = 1
	typeSize32 = 4
	typeSize64 = 8
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

// GetSizeInBytes returns the value indicating how many bytes of memory does
// the type require. This method is only valid for primitive types.
func (dt DataType) GetSizeInBytes() int {
	if dt.IsRefType() {
		panic("operation is valid only for primitive types")
	}
	switch dt {
	case DtInt32, DtUint32, DtFloat32:
		return typeSize32
	case DtInt64, DtUint64, DtFloat64:
		return typeSize64
	case DtBool:
		return typeSize8
	default:
		panic(dt)
	}
}
