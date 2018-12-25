package dymessage

import "fmt"

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

func (rb *RegistryBuilder) AddMessageDef(key interface{}) *MessageDefBuilder {
	def := rb.ensureDef(key)
	if def.def != nil {
		panic(fmt.Sprintf("entity %v has already been added at %v", key, def.index))
	}
	def.def = &MessageDef{
		Registry: rb.registry,
	}
	return def
}

// GetEntityType gets the data type of the field, which value is an entity. The
// method accepts the key by which the entity is referenced in the repository
// builder, and if necessary reserves an index for the proto definition. The
// called is obliged to build the entity by provided key.
func (rb *RegistryBuilder) GetEntityType(key interface{}) DataType {
	def := rb.ensureDef(key)
	return DtEntity | DataType(def.index)
}

func (rb *RegistryBuilder) Build() *Registry {
	for i, def := range rb.registry.Defs {
		if def == nil {
			panic(fmt.Sprintf("definition at %v is empty", i))
		}
	}
	return rb.registry
}

func (rb *RegistryBuilder) ensureDef(key interface{}) *MessageDefBuilder {
	if def, ok := rb.defs[key]; ok {
		return def
	}
	index := len(rb.defs)
	rb.registry.Defs = append(rb.registry.Defs, nil)
	def := &MessageDefBuilder{
		index:    index,
		registry: rb.registry,
	}
	rb.defs[key] = def
	return def
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

func (mb *MessageDefBuilder) Build() *MessageDef {
	if mb.registry.Defs[mb.index] != nil {
		panic(fmt.Sprintf("message definition at %v has already been built", mb.index))
	}
	mb.registry.Defs[mb.index] = mb.def
	return mb.def
}

func (mb *MessageDefBuilder) addField(tag uint64, f *MessageFieldDef) {
	// Getting an offset of the value either in the primitive values array or the
	// references array.
	if f.DataType.IsRefType() || f.Repeated {
		f.Offset = mb.def.entitiesLength
		mb.def.entitiesLength++
	} else {
		f.Offset = mb.def.dataLength
		mb.def.dataLength += f.DataType.GetSizeInBytes()
	}
	mb.def.Fields = append(mb.def.Fields, f)
}

// -----------------------------------------------------------------------------
// Message definition

func (md *MessageDef) NewEntity() *Entity {
	return &Entity{
		Data:     make([]byte, md.dataLength),
		Entities: make([]*Entity, md.entitiesLength),
	}
}
