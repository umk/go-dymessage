package protobuf

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/umk/go-testutil"

	. "github.com/umk/go-dymessage"
	. "github.com/umk/go-dymessage/internal/testing"
)

type testLocator struct {
	// Mapping from the namespace to the content of the file generated for this
	// namespace.
	bufs map[string]*strings.Builder
}

func TestExport(t *testing.T) {
	rb := TestBuilder{
		RegistryBuilder: NewRegistryBuilder(),
	}
	rb.CreateTestProto("Cicada", "marten.colobus", "Cicada").
		WithField("RegEntity", 100, rb.GetEntityType("Hoopoe")).
		WithArrayField("ArrEntity", 101, rb.GetEntityType("Meerkat")).
		Build()
	rb.CreateTestProto("Hoopoe", "marten.colobus", "Hoopoe").
		WithField("RegEntity", 100, rb.GetEntityType("Cicada")).
		Build()
	rb.CreateTestProto("Meerkat", "marten.heron", "Meerkat").
		Build()

	reg, loc := rb.Build(), &testLocator{}
	err := ExportProto(reg, loc)

	require.NoError(t, err)
	require.Len(t, loc.bufs, 2)

	wd, _ := os.Getwd()
	root := filepath.Join(wd, "../internal/testdata")

	if testutil.DoFix() {
		os.MkdirAll(root, os.ModeDir|os.ModePerm)
		for ns, buf := range loc.bufs {
			f, err := os.Create(filepath.Join(root, ns+".src"))
			require.NoError(t, err)
			f.WriteString(buf.String())
			f.Close()
		}
	}

	for ns, buf := range loc.bufs {
		fn := filepath.Join(root, ns+".src")
		f, err := os.Open(fn)
		require.NoError(t, err)
		data, _ := ioutil.ReadAll(f)
		testutil.EqualDiff(t, string(data), buf.String(), fn)
	}
}

// -----------------------------------------------------------------------------
// Helper methods

func (loc *testLocator) CreateWriter(ns string) (io.Writer, error) {
	builder := new(strings.Builder)
	if loc.bufs == nil {
		loc.bufs = make(map[string]*strings.Builder)
	}
	loc.bufs[ns] = builder
	return builder, nil
}

func (loc *testLocator) GetImport(ns string) string { return ns + ".proto" }
