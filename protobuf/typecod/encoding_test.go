package typecod

import (
	"testing"

	"github.com/stretchr/testify/require"
	. "github.com/umk/go-dymessage/internal/testing"
)

func TestEncodeDecode(t *testing.T) {
	def, entity := ArrangeEncodeDecode()
	cache := NewTypeCache(def.Registry)
	// Encoding
	any, err := EncodeAny(entity, cache)
	require.NoError(t, err)
	// Decoding
	entity, err = DecodeAny(any, cache)
	require.NoError(t, err)
	AssertEncodeDecode(t, def, entity)
	// Encoding again
	any, err = EncodeAny(entity, cache)
	require.NoError(t, err)
}

func TestDecodeUnknownType(t *testing.T) {
	def, entity := ArrangeEncodeDecode()
	cache := NewTypeCache(def.Registry)
	// Encoding
	any, err := EncodeAny(entity, cache)
	require.NoError(t, err)
	// Decoding with an unknown type URL
	any.TypeUrl = any.TypeUrl + "Unknown"
	entity, err = DecodeAny(any, cache)
	require.Error(t, err)
}
