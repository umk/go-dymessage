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

	b.Run("encode regular", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Encode(entity, def)
			assert.NoError(b, err)
		}
	})
}

func BenchmarkTestDecodeRegular(b *testing.B) {
	def, entity := ArrangeEncodeDecode()
	data, err := Encode(entity, def)
	assert.NoError(b, err)

	message := def.NewEntity()

	b.Run("decode regular", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Decode(data, def, message)
			assert.NoError(b, err)
		}
	})

	// The decoder won't reuse the structures created for an existing
	// entity.
	b.Run("decode regular new entity", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Decode(data, def, def.NewEntity())
			assert.NoError(b, err)
		}
	})
}

func BenchmarkTestEncodeDecodeAsync(b *testing.B) {
	def, entity := ArrangeEncodeDecode()
	data, err := Encode(entity, def)
	assert.NoError(b, err)

	b.RunParallel(func(pb *testing.PB) {
		message, decode := def.NewEntity(), true
		for pb.Next() {
			if decode {
				_, _ = Decode(data, def, message)
			} else {
				_, _ = Encode(entity, def)
			}
			decode = !decode
		}
	})
}

func BenchmarkTestDecodeRegularShuffled(b *testing.B) {
	def, entity := ArrangeEncodeDecode()
	shuffleRegistryFields(def.Registry)

	data, err := Encode(entity, def)
	assert.NoError(b, err)

	message := def.NewEntity()

	// The optimization to find the fields won't be applied in contrast to
	// the regular decode.
	b.Run("decode regular shuffled", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Decode(data, def, message)
			assert.NoError(b, err)
		}
	})
}

func BenchmarkTestProtoEncodeRegular(b *testing.B) {
	def, entity := ArrangeEncodeDecode()
	data, err := Encode(entity, def)
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
	data, err := Encode(entity, def)
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
