package dymessage

import "fmt"

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
		Namespace string // An optional namespace of the message definition
		Name      string // Name of the message definition

		Registry *Registry // A registry this definition belongs to
		DataType DataType  // An entity data type represented by this instance

		// A collection of fields that belong to the message.
		Fields []*MessageFieldDef

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
		// A collection of extensions which alter the serialization and
		// deserialization behavior of current field.
		Extensions

		Name     string   // A name of the field unique in bounds of the message definition
		DataType DataType // Data type of the message field
		Tag      uint64   // A tag unique in bounds of the message definition
		Repeated bool     // Indicates whether the field contains a collection of items

		// Offset of the field in the array of bytes if the field is of
		// a primitive type and not repeated. Elsewhere, an index in the
		// array of entities.
		Offset int
	}
)

// -----------------------------------------------------------------------------
// Implementation

// GetMessageDef gets the message definition by its data type.
func (r *Registry) GetMessageDef(dt DataType) *MessageDef {
	id, n := int(dt&^DtEntity), len(r.Defs)
	if id >= n {
		message := fmt.Sprintf(
			"expected message definition at %d, but got only %d definitions", id, n)
		panic(message)
	}
	return r.Defs[id]
}

// NewEntity creates a new entity with all of the buffers reserved to store the
// primitive and reference fields of the entity.
func (md *MessageDef) NewEntity() *Entity {
	return &Entity{
		DataType: md.DataType,
		Data:     make([]byte, md.DataBufLength),
		Entities: make([]*Entity, md.EntityBufLength),
	}
}

// TryGetField gets the field with specified tag from the message definition. If
// field doesn't exist, it returns the false flag.
func (md *MessageDef) TryGetField(tag uint64) (*MessageFieldDef, bool) {
	// For the small number of fields (up to ~30, which must be the majority
	// of the cases) the brute-force search is more effective than using a
	// map.
	for _, def := range md.Fields {
		if def.Tag == tag {
			return def, true
		}
	}
	return nil, false
}

// GetField gets the field with specified tag from the message definition. If
// field doesn't exist, the method panics.
func (md *MessageDef) GetField(tag uint64) *MessageFieldDef {
	if def, ok := md.TryGetField(tag); ok {
		return def
	}
	panic(fmt.Sprintf("entity doesn't contain the field with tag %d", tag))
}

// TryGetFieldByName gets the field with specified name from the message
// definition. If field doesn't exist, it returns the false flag.
func (md *MessageDef) TryGetFieldByName(name string) (*MessageFieldDef, bool) {
	for _, def := range md.Fields {
		if def.Name == name {
			return def, true
		}
	}
	return nil, false
}

// GetFieldByName gets the field with specified name from the message
// definition. If field doesn't exist, the method panics.
func (md *MessageDef) GetFieldByName(name string) *MessageFieldDef {
	if def, ok := md.TryGetFieldByName(name); ok {
		return def
	}
	panic(fmt.Sprintf("entity doesn't contain the field with name %q", name))
}
