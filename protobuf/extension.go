package protobuf

import (
	"fmt"
	"github.com/umk/go-dymessage"
)

type (
	// Provides information about the message fields, which can alter the
	// way these fields are serialized.
	extension struct {
		integerKind integerKind
	}
	// Declares the way to represent an integer value when serializing.
	integerKind int
)

const (
	ikDefault = integerKind(iota)
	ikZigZag
	ikVarint
)

var marker dymessage.ExtensionMarker

func init() {
	marker = dymessage.RegisterExtension()
}

// -----------------------------------------------------------------------------
// Extensions

// WithZigZag extends the message field definition to indicate that the value
// must be represented by a ZigZag encoding.
func WithZigZag() func(*dymessage.MessageFieldDef) {
	return withIntegerKind(
		ikZigZag,
		dymessage.DtInt32,
		dymessage.DtInt64)
}

// WithVarint extends the message field definition to indicate that the value
// must be represented by a ZigZag encoding.
func WithVarint() func(*dymessage.MessageFieldDef) {
	return withIntegerKind(
		ikVarint,
		dymessage.DtInt32,
		dymessage.DtInt64,
		dymessage.DtUint32,
		dymessage.DtUint64)
}

func withIntegerKind(
	kind integerKind, types ...dymessage.DataType) func(*dymessage.MessageFieldDef) {
	return func(def *dymessage.MessageFieldDef) {
		for _, t := range types {
			if def.DataType == t {
				extension := ensureExtension(def)
				if extension.integerKind != ikDefault {
					panic("kind of integer has already been specified")
				}
				extension.integerKind = kind
				return
			}
		}
		panic(fmt.Sprintf("field is of an invalid type %d", def.DataType))
	}
}

func tryGetExtension(def *dymessage.MessageFieldDef) (*extension, bool) {
	if ext, ok := def.TryGetExtension(marker); ok {
		return ext.(*extension), true
	}
	return nil, false
}

func ensureExtension(def *dymessage.MessageFieldDef) *extension {
	ext, ok := def.TryGetExtension(marker)
	if !ok {
		ext = &extension{}
		def.SetExtension(marker, ext)
	}
	return ext.(*extension)
}
