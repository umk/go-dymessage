package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/umk/go-dymessage/internal/testing"
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
}

func BenchmarkTestDecodeRegular(b *testing.B) {
	def, entity := ArrangeEncodeDecode()
	enc := Encoder{}
	data, err := enc.Encode(entity, def)
	assert.NoError(b, err)

	message := def.NewEntity()

	b.Run("decode regular", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := enc.Decode(data, def, message)
			assert.NoError(b, err)
		}
	})
}
