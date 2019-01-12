package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/umk/go-dymessage"
)

type (
	// Attaches the methods for the builder to create the test
	// message definitions.
	TestBuilder struct {
		*RegistryBuilder
	}

	arrayReader struct {
		*MessageFieldDef
		t *testing.T
	}
)

const (
	TagRegInt32 = iota + 1
	TagRegInt64
	TagRegUint32
	TagRegUint64
	TagRegFloat32
	TagRegFloat64
	TagRegBool
	TagRegString
	TagRegBytes
	TagRegEntity

	TagArrInt32
	TagArrInt64
	TagArrUint32
	TagArrUint64
	TagArrFloat32
	TagArrFloat64
	TagArrBool
	TagArrString
	TagArrBytes
	TagArrEntity
)

func (rb *TestBuilder) CreateTestMessage(key interface{}, namespace, name string) *MessageDefBuilder {
	return rb.ForMessageDef(key).
		WithNamespace(namespace).
		WithName(name).
		// regular fields
		WithField("RegInt32", TagRegInt32, DtInt32).
		WithField("RegInt64", TagRegInt64, DtInt64).
		WithField("RegUint32", TagRegUint32, DtUint32).
		WithField("RegUint64", TagRegUint64, DtUint64).
		WithField("RegFloat32", TagRegFloat32, DtFloat32).
		WithField("RegFloat64", TagRegFloat64, DtFloat64).
		WithField("RegBool", TagRegBool, DtBool).
		WithField("RegString", TagRegString, DtString).
		WithField("RegBytes", TagRegBytes, DtBytes).
		// repeated fields
		WithArrayField("ArrInt32", TagArrInt32, DtInt32).
		WithArrayField("ArrInt64", TagArrInt64, DtInt64).
		WithArrayField("ArrUint32", TagArrUint32, DtUint32).
		WithArrayField("ArrUint64", TagArrUint64, DtUint64).
		WithArrayField("ArrFloat32", TagArrFloat32, DtFloat32).
		WithArrayField("ArrFloat64", TagArrFloat64, DtFloat64).
		WithArrayField("ArrBool", TagArrBool, DtBool).
		WithArrayField("ArrString", TagArrString, DtString).
		WithArrayField("ArrBytes", TagArrBytes, DtBytes)
}

