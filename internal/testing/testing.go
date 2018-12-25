package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/umk/go-dymessage"
)

type (
	builder struct {
		*dymessage.RegistryBuilder
	}

	arrayReader struct {
		*dymessage.MessageFieldDef
		t *testing.T
	}
)

func (rb *builder) createTestProto(
    key interface{}, ns, name string) *dymessage.MessageDefBuilder {
	return rb.AddMessageDef(key).
		WithNamespace(ns).
		WithName(name).
		// regular fields
		WithField("RegInt32", 1, dymessage.DtInt32).
		WithField("RegInt64", 2, dymessage.DtInt64).
		WithField("RegUint32", 3, dymessage.DtUint32).
		WithField("RegUint64", 4, dymessage.DtUint64).
		WithField("RegFloat32", 5, dymessage.DtFloat32).
		WithField("RegFloat64", 6, dymessage.DtFloat64).
		WithField("RegBool", 7, dymessage.DtBool).
		WithField("RegString", 8, dymessage.DtString).
		WithField("RegBytes", 9, dymessage.DtBytes).
		// repeated fields
		WithArrayField("ArrInt32", 11, dymessage.DtInt32).
		WithArrayField("ArrInt64", 12, dymessage.DtInt64).
		WithArrayField("ArrUint32", 13, dymessage.DtUint32).
		WithArrayField("ArrUint64", 14, dymessage.DtUint64).
		WithArrayField("ArrFloat32", 15, dymessage.DtFloat32).
		WithArrayField("ArrFloat64", 16, dymessage.DtFloat64).
		WithArrayField("ArrBool", 17, dymessage.DtBool).
		WithArrayField("ArrString", 18, dymessage.DtString).
		WithArrayField("ArrBytes", 19, dymessage.DtBytes)
}

