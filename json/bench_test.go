package json

import (
	"bytes"
	"encoding/json"
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
			_, err := DecodeNew(data, def)
			assert.NoError(b, err)
		}
	})
}

// BenchmarkReference explores other options to parse the JSON document.
func BenchmarkReference(b *testing.B) {
	def, entity := ArrangeEncodeDecode()
	data, err := Encode(entity, def)
	assert.NoError(b, err)

	// How much time it will take to enumerate the tokens using the
	// standard API of json package?
	b.Run("json.Decoder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(data)
			dec := json.NewDecoder(buf)
			dec.UseNumber()
			for {
				_, err := dec.Token()
				if err != nil {
					break
				}
			}
		}
	})

	// How much time it will take to unmarshal the JSON into a map
	// using the standard API of json package?
	b.Run("json.Unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var entity map[string]interface{}
			if err = json.Unmarshal(data, &entity); err != nil {
				b.Fatal(err)
			}
		}
	})
}
