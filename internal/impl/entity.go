package impl

// Depending on the context, the entity represents either a regular entity with
// its own primitive and reference values, or the collection of either primitive
// or reference values, sharing the same type.
type Entity struct {
	Data     []byte    // Memory for storing the primitive values
	Entities []*Entity // The entities referenced from the current one
}
