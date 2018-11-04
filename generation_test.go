package dymessage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/umk/go-testutil"
)

type exporter struct {
	// Mapping from the namespace to the content of the file generated for this
	// namespace.
	bufs map[string]*strings.Builder
}

func TestExport(t *testing.T) {
	rb := builder{
		NewRepoBuilder(),
	}
	rb.createTestProto("Message1", "ns1.ns2", "Message1").
		WithField(100, "RegEntity", rb.GetEntityType("Message2")).
		WithArrayField(101, "ArrEntity", rb.GetEntityType("Message3")).
		Build()
	rb.createTestProto("Message2", "ns1.ns2", "Message2").
		WithField(100, "RegEntity", rb.GetEntityType("Message1")).
		Build()
	rb.createTestProto("Message3", "ns1.ns3", "Message3").
		Build()

	repo, exp := rb.Build(), &exporter{}
	err := repo.Export(exp)

	require.NoError(t, err)
	require.Len(t, exp.bufs, 2)

	wd, _ := os.Getwd()
	root := filepath.Join(wd, "testdata")

	if testutil.DoFix() {
		os.MkdirAll(root, os.ModeDir|os.ModePerm)
		for ns, buf := range exp.bufs {
			f, err := os.Create(filepath.Join(root, ns+".src"))
			require.NoError(t, err)
			f.WriteString(buf.String())
			f.Close()
		}
	}

	for ns, buf := range exp.bufs {
		fn := filepath.Join(root, ns+".src")
		f, err := os.Open(fn)
		require.NoError(t, err)
		data, _ := ioutil.ReadAll(f)
		testutil.EqualDiff(t, string(data), buf.String(), fn)
	}
}
