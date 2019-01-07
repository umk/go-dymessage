package protobuf

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

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
