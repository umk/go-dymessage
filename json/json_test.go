package json

import (
	"testing"

	"github.com/stretchr/testify/require"
	. "github.com/umk/go-dymessage/internal/testing"
)

func TestJsonEncodeDecode(t *testing.T) {
	def, entity := ArrangeEncodeDecode()

	// Checking whether the message can be read right after is has been composed.
	AssertEncodeDecode(t, def, entity)

	// Converting message to JSON and back.
	data, err := Encode(entity, def)
	require.NoError(t, err)

	t.Log(string(data))

	entity2, err := Decode(data, def)
	require.NoError(t, err)

	// Checking values of the converted message.
	AssertEncodeDecode(t, def, entity2)
}
