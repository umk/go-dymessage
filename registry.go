package dymessage

import "github.com/umk/go-dymessage/internal/impl"

type (
	// Represents a collection of message definitions. The messages
	// defined in the registry may refer only these messages, which are
	// also defined in the same registry.
	Registry struct{ *impl.Registry }

	// Represents a definition of the message structure.
	MessageDef struct{ *impl.MessageDef }

	// Represents a single field of a message.
	MessageFieldDef struct{ *impl.MessageFieldDef }
)

func (md MessageDef) NewEntity() Entity {
	entity := &impl.Entity{
		Data:     make([]byte, md.DataBufLength),
		Entities: make([]*impl.Entity, md.EntityBufLength),
	}
	return Entity{entity}
}
