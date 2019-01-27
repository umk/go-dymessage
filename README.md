## Dynamic messages for Go

[![GoDoc](https://godoc.org/github.com/umk/go-dymessage?status.svg)](https://godoc.org/github.com/umk/go-dymessage)
[![Go Report Card](https://goreportcard.com/badge/github.com/umk/go-dymessage)](https://goreportcard.com/report/github.com/umk/go-dymessage)

The package provides the structures and functionality to maintain the entities with dynamic structure and serializing them to Protocol Buffers and JSON formats. 

The basic structure is the `Entity`, which stores the data:

```go
type Entity struct {
    DataType DataType

    Data     []byte
    Entities []*Entity
}
```

According to the structure, the `Entity` only maintains the data, leaving the definition of its structure up to the caller. The `DataType` property is meant to give caller a notion, how should it treat the entity data.

The `Data` and `Entities` properties contain respectively the data of the primitive types and the reference types. Upon the creation, the instance of `Entity` must contain the placeholders for the its data, according to the entity structure. If entity contains a collection of primitive or reference values, each of these collections must be represented by a separate `Entity`, which contains either `Data` or `Entities` property populated with the items of collection.

Few rules may apply when planning the structure of entities:

 * In general, the recursive references are supported, but consider how the entity will be used, as neither JSON, nor Protocol Buffers provide support the recursive structures.
 * The `Entity` structure can represent a collection of collections, but the dynamic message library will not provide support for this for the sake of simplicity. Consider using nested entities for this purpose or [representation by a vector](https://en.wikipedia.org/wiki/Matrix_representation), if operating matrices.

The good practice would be reuse of the `Entity` instances, as maintaining the data requires memory allocations. Use the `Reset` method to clean up the data before the entity is made available for reuse.

### Registry

The out of the box approach of maintaining the structure of an entity is to use the `Registry` and `MessageDef` types. The `Registry` aggregates all known message definitions, and `MessageDef` in turn represents the message definition and provides the methods to operate the properties. The following is an example how to create an instance of `Registry`:

```go
// Each of the entity properties must have its unique tag, but several
// entities may share the tags for its own properties. This way the
// following constants represent the tags for properties with specific
// name.
const (
	TagX uint64 = iota
	TagY
	TagZ
	TagPoints
)

builder := dymessage.NewRegistryBuilder()

dtPoint2d := builder.ForMessageDef("2d point").GetDataType()
dtPoint2dVector := builder.ForMessageDef("2d point vector").GetDataType()

builder.ForMessageDef("2d point").
	WithName("Point2D").
	WithField("X", TagX, dymessage.DtInt32).
	WithField("Y", TagY, dymessage.DtInt32).
	Build()

// The Point3D entity shares the same tags with Point2D for the
// properties of the same name.
builder.ForMessageDef("3d point").
	WithName("Point3D").
	WithField("X", TagX, dymessage.DtInt32).
	WithField("Y", TagY, dymessage.DtInt32).
	WithField("Z", TagZ, dymessage.DtInt32).
	Build()

builder.ForMessageDef("2d point vector").
	WithName("Point2DVector").
	WithArrayField("Points", TagPoints, dtPoint2d).
	Build()

registry := builder.Build()
```

The following is an example of how the registry may be used. This builds an 2D point vector, that consists of a single point:

```go
defPoint2d := registry.GetMessageDef(dtPoint2d)
defPoint2dVector := registry.GetMessageDef(dtPoint2dVector)

// NewEntity will allocate necessary memory in the entity to accomodate
// its properties.
point := defPoint2d.NewEntity()
defPoint2d.GetField(TagX).SetPrimitive(point, dymessage.FromInt32(42))
defPoint2d.GetField(TagY).SetPrimitive(point, dymessage.FromInt32(123))

vector := defPoint2dVector.NewEntity()
// This will grow the vector by specified number of items. The returned
// value is the number of items before the collection has grown.
n := defPoint2dVector.GetField(TagPoints).Reserve(vector, 1)
defPoint2dVector.GetField(TagPoints).
	SetReferenceAt(vector, n, dymessage.FromEntity(point))
```

And this will restore the content of the entities and print it to the standard output:

```go
fmt.Println(defPoint2dVector.GetField(TagPoints).Len(vector)) // output: 1

actualPoint := defPoint2dVector.GetField(TagPoints).
	GetReferenceAt(vector, n).ToEntity()

fmt.Println(defPoint2d.GetField(TagX).GetPrimitive(actualPoint).ToInt32()) // output: 42
fmt.Println(defPoint2d.GetField(TagY).GetPrimitive(actualPoint).ToInt32()) // output: 123
```

The infrastrucure deliberately won't maintain correctness of usage of the properties by tracking whether the property is assigned with the value of correct type; or whether the user goes out of the collection range. If something unexpected occurs, the application will just panic.

:construction: :construction: :construction:

### Benchmarks

The benchmarks include encoding and decoding of the entity, which contains all variety of the fields, that a message definition can represent: primitive types, reference to strings, byte arrays and other entities, collections of values of primitive and reference types. 

:warning: These benchmarks represent only performance of encoding and decoding the entities. Reading or writing values to entity, which is being encoded or decoded, will require additional CPU time.

#### JSON

The benchmarks of encoding and decoding of an entity to and from JSON:

```
BenchmarkTestEncodeRegular/encode_regular-8         	   20000	     84517 ns/op	   13956 B/op	     381 allocs/op
BenchmarkTestDecodeRegular/decode_regular-8         	   20000	     65559 ns/op	    6648 B/op	     384 allocs/op
```

The benchmark of iterating through all of the tokens using `json.Decoder`: 

```
BenchmarkReference/json.Decoder-8                   	   10000	    166830 ns/op	   31936 B/op	    1840 allocs/op
```

The benchmark of decoding the JSON to map, which would contain the mapping from the name of property to its value:

```
BenchmarkReference/json.Unmarshal-8                 	   20000	     82786 ns/op	   19197 B/op	     345 allocs/op
```

#### Protocol Buffers

The benchmarks of encoding and decoding of an entity to and from Protocol Buffers:

```
BenchmarkEncodeRegular/encode_regular-8         	  300000	      5499 ns/op	    1320 B/op	      17 allocs/op
BenchmarkDecodeRegular/decode_regular-8         	  200000	      8936 ns/op	    3064 B/op	      69 allocs/op
```

The same, but in multiple threads:

```
BenchmarkEncodeParallel-8                                	  300000	      3954 ns/op	    1322 B/op	      17 allocs/op
BenchmarkDecodeParallel-8                                	  300000	      5200 ns/op	    3065 B/op	      69 allocs/op
BenchmarkEncodeDecodeParallel-8                          	  200000	      9465 ns/op	    4398 B/op	      86 allocs/op
```

The benchmark of decoding into an existing entity, which avoids memory allocations:

```
BenchmarkDecodeRegular/decode_regular_existing-8         	  200000	      6253 ns/op	    1064 B/op	      22 allocs/op
```

For comparison, benchmarks of encoding and decoding the entity by the [reference implementation](https://github.com/golang/protobuf) of the Protocol Buffers:

```
BenchmarkReferenceEncode/proto.Marshal-8                 	  200000	      6698 ns/op	    2728 B/op	      25 allocs/op
BenchmarkReferenceDecode/proto.Unmarshal-8               	  200000	      6495 ns/op	    2568 B/op	      33 allocs/op
```
