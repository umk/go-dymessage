package protobuf

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	. "github.com/umk/go-dymessage"
	. "github.com/umk/go-dymessage/internal/testing"
	"github.com/umk/go-testutil"
)

type testLocator struct {
	// Mapping from the namespace to the content of the file generated for this
	// namespace.
	bufs map[string]*strings.Builder
}

const (
	TagCicadaRegEntity = iota + 100
	TagCicadaArrEntity
)

const (
	TagCicadaRegZigzagInt32 = iota + 200
	TagCicadaRegZigzagInt64
	TagCicadaRegVarintInt32
	TagCicadaRegVarintInt64
	TagCicadaRegVarintUint32
	TagCicadaRegVarintUint64
)

const (
	TagHoopoeRegEntity = iota + 100
)

func TestExport(t *testing.T) {
	rb := TestBuilder{
		RegistryBuilder: NewRegistryBuilder(),
	}

	// Cicada
	rb.CreateTestMessage("Cicada", "marten.colobus", "Cicada").
		WithField("RegEntity", TagCicadaRegEntity, rb.ForMessageDef("Hoopoe").GetDataType()).
		WithArrayField("ArrEntity", TagCicadaArrEntity, rb.ForMessageDef("Meerkat").GetDataType()).
		// extended fields
		WithField("RegZigzagInt32", TagCicadaRegZigzagInt32, DtInt32).ExtendField(WithZigZag()).
		WithField("RegZigzagInt64", TagCicadaRegZigzagInt64, DtInt64).ExtendField(WithZigZag()).
		WithField("RegVarintInt32", TagCicadaRegVarintInt32, DtInt32).ExtendField(WithVarint()).
		WithField("RegVarintInt64", TagCicadaRegVarintInt64, DtInt64).ExtendField(WithVarint()).
		WithField("RegVarintUint32", TagCicadaRegVarintUint32, DtUint32).ExtendField(WithVarint()).
		WithField("RegVarintUint64", TagCicadaRegVarintUint64, DtUint64).ExtendField(WithVarint()).
		Build()

	// Hoopoe
	rb.CreateTestMessage("Hoopoe", "marten.colobus", "Hoopoe").
		WithField("RegEntity", TagHoopoeRegEntity, rb.ForMessageDef("Cicada").GetDataType()).
		Build()

	// Meerkat
	rb.CreateTestMessage("Meerkat", "marten.heron", "Meerkat").
		Build()

	reg, loc := rb.Build(), &testLocator{}
	err := ExportToProto(reg, loc)

	require.NoError(t, err)
	require.Len(t, loc.bufs, 2)

	wd, _ := os.Getwd()
	root := filepath.Join(wd, "internal/testdata")

	if testutil.DoFix() {
		_ = os.MkdirAll(root, os.ModeDir|os.ModePerm)
		for ns, buf := range loc.bufs {
			f, err := os.Create(filepath.Join(root, ns+".src"))
			require.NoError(t, err)
			_, _ = f.WriteString(buf.String())
			_ = f.Close()
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
