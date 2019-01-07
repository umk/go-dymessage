package protobuf

import (
	"math/rand"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/umk/go-dymessage"
	. "github.com/umk/go-dymessage/internal/testing"
	"github.com/umk/go-dymessage/protobuf/internal/testdata"
)

func BenchmarkTestEncodeRegular(b *testing.B) {
	def, entity := ArrangeEncodeDecode()
	enc := Encoder{}

	b.Run("encode regular", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := enc.Encode(entity, def)
			assert.NoError(b, err)
		}
	})

	// Each iteration will have its own encoder, so its internal
	// structures won't be shared between different calls.
	b.Run("encode regular new encoder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			enc := Encoder{}
			_, err := enc.Encode(entity, def)
			assert.NoError(b, err)
		}
	})
}

func BenchmarkTestDecodeRegular(b *testing.B) {
	def, entity := ArrangeEncodeDecode()
	enc := Encoder{}
	data, err := enc.Encode(entity, def)
	assert.NoError(b, err)

	message := def.NewEntity()

	b.Run("decode regular", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := enc.DecodeInto(data, def, message)
			assert.NoError(b, err)
		}
	})

	// The decoder won't reuse the structures created for an existing
	// entity.
	b.Run("decode regular new entity", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := enc.DecodeInto(data, def, def.NewEntity())
			assert.NoError(b, err)
		}
	})

	// Each iteration will have its own encoder, so its internal
	// structures won't be shared between different calls.
	b.Run("decode regular new encoder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			enc := Encoder{}
			_, err := enc.DecodeInto(data, def, message)
			assert.NoError(b, err)
		}
	})

	// Neither the entity structures, nor the internal structures of encoder
	// are shared between the different calls.
	b.Run("decode regular new entity and encoder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			enc := Encoder{}
			_, err := enc.DecodeInto(data, def, def.NewEntity())
			assert.NoError(b, err)
		}
	})
}

func BenchmarkTestDecodeRegularShuffled(b *testing.B) {
	def, entity := ArrangeEncodeDecode()
	shuffleRegistryFields(def.Registry)

	enc := Encoder{}
	data, err := enc.Encode(entity, def)
	assert.NoError(b, err)

	message := def.NewEntity()

	// The optimization to find the fields won't be applied in contrast to
	// the regular decode.
	b.Run("decode regular shuffled", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := enc.DecodeInto(data, def, message)
			assert.NoError(b, err)
		}
	})

	// Neither of the optimization rules are followed.
	b.Run("decode regular worst case", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			enc := Encoder{}
			_, err := enc.DecodeInto(data, def, def.NewEntity())
			assert.NoError(b, err)
		}
	})
}

func BenchmarkTestProtoEncodeRegular(b *testing.B) {
	def, entity := ArrangeEncodeDecode()
	enc := Encoder{}
	data, err := enc.Encode(entity, def)
	assert.NoError(b, err)

	message := testdata.TestMessageRegular{}
	err = proto.Unmarshal(data, &message)
	assert.NoError(b, err)

	buf := proto.NewBuffer(data)
	b.Run("encode protobuf regular", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := buf.Marshal(&message)
			assert.NoError(b, err)
		}
	})
}

func BenchmarkTestProtoDecodeRegular(b *testing.B) {
	def, entity := ArrangeEncodeDecode()
	enc := Encoder{}
	data, err := enc.Encode(entity, def)
	assert.NoError(b, err)

	message := testdata.TestMessageRegular{}

	b.Run("decode protobuf regular", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err = proto.Unmarshal(data, &message)
			assert.NoError(b, err)
		}
	})
}

// -----------------------------------------------------------------------------
// Helper functions

// shuffleRegistryFields randomly shuffles the fields of all dynamic entities
// declared in the registry.
func shuffleRegistryFields(reg *dymessage.Registry) {
	for _, def := range reg.Defs {
		rand.Shuffle(len(def.Fields), func(i, j int) {
			def.Fields[i], def.Fields[j] = def.Fields[j], def.Fields[i]
		})
	}
}
