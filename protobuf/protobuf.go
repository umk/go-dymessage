package protobuf

import (
	"errors"

	. "github.com/umk/go-dymessage/protobuf/internal/impl"
)

type Encoder struct {
	// Indicates whether the unknown fields must be silently skipped.
	Relaxed bool
	buf     Buffer
}

var BadMessageErr = errors.New("bad message")
var RepeatedNullErr = errors.New("repeated field has null item")