func ArrangeEncodeDecode() (*MessageDef, *Entity) {
	rb := TestBuilder{NewRegistryBuilder()}

	def := rb.CreateTestMessage("message", "koala.goshawk", "Message").
		WithField("RegEntity", 10, rb.ForMessageDef("message").GetDataType()).
		WithArrayField("ArrEntity", 20, rb.ForMessageDef("message").GetDataType()).
		Build()

	child := def.NewEntity()

	def.GetField(TagRegInt32).SetPrimitive(child, FromInt32(868929107))
	def.GetField(TagRegInt64).SetPrimitive(child, FromInt64(-601380853565279092))
	def.GetField(TagRegUint32).SetPrimitive(child, FromUint32(783509315))
	def.GetField(TagRegUint64).SetPrimitive(child, FromUint64(54182615856980345))
	def.GetField(TagRegFloat32).SetPrimitive(child, FromFloat32(80116.7676))
	def.GetField(TagRegFloat64).SetPrimitive(child, FromFloat64(1.2262663))
	def.GetField(TagRegBool).SetPrimitive(child, FromBool(true))
	def.GetField(TagRegString).SetReference(child, FromString("Zy0RVazdEe459Y0DErUJ"))
	def.GetField(TagRegBytes).SetReference(child, FromBytes([]byte{232, 153, 178, 190, 4, 82}, false))
	def.GetField(TagRegEntity).SetReference(child, FromEntity(def.NewEntity()))

	def.GetField(TagArrInt32).Reserve(child, 2)
	def.GetField(TagArrInt32).SetPrimitiveAt(child, 0, FromInt32(313261865))
	def.GetField(TagArrInt32).SetPrimitiveAt(child, 1, FromInt32(209295014))

	def.GetField(TagArrInt64).Reserve(child, 2)
	def.GetField(TagArrInt64).SetPrimitiveAt(child, 0, FromInt64(-394578838447094537))
	def.GetField(TagArrInt64).SetPrimitiveAt(child, 1, FromInt64(7197041536234632))

	def.GetField(TagArrUint32).Reserve(child, 2)
	def.GetField(TagArrUint32).SetPrimitiveAt(child, 0, FromUint32(849851282))
	def.GetField(TagArrUint32).SetPrimitiveAt(child, 1, FromUint32(248557441))

	def.GetField(TagArrUint64).Reserve(child, 2)
	def.GetField(TagArrUint64).SetPrimitiveAt(child, 0, FromUint64(4416324197982829))
	def.GetField(TagArrUint64).SetPrimitiveAt(child, 1, FromUint64(218233954665294213))

	def.GetField(TagArrFloat32).Reserve(child, 2)
	def.GetField(TagArrFloat32).SetPrimitiveAt(child, 0, FromFloat32(9296232.53))
	def.GetField(TagArrFloat32).SetPrimitiveAt(child, 1, FromFloat32(-54836.8569))

	def.GetField(TagArrFloat64).Reserve(child, 2)
	def.GetField(TagArrFloat64).SetPrimitiveAt(child, 0, FromFloat64(-682925.9662517307))
	def.GetField(TagArrFloat64).SetPrimitiveAt(child, 1, FromFloat64(719704.153))

	def.GetField(TagArrBool).Reserve(child, 2)
	def.GetField(TagArrBool).SetPrimitiveAt(child, 0, FromBool(true))
	def.GetField(TagArrBool).SetPrimitiveAt(child, 1, FromBool(false))

	def.GetField(TagArrString).Reserve(child, 2)
	def.GetField(TagArrString).SetReferenceAt(child, 0, FromString("HN89fTSfx2it9Ma11Ufj"))
	def.GetField(TagArrString).SetReferenceAt(child, 1, FromString("f4nuZTeXQmsvR6MBPkC"))

	def.GetField(TagArrBytes).Reserve(child, 2)
	def.GetField(TagArrBytes).SetReferenceAt(child, 0, FromBytes([]byte{189, 248, 87, 249, 19, 15}, false))
	def.GetField(TagArrBytes).SetReferenceAt(child, 1, FromBytes([]byte{22, 72, 74, 121, 208}, false))

	def.GetField(TagArrEntity).Reserve(child, 3)
	def.GetField(TagArrEntity).SetReferenceAt(child, 0, FromEntity(def.NewEntity()))
	def.GetField(TagArrEntity).SetReferenceAt(child, 1, FromEntity(def.NewEntity()))
	def.GetField(TagArrEntity).SetReferenceAt(child, 2, FromEntity(def.NewEntity()))

	entity := def.NewEntity()

	def.GetField(TagRegInt32).SetPrimitive(entity, FromInt32(-33512104))
	def.GetField(TagRegInt64).SetPrimitive(entity, FromInt64(-254715376635680503))
	def.GetField(TagRegUint32).SetPrimitive(entity, FromUint32(799283559))
	def.GetField(TagRegUint64).SetPrimitive(entity, FromUint64(65911047815132225))
	def.GetField(TagRegFloat32).SetPrimitive(entity, FromFloat32(-204860.936))
	def.GetField(TagRegFloat64).SetPrimitive(entity, FromFloat64(510.972845))
	def.GetField(TagRegBool).SetPrimitive(entity, FromBool(false))
	def.GetField(TagRegString).SetReference(entity, FromString("LJFzUzsO2O8auQAlVmJy"))
	def.GetField(TagRegBytes).SetReference(entity, FromBytes([]byte{24, 40, 107, 129, 64}, false))
	def.GetField(TagRegEntity).SetReference(entity, FromEntity(child))

	return def, entity
}

