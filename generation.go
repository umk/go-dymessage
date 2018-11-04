package dymessage

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/template"

	. "github.com/umk/go-fslayer"
	"github.com/umk/go-stringutil"
)

type (
	Exporter interface {
		// CreateWriter should create a writer for definitions of the message, found
		// at specified namespace in the repository. After the writer has been
		// returned, the caller is supposed to close it, if necessary.
		CreateWriter(ns string) (io.Writer, error)
		// GetImport gets a relative path to the file that implements messages
		// declared in provided namespace.
		GetImport(ns string) string
	}

	// Implements an exporter that puts the created proto definitions into
	// specified directory.
	FileExporter struct {
		root string
		flat bool
	}
)

var protoTypes = map[DataType]string{
	DtInt32:   "int32",
	DtInt64:   "int64",
	DtUint32:  "uint32",
	DtUint64:  "uint64",
	DtFloat32: "float",
	DtFloat64: "double",
	DtBool:    "bool",
	DtString:  "string",
	DtBytes:   "bytes",
}

// -----------------------------------------------------------------------------
// Exporters

// ExportToFiles creates a file exporter that puts the message definitions into
// the file system. The flat parameter indicates whether files from all
// namespaces must be put into a single directory. Otherwise they are located in
// a hierarchy according to entries of namespace.
func ExportToFiles(root string, flat bool) *FileExporter {
	return &FileExporter{
		root: root,
		flat: flat,
	}
}

func (f *FileExporter) CreateWriter(ns string) (io.Writer, error) {
	fp := filepath.Join(f.root, f.GetImport(ns))
	if f, err := Fs().Create(fp); err == nil {
		return f, nil
	} else {
		return nil, err
	}
}

func (f *FileExporter) GetImport(ns string) string {
	name := ns + ".proto"
	if f.flat {
		return name
	}
	entries := append(strings.Split(ns, "."), name)
	return filepath.Join(entries...)
}

// -----------------------------------------------------------------------------
// Exporting

func (r *ProtoRepo) Export(xp Exporter) error {
	files := make(map[string]map[string]*ProtoDef)
	for _, def := range r.defs {
		p, ok := files[def.Ns]
		if !ok {
			p = make(map[string]*ProtoDef)
			files[def.Ns] = p
		}
		if _, ok = p[def.Name]; ok {
			return fmt.Errorf("duplicate name %v at %v", def.Name, def.Ns)
		}
		p[def.Name] = def
	}
	for ns, defs := range files {
		err := r.exportNs(ns, defs, xp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ProtoRepo) exportNs(
	ns string, defs map[string]*ProtoDef, xp Exporter) error {
	imports := make(map[string]interface{})
	for _, def := range defs {
		for _, f := range def.Fields {
			if f.Tag == 0 || f.Tag >= 19000 && f.Tag < 20000 {
				return fmt.Errorf("tag %v is out of range", f.Tag)
			}
			if (f.DataType & DtEntity) != 0 {
				dt := r.defs[f.DataType&^DtEntity]
				if dt.Ns != ns {
					imports[dt.Ns] = nil
				}
			}
		}
	}
	wr, err := xp.CreateWriter(ns)
	if err != nil {
		return err
	}
	defer func() {
		if closer, ok := wr.(io.Closer); ok {
			closer.Close()
		}
	}()
	return createTemplate(r, ns, xp).Execute(wr, struct {
		Ns      string
		Imports map[string]interface{}
		Defs    map[string]*ProtoDef
	}{
		Ns:      ns,
		Imports: imports,
		Defs:    defs,
	})
}

func createTemplate(
	repo *ProtoRepo, ns string, xp Exporter) *template.Template {
	return template.Must(
		template.New("protodef").Funcs(template.FuncMap{
			"typename": func(f *ProtoField) string {
				if name, ok := protoTypes[f.DataType]; ok {
					return name
				}
				if (f.DataType & DtEntity) != 0 {
					t := repo.defs[f.DataType&^DtEntity]
					if ns == t.Ns {
						return t.Name
					} else {
						return "." + t.Ns + "." + t.Name
					}
				} else {
					panic("unknown data type")
				}
			},
			"fieldname": func(s string) string {
				return strings.ToLower(stringutil.SnakeCaps(s))
			},
			"modifier": func(f *ProtoField) string {
				if f.Repeated {
					return "repeated "
				}
				return ""
			},
			"import": func(ns string) string {
				return xp.GetImport(ns)
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
