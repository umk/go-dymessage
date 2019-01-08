package protobuf

import (
	"github.com/umk/go-dymessage/protobuf/internal/impl"
)

// Provides the methods to encode the dynamic message to and from the protocol
// buffers; provides the parameters of encoding.
//
// The methods, which implement encoding and decoding of the dynamic entities,
// are NOT thread-safe, but the instance of Encoder can be reused in a
// non-concurrent manner.
type Encoder struct {
	// The buffer currently being read or written by Encoder.
	buf *impl.Buffer
	// A collection of buffers to reuse for encoding and decoding of the
	// nested entities.
	bufs []*impl.Buffer
}

// borrowBuf gets a new buffer from the bufs collection or creates a new one if
// collection is empty. Then the buffer is assigned as current one. The returned
// value is a function to be used to return the buffer back to bufs collection
// and restore the previous buffer.
func (ec *Encoder) borrowBuf() func() {
	n := len(ec.bufs)
	var buf *impl.Buffer
	if n > 0 {
		buf, ec.bufs = ec.bufs[n-1], ec.bufs[:n-1]
	} else {
		buf = &impl.Buffer{}
	}
	prev := ec.buf
	ec.buf = buf
	return func() {
		ec.buf.Reset()
		ec.bufs = append(ec.bufs, ec.buf)
		ec.buf = prev
	}
}

// pushBuf creates a new buffer and substitutes the one currently being read or
// written by this instance of Encoder. The returned value is a function to be
// used to restore the previous buffer.
func (ec *Encoder) pushBuf(data []byte) func() {
	n := len(ec.bufs)
	var buf *impl.Buffer
	if n > 0 {
		buf, ec.bufs = ec.bufs[n-1], ec.bufs[:n-1]
	} else {
		buf = &impl.Buffer{}
	}
	prev := ec.buf
	ec.buf = buf
	bufData := ec.buf.Bytes()
	ec.buf.SetBuf(data)
	return func() {
		ec.buf.SetBuf(bufData)
		ec.bufs = append(ec.bufs, ec.buf)
		ec.buf = prev
	}
}