func ArrangeEncodeDecode() (*dymessage.MessageDef, *dymessage.Entity) {
	rb := builder{dymessage.NewRegistryBuilder()}

	def := rb.createTestProto("message", "koala.goshawk", "Message").
		WithField("RegEntity", 10, rb.GetEntityType("message")).
		WithArrayField("ArrEntity", 20, rb.GetEntityType("message")).
		Build()

	entity2 := def.NewEntity()

	def.Fields[1].SetValue(entity2,
		dymessage.FromInt32(868929107))
	def.Fields[2].SetValue(entity2,
		dymessage.FromInt64(-601380853565279092))
	def.Fields[3].SetValue(entity2,
		dymessage.FromUint32(783509315))
	def.Fields[4].SetValue(entity2,
		dymessage.FromUint64(54182615856980345))
	def.Fields[5].SetValue(entity2,
		dymessage.FromFloat32(80116.7676))
	def.Fields[6].SetValue(entity2,
		dymessage.FromFloat64(1.2262663))
	def.Fields[7].SetValue(entity2,
		dymessage.FromBool(true))
	def.Fields[8].SetEntity(entity2,
		dymessage.FromString("Zy0RVazdEe459Y0DErUJ"))
	def.Fields[9].SetEntity(entity2,
		dymessage.FromBytes([]byte{232, 153, 178, 190, 4, 82}, false))
	def.Fields[10].SetEntity(entity2,
		dymessage.FromEntity(def.NewEntity()))

	def.Fields[11].Reserve(entity2, 2)
	def.Fields[11].SetValueAt(entity2, 0,
		dymessage.FromInt32(313261865))
	def.Fields[11].SetValueAt(entity2, 1,
		dymessage.FromInt32(209295014))

	def.Fields[12].Reserve(entity2, 2)
	def.Fields[12].SetValueAt(entity2, 0,
		dymessage.FromInt64(-394578838447094537))
	def.Fields[12].SetValueAt(entity2, 1,
		dymessage.FromInt64(7197041536234632))

	def.Fields[13].Reserve(entity2, 2)
	def.Fields[13].SetValueAt(entity2, 0,
		dymessage.FromUint32(849851282))
	def.Fields[13].SetValueAt(entity2, 1,
		dymessage.FromUint32(248557441))

	def.Fields[14].Reserve(entity2, 2)
	def.Fields[14].SetValueAt(entity2, 0,
		dymessage.FromUint64(4416324197982829))
	def.Fields[14].SetValueAt(entity2, 1,
		dymessage.FromUint64(218233954665294213))

	def.Fields[15].Reserve(entity2, 2)
	def.Fields[15].SetValueAt(entity2, 0,
		dymessage.FromFloat32(9296232.53))
	def.Fields[15].SetValueAt(entity2, 1,
		dymessage.FromFloat32(-54836.8569))

	def.Fields[16].Reserve(entity2, 2)
	def.Fields[16].SetValueAt(entity2, 0,
		dymessage.FromFloat64(-682925.9662517307))
	def.Fields[16].SetValueAt(entity2, 1,
		dymessage.FromFloat64(719704.153))

	def.Fields[17].Reserve(entity2, 2)
	def.Fields[17].SetValueAt(entity2, 0,
		dymessage.FromBool(true))
	def.Fields[17].SetValueAt(entity2, 1,
		dymessage.FromBool(false))

	def.Fields[18].Reserve(entity2, 2)
	def.Fields[18].SetEntityAt(entity2, 0,
		dymessage.FromString("HN89fTSfx2it9Ma11Ufj"))
	def.Fields[18].SetEntityAt(entity2, 1,
		dymessage.FromString("f4nuZTeXQmsvR6MBPkC"))

	def.Fields[19].Reserve(entity2, 2)
	def.Fields[19].SetEntityAt(entity2, 0,
		dymessage.FromBytes([]byte{189, 248, 87, 249, 19, 15}, false))
	def.Fields[19].SetEntityAt(entity2, 1,
		dymessage.FromBytes([]byte{22, 72, 74, 121, 208}, false))

	def.Fields[20].Reserve(entity2, 3)
	def.Fields[20].SetEntityAt(entity2, 0,
		dymessage.FromEntity(def.NewEntity()))
	def.Fields[20].SetEntityAt(entity2, 1,
		dymessage.FromEntity(def.NewEntity()))
	def.Fields[20].SetEntityAt(entity2, 2,
		dymessage.FromEntity(def.NewEntity()))

	entity := def.NewEntity()

	def.Fields[1].SetValue(entity,
		dymessage.FromInt32(-33512104))
	def.Fields[2].SetValue(entity,
		dymessage.FromInt64(-254715376635680503))
	def.Fields[3].SetValue(entity,
		dymessage.FromUint32(799283559))
	def.Fields[4].SetValue(entity,
		dymessage.FromUint64(65911047815132225))
	def.Fields[5].SetValue(entity,
		dymessage.FromFloat32(-204860.936))
	def.Fields[6].SetValue(entity,
		dymessage.FromFloat64(510.972845))
	def.Fields[7].SetValue(entity,
		dymessage.FromBool(false))
	def.Fields[8].SetEntity(entity,
		dymessage.FromString("LJFzUzsO2O8auQAlVmJy"))
	def.Fields[9].SetEntity(entity,
		dymessage.FromBytes([]byte{24, 40, 107, 129, 64}, false))
	def.Fields[10].SetEntity(entity,
		dymessage.FromEntity(entity2))

	return def, entity
}

