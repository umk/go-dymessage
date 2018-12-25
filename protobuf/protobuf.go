package protobuf

import (
	"errors"

	. "github.com/umk/go-dymessage/protobuf/internal/impl"
)

type Encoder struct {
	// Indicates whether the unknown fields must be silently skipped.
	IgnoreUnknown bool
	buf           Buffer
}

var RepeatedNullErr = errors.New("repeated field has null item")
