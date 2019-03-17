package grpc

import (
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"

	"github.com/umk/go-dymessage"
	"github.com/umk/go-dymessage/protobuf"
)

type Encoder struct {
	reg *dymessage.Registry
	// The mapping from the qualified name of the messages to its definitions.
	types map[string]*dymessage.MessageDef
}

func NewEncoder(reg *dymessage.Registry) *Encoder {
	types := make(map[string]*dymessage.MessageDef)
	for _, def := range reg.Defs {
		qname := getQualifiedName(def)
		types[qname] = def
	}
	return &Encoder{reg: reg, types: types}
}

func (ec *Encoder) Encode(value *dymessage.Entity) (*any.Any, error) {
	def := ec.reg.GetMessageDef(value.DataType)
	if data, err := protobuf.Encode(value, def); err != nil {
		return nil, err
	} else {
		return &any.Any{
			TypeUrl: "type.googleapis.com/" + getQualifiedName(def),
			Value:   data,
		}, nil
	}
}

func (ec *Encoder) Decode(value *any.Any) (*dymessage.Entity, error) {
	if name, err := ptypes.AnyMessageName(value); err != nil {
		return nil, err
	} else if def, ok := ec.types[name]; !ok {
		return nil, fmt.Errorf("dymessage: type %q could not be found", name)
	} else {
		return protobuf.DecodeNew(value.Value, def)
	}
}

func getQualifiedName(def *dymessage.MessageDef) string {
	if len(def.Namespace) == 0 {
		return def.Name
	} else {
		return def.Namespace + "." + def.Name
	}
}
