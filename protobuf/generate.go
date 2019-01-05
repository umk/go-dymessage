package protobuf

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/template"

	. "github.com/umk/go-dymessage"
	. "github.com/umk/go-fslayer"
	"github.com/umk/go-stringutil"
)

type (
	// Provides methods to locate where the exported protocol definition
	// files will be located. See the ExportProto function for details.
	ExportLocator interface {
		// CreateWriter should create a writer for definitions of the message, found
		// at specified namespace in the repository. After the writer has been
		// returned, the caller is supposed to close it, if necessary.
		CreateWriter(ns string) (io.Writer, error)
		// GetImport gets a relative path to the file that implements messages
		// declared in provided namespace.
		GetImport(ns string) string
	}

	// Implements a location that puts the created proto definitions into
	// specified directory.
	FileSystemLocator struct {
		root    string
		flatten bool
	}
)

var defaultProtoTypes = map[DataType]string{
	DtInt32:   "sfixed32",
	DtInt64:   "sfixed64",
	DtUint32:  "fixed32",
	DtUint64:  "fixed64",
	DtFloat32: "float",
	DtFloat64: "double",
	DtBool:    "bool",
	DtString:  "string",
	DtBytes:   "bytes",
}

var zigZagProtoTypes = map[DataType]string{
	DtInt32: "sint32",
	DtInt64: "sint64",
}

var varintProtoTypes = map[DataType]string{
	DtInt32:  "int32",
	DtInt64:  "int64",
	DtUint32: "uint32",
	DtUint64: "uint64",
}

// -----------------------------------------------------------------------------
// Locators

// NewFileSystemLocator creates a file locator that puts the message definitions
// into the file system. The flatten parameter indicates whether files from all
// namespaces must be put into a single directory. Otherwise they are located in
// a hierarchy according to entries of namespace.
func NewFileSystemLocator(root string, flatten bool) *FileSystemLocator {
	return &FileSystemLocator{
		root:    root,
		flatten: flatten,
	}
}

func (f *FileSystemLocator) CreateWriter(ns string) (io.Writer, error) {
	fp := filepath.Join(f.root, f.GetImport(ns))
	if f, err := Fs().Create(fp); err == nil {
		return f, nil
	} else {
		return nil, err
	}
}

func (f *FileSystemLocator) GetImport(ns string) string {
	name := ns + ".proto"
	if f.flatten {
		return name
	}
	entries := append(strings.Split(ns, "."), name)
	return filepath.Join(entries...)
}

// -----------------------------------------------------------------------------
// Export

// ExportProto transforms the whole registry into the .proto files, so the
// clients of the application, which use the dynamic messages library, could
// generate their own sources on their favorite languages and then communicate
// with the application.
func ExportProto(r *Registry, loc ExportLocator) error {
	files := make(map[string]map[string]*MessageDef)
	for _, def := range r.Defs {
		p, ok := files[def.Namespace]
		if !ok {
			p = make(map[string]*MessageDef)
			files[def.Namespace] = p
		}
		if _, ok = p[def.Name]; ok {
			return fmt.Errorf("duplicate name %s at %s", def.Name, def.Namespace)
		}
		p[def.Name] = def
	}
	for ns, defs := range files {
		err := export(r, ns, defs, loc)
		if err != nil {
			return err
		}
	}
	return nil
}

func export(r *Registry, namespace string, defs map[string]*MessageDef, loc ExportLocator) error {
	imports := make(map[string]interface{})
	for _, def := range defs {
		for _, f := range def.Fields {
			if f.Tag == 0 || f.Tag >= 19000 && f.Tag < 20000 {
				return fmt.Errorf("tag %v is out of range", f.Tag)
			}
			if (f.DataType & DtEntity) != 0 {
				dt := r.GetMessageDef(f.DataType)
				if dt.Namespace != namespace {
					imports[dt.Namespace] = nil
				}
			}
		}
	}
	wr, err := loc.CreateWriter(namespace)
	if err != nil {
		return err
	}
	defer func() {
		if closer, ok := wr.(io.Closer); ok {
			_ = closer.Close()
		}
	}()
	return createTemplate(r, namespace, loc).Execute(wr, struct {
		Ns      string
		Imports map[string]interface{}
		Defs    map[string]*MessageDef
	}{
		Ns:      namespace,
		Imports: imports,
		Defs:    defs,
	})
}

func getBuiltInTypeName(f *MessageFieldDef) string {
	extension, ok := tryGetExtension(f)
	if ok && extension.integerKind != ikDefault {
		ik := extension.integerKind
		switch ik {
		case ikZigZag:
			return zigZagProtoTypes[f.DataType]
		case ikVarint:
			return varintProtoTypes[f.DataType]
		default:
			panic(fmt.Sprintf("unsupported value of integer kind %d", ik))
		}
	} else {
		if name, ok := defaultProtoTypes[f.DataType]; ok {
			return name
		}
		panic(fmt.Sprintf("unable to determine name of the type %d", f.DataType))
	}
}

func createTemplate(reg *Registry, ns string, loc ExportLocator) *template.Template {
	return template.Must(
		template.New("protodef").Funcs(template.FuncMap{
			"typename": func(f *MessageFieldDef) string {
				if (f.DataType & DtEntity) != 0 {
					t := reg.GetMessageDef(f.DataType)
					if ns == t.Namespace {
						return t.Name
					} else {
						return "." + t.Namespace + "." + t.Name
					}
				} else {
					return getBuiltInTypeName(f)
				}
			},
			"fieldname": func(s string) string {
				return strings.ToLower(stringutil.SnakeCaps(s))
			},
			"modifier": func(f *MessageFieldDef) string {
				if f.Repeated {
					return "repeated "
				}
				return ""
			},
			"import": func(ns string) string {
				return loc.GetImport(ns)
			},
		}).Delims("<", ">").Parse(`syntax = "proto3";

package < .Ns >;

< range $index, $element := .Imports >import "< import $index >";
< end >< range .Defs >
message < .Name >
{< range .Fields >
	< modifier . >< typename . > < fieldname .Name > = < .Tag >;
< end >}
< end >`))
}
