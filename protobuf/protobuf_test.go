package protobuf

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"

	"github.com/umk/go-dymessage/internal/testdata"
	. "github.com/umk/go-dymessage/internal/testing"
)

func TestEncodeDecode(t *testing.T) {
	def, entity := ArrangeEncodeDecode()

	// Checking whether the message can be read right after is has been composed.
	AssertEncodeDecode(t, def, entity)

	// Converting message to protobuf message and back.
	enc := &Encoder{IgnoreUnknown: false}
	data, err := enc.Encode(entity, def)
	require.NoError(t, err)

	message := new(testdata.TestMessage)
	err = proto.Unmarshal(data, message)
	require.NoError(t, err)

	data, err = proto.Marshal(message)
	require.NoError(t, err)

	entity2, err := enc.Decode(data, def)
	require.NoError(t, err)

	// Checking values of the converted message.
	AssertEncodeDecode(t, def, entity2)
}
