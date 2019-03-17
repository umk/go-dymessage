package grpc

import (
	"testing"

	"github.com/stretchr/testify/require"
	. "github.com/umk/go-dymessage/internal/testing"
)

func TestEncodeDecode(t *testing.T) {
	def, entity := ArrangeEncodeDecode()
	encoder := NewEncoder(def.Registry)
	// Encoding
	any, err := encoder.Encode(entity)
	require.NoError(t, err)
	// Decoding
	entity, err = encoder.Decode(any)
	require.NoError(t, err)
	AssertEncodeDecode(t, def, entity)
	// Encoding again
	any, err = encoder.Encode(entity)
	require.NoError(t, err)
}

func TestDecodeUnknownType(t *testing.T) {
	def, entity := ArrangeEncodeDecode()
	encoder := NewEncoder(def.Registry)
	// Encoding
	any, err := encoder.Encode(entity)
	require.NoError(t, err)
	// Decoding with corrupted type URL
	any.TypeUrl = any.TypeUrl + "Unknown"
	entity, err = encoder.Decode(any)
	require.Error(t, err)
}
