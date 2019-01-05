package protobuf

import (
	"errors"
	"github.com/umk/go-dymessage/protobuf/internal/impl"
)

// Provides the methods to encode the dynamic message to and from the protocol
// buffers; provides the parameters of encoding.
//
// The methods, which implement encoding and decoding of the dynamic entities,
// are NOT thread-safe, but the instance of Encoder can be reused in a
// non-concurrent manner.
type Encoder struct {
	// Indicates whether the unknown fields must be silently skipped.
	IgnoreUnknown bool
	buf           impl.Buffer
}

var ErrRepeatedNull = errors.New("repeated field has null item")

func (ec *Encoder) clone() *Encoder {
	return &Encoder{
		IgnoreUnknown: ec.IgnoreUnknown,
	}
}
