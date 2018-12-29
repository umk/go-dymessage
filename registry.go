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
		// A collection of fields that belong to the message.
		Fields map[uint64]*MessageFieldDef
		// Number of bytes taken by primitive values. These doesn't
		// include the repeated values, which are represented by a
		// separate entity.
		DataBufLength int
		// Number of entities referenced by the root. The collections
		// of entities and repeated primitive values are represented
		// by a single entity.
		EntityBufLength int
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
)

func (md MessageDef) NewEntity() *Entity {
	return &Entity{
		Data:     make([]byte, md.DataBufLength),
		Entities: make([]*Entity, md.EntityBufLength),
	}
}