func AssertEncodeDecode(
	t *testing.T, def *dymessage.MessageDef, entity *dymessage.Entity) {
	ref := def.Fields[10].GetEntity(entity)
	require.NotNil(t, ref)
	linnet := ref.ToEntity()

	require.Equal(t, 3, def.Fields[20].Len(linnet))

	assert.Equal(t, int32(-33512104),
		def.Fields[1].GetValue(entity).ToInt32())
	assert.Equal(t, int64(-254715376635680503),
		def.Fields[2].GetValue(entity).ToInt64())
	assert.Equal(t, uint32(799283559),
		def.Fields[3].GetValue(entity).ToUint32())
	assert.Equal(t, uint64(65911047815132225),
		def.Fields[4].GetValue(entity).ToUint64())
	assert.Equal(t, float32(-204860.936),
		def.Fields[5].GetValue(entity).ToFloat32())
	assert.Equal(t, float64(510.972845),
		def.Fields[6].GetValue(entity).ToFloat64())

	assert.Equal(t, false,
		def.Fields[7].GetValue(entity).ToBool())
	assert.Equal(t, "LJFzUzsO2O8auQAlVmJy",
		def.Fields[8].GetEntity(entity).ToString())
	assert.Equal(t, []byte{24, 40, 107, 129, 64},
		def.Fields[9].GetEntity(entity).ToBytes())

	assert.Equal(t, int32(868929107),
		def.Fields[1].GetValue(linnet).ToInt32())
	assert.Equal(t, int64(-601380853565279092),
		def.Fields[2].GetValue(linnet).ToInt64())
	assert.Equal(t, uint32(783509315),
		def.Fields[3].GetValue(linnet).ToUint32())
	assert.Equal(t, uint64(54182615856980345),
		def.Fields[4].GetValue(linnet).ToUint64())
	assert.Equal(t, float32(80116.7676),
		def.Fields[5].GetValue(linnet).ToFloat32())
	assert.Equal(t, float64(1.2262663),
		def.Fields[6].GetValue(linnet).ToFloat64())

	assert.Equal(t, true,
		def.Fields[7].GetValue(linnet).ToBool())
	assert.Equal(t, "Zy0RVazdEe459Y0DErUJ",
		def.Fields[8].GetEntity(linnet).ToString())
	assert.Equal(t, []byte{232, 153, 178, 190, 4, 82},
		def.Fields[9].GetEntity(linnet).ToBytes())

	assert.Equal(t, int32(313261865),
		readArr(t, def.Fields[11]).getValueAt(linnet, 0).ToInt32())
	assert.Equal(t, int32(209295014),
		readArr(t, def.Fields[11]).getValueAt(linnet, 1).ToInt32())

	assert.Equal(t, int64(-394578838447094537),
		readArr(t, def.Fields[12]).getValueAt(linnet, 0).ToInt64())
	assert.Equal(t, int64(7197041536234632),
		readArr(t, def.Fields[12]).getValueAt(linnet, 1).ToInt64())

	assert.Equal(t, uint32(849851282),
		readArr(t, def.Fields[13]).getValueAt(linnet, 0).ToUint32())
	assert.Equal(t, uint32(248557441),
		readArr(t, def.Fields[13]).getValueAt(linnet, 1).ToUint32())

	assert.Equal(t, uint64(4416324197982829),
		readArr(t, def.Fields[14]).getValueAt(linnet, 0).ToUint64())
	assert.Equal(t, uint64(218233954665294213),
		readArr(t, def.Fields[14]).getValueAt(linnet, 1).ToUint64())

	assert.Equal(t, float32(9296232.53),
		readArr(t, def.Fields[15]).getValueAt(linnet, 0).ToFloat32())
	assert.Equal(t, float32(-54836.8569),
		readArr(t, def.Fields[15]).getValueAt(linnet, 1).ToFloat32())

	assert.Equal(t, float64(-682925.966251730668),
		readArr(t, def.Fields[16]).getValueAt(linnet, 0).ToFloat64())
	assert.Equal(t, float64(719704.153),
		readArr(t, def.Fields[16]).getValueAt(linnet, 1).ToFloat64())

	assert.Equal(t, true,
		readArr(t, def.Fields[17]).getValueAt(linnet, 0).ToBool())
	assert.Equal(t, false,
		readArr(t, def.Fields[17]).getValueAt(linnet, 1).ToBool())

	assert.Equal(t, "HN89fTSfx2it9Ma11Ufj",
		readArr(t, def.Fields[18]).getEntityAt(linnet, 0).ToString())
	assert.Equal(t, "f4nuZTeXQmsvR6MBPkC",
		readArr(t, def.Fields[18]).getEntityAt(linnet, 1).ToString())

	assert.Equal(t, []byte{189, 248, 87, 249, 19, 15},
		readArr(t, def.Fields[19]).getEntityAt(linnet, 0).ToBytes())
	assert.Equal(t, []byte{22, 72, 74, 121, 208},
		readArr(t, def.Fields[19]).getEntityAt(linnet, 1).ToBytes())
}

// -----------------------------------------------------------------------------
// Helper functions

// Returns a proto field, which represents a collection and checks for errors
// before returning a value at particular index.
func readArr(t *testing.T, f *dymessage.MessageFieldDef) *arrayReader {
	return &arrayReader{MessageFieldDef: f, t: t}
}

func (f *arrayReader) getValueAt(e *dymessage.Entity, n int) dymessage.Primitive {
	value, err := f.MessageFieldDef.GetValueAt(e, n)
	require.NoError(f.t, err)
	return value
}

func (f *arrayReader) getEntityAt(e *dymessage.Entity, n int) *dymessage.Reference {
	value, err := f.MessageFieldDef.GetEntityAt(e, n)
	require.NoError(f.t, err)
	return value
}
