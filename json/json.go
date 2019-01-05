package json

// Provides the methods to encode the dynamic message to and from the JSON;
// provides the parameters of encoding.
//
// The methods, which implement encoding and decoding of the dynamic entities,
// are thread-safe, so you can reuse the instance of Encoder is different
// threads.
type Encoder struct {
	// Indicates whether encoder should produce human-readable output.
	Ident bool
	// Indicates whether the unknown fields must be silently skipped.
	IgnoreUnknown bool
	// Indicates whether the message must contain all of the fields specified
	// in the message definition.
	RequireAll bool
}
