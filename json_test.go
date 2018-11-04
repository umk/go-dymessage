package dymessage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJsonEncodeDecode(t *testing.T) {
	def, entity := arrangeEncodeDecode()

	// Checking whether the message can be read right after is has been composed.
	assertEncodeDecode(t, def, entity)

	// Converting message to JSON and back.
	enc := &JsonEncoder{Relaxed: false, Ident: true}
	data, err := enc.Encode(entity, def)
	require.NoError(t, err)

	entity2, err := enc.Decode(data, def)
	require.NoError(t, err)

	// Checking values of the converted message.
	assertEncodeDecode(t, def, entity2)
}
