package json

type (
	Encoder struct {
		// Indicates whether encoder should produce human-readable output.
		Ident bool
		// Indicates whether the unknown fields must be silently skipped.
		Relaxed bool
	}

	// A mapping between the names of JSON fields and its values.
	fields map[string]interface{}
)
