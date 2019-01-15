package protobuf

import (
	"sync"

	"github.com/umk/go-dymessage"
	"github.com/umk/go-dymessage/protobuf/internal/impl"
)

var encoders sync.Pool

// Provides the methods to encode the dynamic message to and from the protocol
// buffers; provides the parameters of encoding.
//
// The methods, which implement encoding and decoding of the dynamic entities,
// are NOT thread-safe, but the instance of encoder can be reused in a
// non-concurrent manner.
type encoder struct {
	// The buffer currently being read or written by encoder.
	cur *impl.Buffer
	// A collection of buffers to reuse for encoding and decoding of the
	// nested entities.
	bufs []*impl.Buffer
}

func init() {
	encoders.New = func() interface{} { return new(encoder) }
}

// Encode encodes the data from the dynamic entity into a protocol buffers
// format. If the entity type doesn't correspond the data type of the message
// definition, the method will panic.
func Encode(e *dymessage.Entity, pd *dymessage.MessageDef) ([]byte, error) {
	encoder := getEncoder()
	buf, err := encoder.encode(e, pd, true)
	putEncoder(encoder)
	return buf, err
}

// DecodeNew transforms the protocol buffers representation of the message to a
// dynamic entity against the provided message definition.
func DecodeNew(b []byte, pd *dymessage.MessageDef) (*dymessage.Entity, error) {
	return Decode(b, pd, pd.NewEntity())
}

// Decode transforms the protocol buffers representation of the message to
// specified dynamic entity against the provided message definition. The
// returned entity is the one that has been provided as an input parameter e,
// but now populated with the data.
//
// If the entity type doesn't correspond the data type of the message
// definition, the method will panic.
func Decode(b []byte, pd *dymessage.MessageDef, e *dymessage.Entity) (*dymessage.Entity, error) {
	ec := getEncoder()
	err := ec.decode(b, pd, e)
	putEncoder(ec)
	return e, err
}

func getEncoder() *encoder   { return encoders.Get().(*encoder) }
func putEncoder(ec *encoder) { encoders.Put(ec) }

// -----------------------------------------------------------------------------
// Buffer manipulation

// borrowBuf gets a new buffer from the bufs collection or creates a new one if
// collection is empty. Then the buffer is assigned as current one, and previous
// one is returned.
func (ec *encoder) borrowBuf() (prevBuf *impl.Buffer) {
	n := len(ec.bufs)
	var buf *impl.Buffer
	if n > 0 {
		buf, ec.bufs = ec.bufs[n-1], ec.bufs[:n-1]
	} else {
		buf = &impl.Buffer{}
	}
	prevBuf = ec.cur
	ec.cur = buf
	return
}

// returnBuf returns the borrowed buffer back to the bufs collection and puts
// the provided buffer as the current one. The only parameter must be what did
// the borrowBuf method return.
func (ec *encoder) returnBuf(prevBuf *impl.Buffer) {
	ec.cur.Reset()
	ec.bufs = append(ec.bufs, ec.cur)
	ec.cur = prevBuf
}

// replaceBytes replaces the data in the current buffer with specified data and
// returns the previous data.
func (ec *encoder) replaceBytes(bytes []byte) (prev []byte) {
	prev = ec.cur.Bytes()
	ec.cur.SetBuf(bytes)
	return
}
