package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/umk/go-dymessage/internal/testing"
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

	b.Run("decode regular", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Decode(data, def)
			assert.NoError(b, err)
		}
	})
}
