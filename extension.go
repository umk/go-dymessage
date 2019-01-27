package dymessage

type (
	// Provides the information how to locate the extension in the container
	// of extensions. Pass this marker to the methods, which provide the
	// access to extensions.
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
// be called during init() of the package, which operates the extension. The
// returned marker must be provided to TryGetExtension to check if the container
// of extensions has the extension applied.
func RegisterExtension() ExtensionMarker {
	id := extensions.index
	extensions.index++
	return ExtensionMarker{index: id}
}

// TryGetExtension tries to get the extension by the marker, returned by the
// RegisterExtension method. The returned values are the instance of the
// extension and boolean value, indicating whether the extension has been found.
func (xt *Extensions) TryGetExtension(mk ExtensionMarker) (interface{}, bool) {
	if xt.ext == nil {
		return nil, false
	}
	extension := xt.ext[mk.index]
	return extension, extension != nil
}

// SetExtension sets an instance describing the extension in the container. The
// extension object is returned then by the TryGetExtension method if
// corresponding marker is provided.
func (xt *Extensions) SetExtension(mk ExtensionMarker, extension interface{}) {
	if xt.ext == nil {
		xt.ext = make([]interface{}, extensions.index)
	}
	xt.ext[mk.index] = extension
}
