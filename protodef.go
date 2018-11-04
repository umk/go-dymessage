package dymessage

import (
	"fmt"
)

type (
	ProtoRepo struct {
		// A collection of message definitions at the positions by which these
		// definitions are referenced from other ones and outside.
		defs []*ProtoDef
	}

	// A definition of dynamic protobuf message.
	ProtoDef struct {
		// The namespace and name of the message as it would be specified in the
		// message definition.
		Ns, Name string
		// Number of bytes taken by primitive values.
		dataLen int
		// Number of entities referenced by the root.
		entitiesLen int
		// A repository the definition belongs to.
		repo *ProtoRepo
		// Mapping from the field tag to the field description.
		Fields map[uint64]*ProtoField
	}

	ProtoField struct {
		Name     string
		DataType DataType
		Tag      uint64
		Repeated bool
		// Index of the field in the array of bytes when the field is of primitive
		// type and not repeated. Elsewhere, an index in the array of entities.
		index int
	}

	ProtoRepoBuilder struct {
		defs map[interface{}]*ProtoDefBuilder
		repo *ProtoRepo
	}

	ProtoDefBuilder struct {
		index int        // Index of proto definition in repository
		repo  *ProtoRepo // Repository that built definition belongs to
		def   *ProtoDef  // Dynamic message definition being built
	}

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
	bytes8  = 1
	bytes32 = 4
	bytes64 = 8
)

func (dt DataType) isRef() bool {
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

func (dt DataType) getSize() int {
	if dt.isRef() {
		panic("operation is valid only for primitive types")
	}
	switch dt {
	case DtInt32, DtUint32, DtFloat32:
		return bytes32
	case DtInt64, DtUint64, DtFloat64:
		return bytes64
	case DtBool:
		return bytes8
	default:
		panic(dt)
	}
}

// -----------------------------------------------------------------------------
// Repository builder

func NewRepoBuilder() *ProtoRepoBuilder {
	return &ProtoRepoBuilder{
		defs: make(map[interface{}]*ProtoDefBuilder),
		repo: &ProtoRepo{},
	}
}

func (rb *ProtoRepoBuilder) AddProtoDef(
	key interface{}, ns, name string) *ProtoDefBuilder {
	def := rb.ensureDef(key)
	if def.def != nil {
		panic(fmt.Sprintf(
			"entity %v has already been added at %v", key, def.index))
	}
	def.def = &ProtoDef{
		Ns:     ns,
		Name:   name,
		repo:   rb.repo,
		Fields: make(map[uint64]*ProtoField),
	}
	return def
}

// GetEntityType gets the data type of the field, which value is an entity. The
// method accepts the key by which the entity is referenced in the repository
// builder, and if necessary reserves an index for the proto definition. The
// called is obliged to build the entity by provided key.
func (rb *ProtoRepoBuilder) GetEntityType(key interface{}) DataType {
	def := rb.ensureDef(key)
	return DtEntity | DataType(def.index)
}

func (rb *ProtoRepoBuilder) ensureDef(key interface{}) *ProtoDefBuilder {
	if def, ok := rb.defs[key]; ok {
		return def
	}
	index := len(rb.defs)
	rb.repo.defs = append(rb.repo.defs, nil)
	def := &ProtoDefBuilder{
		index: index,
		repo:  rb.repo,
	}
	rb.defs[key] = def
	return def
}

func (rb *ProtoRepoBuilder) Build() *ProtoRepo {
	for i, def := range rb.repo.defs {
		if def == nil {
			panic(fmt.Sprintf("definition at %v is empty", i))
		}
	}
	return rb.repo
}

// -----------------------------------------------------------------------------
// Proto definition builder

func (pb *ProtoDefBuilder) WithField(
	tag uint64, name string, dataType DataType) *ProtoDefBuilder {
	pb.addField(tag, &ProtoField{
		Name:     name,
		DataType: dataType,
		Tag:      tag,
		Repeated: false,
	})
	return pb
}

func (pb *ProtoDefBuilder) WithArrayField(
	tag uint64, name string, dataType DataType) *ProtoDefBuilder {
	pb.addField(tag, &ProtoField{
		Name:     name,
		DataType: dataType,
		Tag:      tag,
		Repeated: true,
	})
	return pb
}

func (pb *ProtoDefBuilder) addField(tag uint64, f *ProtoField) {
	if _, ok := pb.def.Fields[tag]; ok {
		panic(fmt.Sprintf("field with tag %v already exists", tag))
	}
	// Getting an offset of the value either in the primitive values array or the
	// references array.
	if f.DataType.isRef() || f.Repeated {
		f.index = pb.def.entitiesLen
		pb.def.entitiesLen++
	} else {
		f.index = pb.def.dataLen
		pb.def.dataLen += f.DataType.getSize()
	}
	pb.def.Fields[tag] = f
}

func (pb *ProtoDefBuilder) Build() *ProtoDef {
	if pb.repo.defs[pb.index] != nil {
		panic(fmt.Sprintf("proto def at %v has already been built", pb.index))
	}
	pb.repo.defs[pb.index] = pb.def
	return pb.def
}

// -----------------------------------------------------------------------------
// Proto definition

func (p *ProtoDef) NewEntity() *Entity {
	return &Entity{
		Data:     make([]byte, p.dataLen),
		Entities: make([]*Entity, p.entitiesLen),
	}
}
