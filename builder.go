package dymessage

import (
	"fmt"
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
		// Dynamic message definition which is being built by this
		// instance of builder.
		def *MessageDef
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

func (rb *RegistryBuilder) ForMessageDef(key interface{}) *MessageDefBuilder {
	if def, ok := rb.defs[key]; ok {
		return def
	}
	index := len(rb.defs)
	rb.registry.Defs = append(rb.registry.Defs, nil)
	def := &MessageDefBuilder{
		index:    index,
		registry: rb.registry,
		def: &MessageDef{
			Registry: rb.registry,
			DataType: DtEntity | DataType(index),
			Fields:   make(map[uint64]*MessageFieldDef),
		},
	}
	rb.defs[key] = def
	return def
}

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
	mb.def.Name = name
	return mb
}

func (mb *MessageDefBuilder) WithNamespace(name string) *MessageDefBuilder {
	mb.def.Namespace = name
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

func (mb *MessageDefBuilder) GetDataType() DataType { return mb.def.DataType }

func (mb *MessageDefBuilder) Build() *MessageDef {
	if mb.registry.Defs[mb.index] != nil {
		panic(fmt.Sprintf("message definition at %v has already been built", mb.index))
	}
	mb.registry.Defs[mb.index] = mb.def
	return mb.def
}

func (mb *MessageDefBuilder) addField(tag uint64, f *MessageFieldDef) {
	// Getting an offset of the value either in the primitive values array
	// or the references array.
	if f.DataType.IsRefType() || f.Repeated {
		f.Offset = mb.def.EntityBufLength
		mb.def.EntityBufLength++
	} else {
		f.Offset = mb.def.DataBufLength
		mb.def.DataBufLength += f.DataType.GetWidthInBytes()
	}
	mb.def.Fields[tag] = f
}
