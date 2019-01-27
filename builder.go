package dymessage

import (
	"fmt"
	"sort"
)

type (
	RegistryBuilder struct {
		defs     map[interface{}]*MessageDefBuilder
		registry *Registry
	}

	MessageDefBuilder struct {
		// Index of this message definition in the registry.
		index    int
		registry *Registry

		message *MessageDef      // Message definition being built by this builder
		field   *MessageFieldDef // Field definition added the last time
	}
)

// -----------------------------------------------------------------------------
// Registry builder

func NewRegistryBuilder() *RegistryBuilder {
	return &RegistryBuilder{
		defs:     make(map[interface{}]*MessageDefBuilder),
		registry: &Registry{},
	}
}

// ForMessageDef returns a builder for the message definition, which corresponds
// to the provided key. The key can be represented by an arbitrary value. Two
// calls to this method will return the builder for the same message definition,
// if the keys are equal, following Go's rules for interface comparison.
func (rb *RegistryBuilder) ForMessageDef(key interface{}) *MessageDefBuilder {
	if def, ok := rb.defs[key]; ok {
		return def
	}
	index := len(rb.defs)
	rb.registry.Defs = append(rb.registry.Defs, nil)
	def := &MessageDefBuilder{
		index:    index,
		registry: rb.registry,
		message: &MessageDef{
			Registry: rb.registry,
			DataType: DtEntity | DataType(index),
		},
	}
	rb.defs[key] = def
	return def
}

// Build creates the registry. Any subsequent calls to the builder will lead to
// undefined behavior.
func (rb *RegistryBuilder) Build() *Registry {
	for i, def := range rb.registry.Defs {
		if def == nil {
			panic(fmt.Sprintf("definition at %v is empty", i))
		}
	}
	return rb.registry
}

// -----------------------------------------------------------------------------
// Message definition builder

func (mb *MessageDefBuilder) WithName(name string) *MessageDefBuilder {
	mb.message.Name = name
	return mb
}

func (mb *MessageDefBuilder) WithNamespace(name string) *MessageDefBuilder {
	mb.message.Namespace = name
	return mb
}

func (mb *MessageDefBuilder) WithField(
	name string, tag uint64, dataType DataType) *MessageDefBuilder {
	mb.addField(tag, &MessageFieldDef{
		Name:     name,
		DataType: dataType,
		Tag:      tag,
		Repeated: false,
	})
	return mb
}

func (mb *MessageDefBuilder) WithArrayField(
	name string, tag uint64, dataType DataType) *MessageDefBuilder {
	mb.addField(tag, &MessageFieldDef{
		Name:     name,
		DataType: dataType,
		Tag:      tag,
		Repeated: true,
	})
	return mb
}

// ExtendField updates the last time added field with an extension, which may
// alter the way the field is serialized or deserialized.
func (mb *MessageDefBuilder) ExtendField(ext func(*MessageFieldDef)) *MessageDefBuilder {
	ext(mb.ensureFieldDef())
	return mb
}

func (mb *MessageDefBuilder) GetDataType() DataType { return mb.message.DataType }

// Build builds the message definition. If not called, the Build method of the
// RegistryBuilder will panic. Any subsequent calls to the builder will lead to
// undefined behavior.
func (mb *MessageDefBuilder) Build() *MessageDef {
	if mb.registry.Defs[mb.index] != nil {
		panic(fmt.Sprintf("message definition at %v has already been built", mb.index))
	}
	// Sorting the fields in order to optimize serialization and
	// deserialization of the messages.
	fields := mb.message.Fields
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Tag < fields[j].Tag
	})
	mb.registry.Defs[mb.index] = mb.message
	return mb.message
}

func (mb *MessageDefBuilder) addField(tag uint64, f *MessageFieldDef) {
	// Getting an offset of the value either in the primitive values array
	// or the references array.
	if f.DataType.IsRefType() || f.Repeated {
		f.Offset = mb.message.EntityBufLength
		mb.message.EntityBufLength++
	} else {
		f.Offset = mb.message.DataBufLength
		mb.message.DataBufLength += f.DataType.GetWidthInBytes()
	}
	mb.message.Fields = append(mb.message.Fields, f)
	mb.field = f
}

func (mb *MessageDefBuilder) ensureFieldDef() *MessageFieldDef {
	current := mb.field
	if current == nil {
		panic("method cannot be called before a field is added")
	}
	return current
}
