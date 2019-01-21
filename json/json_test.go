package json

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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

	entity2, err := DecodeNew(data, def)
	require.NoError(t, err)

	// Checking values of the converted message.
	AssertEncodeDecode(t, def, entity2)
}

// TestJsonDecodeRandom tests whether the unknown fields are processed
// correctly.
func TestJsonDecodeRandom(t *testing.T) {
	// Whatever message definition it is, it's still useful.
	def, _ := ArrangeEncodeDecode()

	root, err := os.Getwd()
	if err != nil {
		panic("could not get working directory")
	}
	path := filepath.Join(root, "internal/testdata/random")
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() && filepath.Ext(path) == ".json" {
			t.Run(filepath.Base(path), func(t *testing.T) {
				data, err := ioutil.ReadFile(path)
				if err != nil {
					panic(err)
				}
				_, err = DecodeNew(data, def)
				assert.NoError(t, err)
			})
		}
		return nil
	})
}
