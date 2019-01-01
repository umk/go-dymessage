package dymessage

type (
	// Provides the information how to locate the extension in the container
	// of extensions.
	ExtensionMarker struct{ index int }

	// A container for the extensions, applied to specific structure the
	// current one is a part of.
	Extensions struct {
		// A collection of extensions, which can be either empty, or
		// contain the number of items, which is equal to the number of
		// extensions, registered in the system.
		ext []interface{}
	}
)

var extensions = struct{ index int }{index: 0}

// RegisterExtension registers the dynamic message extension globally. This must
// be called during init() of the package, which operates the extension.
func RegisterExtension() ExtensionMarker {
	id := extensions.index
	extensions.index++
	return ExtensionMarker{index: id}
}

func (xt *Extensions) TryGetExtension(mk ExtensionMarker) (interface{}, bool) {
	if xt.ext == nil {
		return nil, false
	}
	extension := xt.ext[mk.index]
	return extension, extension != nil
}

func (xt *Extensions) SetExtension(mk ExtensionMarker, extension interface{}) {
	if len(xt.ext) == 0 {
		xt.ext = make([]interface{}, extensions.index)
	}
	xt.ext[mk.index] = extension
}