func AssertEncodeDecode(t *testing.T, def *MessageDef, entity *Entity) {
	ref := def.GetField(TagRegEntity).GetReference(entity)
	require.NotNil(t, ref)
	linnet := ref.ToEntity()

	require.Equal(t, 3, def.GetField(TagArrEntity).Len(linnet))

	assert.Equal(t, int32(-33512104), def.GetField(TagRegInt32).GetPrimitive(entity).ToInt32())
	assert.Equal(t, int64(-254715376635680503), def.GetField(TagRegInt64).GetPrimitive(entity).ToInt64())
	assert.Equal(t, uint32(799283559), def.GetField(TagRegUint32).GetPrimitive(entity).ToUint32())
	assert.Equal(t, uint64(65911047815132225), def.GetField(TagRegUint64).GetPrimitive(entity).ToUint64())
	assert.Equal(t, float32(-204860.936), def.GetField(TagRegFloat32).GetPrimitive(entity).ToFloat32())
	assert.Equal(t, float64(510.972845), def.GetField(TagRegFloat64).GetPrimitive(entity).ToFloat64())

	assert.Equal(t, false, def.GetField(TagRegBool).GetPrimitive(entity).ToBool())
	assert.Equal(t, "LJFzUzsO2O8auQAlVmJy", def.GetField(TagRegString).GetReference(entity).ToString())
	assert.Equal(t, []byte{24, 40, 107, 129, 64}, def.GetField(TagRegBytes).GetReference(entity).ToBytes())

	assert.Equal(t, int32(868929107), def.GetField(TagRegInt32).GetPrimitive(linnet).ToInt32())
	assert.Equal(t, int64(-601380853565279092), def.GetField(TagRegInt64).GetPrimitive(linnet).ToInt64())
	assert.Equal(t, uint32(783509315), def.GetField(TagRegUint32).GetPrimitive(linnet).ToUint32())
	assert.Equal(t, uint64(54182615856980345), def.GetField(TagRegUint64).GetPrimitive(linnet).ToUint64())
	assert.Equal(t, float32(80116.7676), def.GetField(TagRegFloat32).GetPrimitive(linnet).ToFloat32())
	assert.Equal(t, float64(1.2262663), def.GetField(TagRegFloat64).GetPrimitive(linnet).ToFloat64())

	assert.Equal(t, true, def.GetField(TagRegBool).GetPrimitive(linnet).ToBool())
	assert.Equal(t, "Zy0RVazdEe459Y0DErUJ", def.GetField(TagRegString).GetReference(linnet).ToString())
	assert.Equal(t, []byte{232, 153, 178, 190, 4, 82}, def.GetField(TagRegBytes).GetReference(linnet).ToBytes())

	assert.Equal(t, int32(313261865), readArr(t, def.GetField(TagArrInt32)).getValueAt(linnet, 0).ToInt32())
	assert.Equal(t, int32(209295014), readArr(t, def.GetField(TagArrInt32)).getValueAt(linnet, 1).ToInt32())

	assert.Equal(t, int64(-394578838447094537), readArr(t, def.GetField(TagArrInt64)).getValueAt(linnet, 0).ToInt64())
	assert.Equal(t, int64(7197041536234632), readArr(t, def.GetField(TagArrInt64)).getValueAt(linnet, 1).ToInt64())

	assert.Equal(t, uint32(849851282), readArr(t, def.GetField(TagArrUint32)).getValueAt(linnet, 0).ToUint32())
	assert.Equal(t, uint32(248557441), readArr(t, def.GetField(TagArrUint32)).getValueAt(linnet, 1).ToUint32())

	assert.Equal(t, uint64(4416324197982829), readArr(t, def.GetField(TagArrUint64)).getValueAt(linnet, 0).ToUint64())
	assert.Equal(t, uint64(218233954665294213), readArr(t, def.GetField(TagArrUint64)).getValueAt(linnet, 1).ToUint64())

	assert.Equal(t, float32(9296232.53), readArr(t, def.GetField(TagArrFloat32)).getValueAt(linnet, 0).ToFloat32())
	assert.Equal(t, float32(-54836.8569), readArr(t, def.GetField(TagArrFloat32)).getValueAt(linnet, 1).ToFloat32())

	assert.Equal(t, float64(-682925.966251730668), readArr(t, def.GetField(TagArrFloat64)).getValueAt(linnet, 0).ToFloat64())
	assert.Equal(t, float64(719704.153), readArr(t, def.GetField(TagArrFloat64)).getValueAt(linnet, 1).ToFloat64())

	assert.Equal(t, true, readArr(t, def.GetField(TagArrBool)).getValueAt(linnet, 0).ToBool())
	assert.Equal(t, false, readArr(t, def.GetField(TagArrBool)).getValueAt(linnet, 1).ToBool())

	assert.Equal(t, "HN89fTSfx2it9Ma11Ufj", readArr(t, def.GetField(TagArrString)).getEntityAt(linnet, 0).ToString())
	assert.Equal(t, "f4nuZTeXQmsvR6MBPkC", readArr(t, def.GetField(TagArrString)).getEntityAt(linnet, 1).ToString())

	assert.Equal(t, []byte{189, 248, 87, 249, 19, 15}, readArr(t, def.GetField(TagArrBytes)).getEntityAt(linnet, 0).ToBytes())
	assert.Equal(t, []byte{22, 72, 74, 121, 208}, readArr(t, def.GetField(TagArrBytes)).getEntityAt(linnet, 1).ToBytes())
}

// -----------------------------------------------------------------------------
// Helper functions

// Returns a proto field, which represents a collection and checks for errors
// before returning a value at particular index.
func readArr(t *testing.T, f *MessageFieldDef) *arrayReader {
	return &arrayReader{MessageFieldDef: f, t: t}
}

func (f *arrayReader) getValueAt(e *Entity, n int) Primitive {
	return f.MessageFieldDef.GetPrimitiveAt(e, n)
}

func (f *arrayReader) getEntityAt(e *Entity, n int) Reference {
	return f.MessageFieldDef.GetReferenceAt(e, n)
}
