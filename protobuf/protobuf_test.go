package protobuf

import (
	"github.com/umk/go-dymessage"
	"github.com/umk/go-dymessage/protobuf/internal/testdata"
	"math/rand"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"

	. "github.com/umk/go-dymessage/internal/testing"
)

func TestEncodeDecodeRegular(t *testing.T) {
	testEncodeDecode(
		t,
		new(testdata.TestMessageRegular),
		func(*dymessage.MessageDef) {})
}

func TestEncodeDecodeRegularShuffled(t *testing.T) {
	testEncodeDecode(
		t,
		new(testdata.TestMessageRegular),
		func(def *dymessage.MessageDef) {
			rand.Shuffle(
				len(def.Fields),
				func(i, j int) {
					def.Fields[i], def.Fields[j] = def.Fields[j], def.Fields[i]
				})
		})
}

func TestEncodeDecodeVarint(t *testing.T) {
	testEncodeDecode(
		t,
		new(testdata.TestMessageVarint),
		func(def *dymessage.MessageDef) {
			// Regular fields
			WithVarint()(def.GetField(TagRegInt32))
			WithVarint()(def.GetField(TagRegInt64))
			WithVarint()(def.GetField(TagRegUint32))
			WithVarint()(def.GetField(TagRegUint64))
			// Array fields
			WithVarint()(def.GetField(TagArrInt32))
			WithVarint()(def.GetField(TagArrInt64))
			WithVarint()(def.GetField(TagArrUint32))
			WithVarint()(def.GetField(TagArrUint64))
		})
}

func TestEncodeDecodeZigzag(t *testing.T) {
	testEncodeDecode(
		t,
		new(testdata.TestMessageZigzag),
		func(def *dymessage.MessageDef) {
			// Regular fields
			WithVarint()(def.GetField(TagRegInt32))
			WithVarint()(def.GetField(TagRegInt64))
			// Array fields
			WithVarint()(def.GetField(TagArrInt32))
			WithVarint()(def.GetField(TagArrInt64))
		})
}

func testEncodeDecode(t *testing.T, message proto.Message, setup func(*dymessage.MessageDef)) {
	def, entity := ArrangeEncodeDecode()
	setup(def)

	// Checking whether the message can be read right after is has been composed.
	AssertEncodeDecode(t, def, entity)

	// Converting message to protobuf message and back.
	enc := Encoder{IgnoreUnknown: false}
	data, err := enc.Encode(entity, def)
	require.NoError(t, err)

	err = proto.Unmarshal(data, message)
	require.NoError(t, err)

	data, err = proto.Marshal(message)
	require.NoError(t, err)

	entity2, err := enc.DecodeNew(data, def)
	require.NoError(t, err)

	// Checking values of the converted message.
	AssertEncodeDecode(t, def, entity2)
}
