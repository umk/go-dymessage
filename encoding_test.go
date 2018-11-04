package dymessage

import (
	"testing"

	. "github.com/umk/go-dymessage/testdata"

	"github.com/stretchr/testify/require"
	"github.com/umk/protobuf/proto"
)

func TestEncodeDecode(t *testing.T) {
	def, entity := arrangeEncodeDecode()

	// Checking whether the message can be read right after is has been composed.
	assertEncodeDecode(t, def, entity)

	// Converting message to protobuf message and back.
	enc := &Encoder{Relaxed: false}
	data, err := enc.Encode(entity, def)
	require.NoError(t, err)

	message := new(TestMessage)
	err = proto.Unmarshal(data, message)
	require.NoError(t, err)

	data, err = proto.Marshal(message)
	require.NoError(t, err)

	entity2, err := enc.Decode(data, def)
	require.NoError(t, err)

	// Checking values of the converted message.
	assertEncodeDecode(t, def, entity2)
}
