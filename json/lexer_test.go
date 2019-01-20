package json

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/umk/go-testutil"
)

func TestLexer(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		panic("could not get working directory")
	}
	for _, dir := range []string{"positive", "negative", "indefinite"} {
		path := filepath.Join(root, "internal/testdata", dir)
		var paths []string
		_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if info != nil && !info.IsDir() && filepath.Ext(path) == ".json" {
				paths = append(paths, path)
			}
			return nil
		})
		t.Logf("got %d file(s) for \"%s\" test", len(paths), dir)
		for _, path := range paths {
			path := path
			name := dir + "_" + filepath.Base(path)
			t.Run(name, func(t *testing.T) {
				testSourceFile(t, path, dir == "positive")
			})
		}
	}
}

func testSourceFile(t *testing.T, path string, positive bool) {
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var builder strings.Builder
	err = createLexerOutput(f, &builder)
	if positive && err != nil {
		t.Fatal(err)
	}
	outputPath := path + ".lex.txt"
	if testutil.DoFix() {
		f, err := os.Create(outputPath)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		_, _ = f.WriteString(builder.String())
	}
	fout, err := ioutil.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	expected := string(fout)
	actual := builder.String()
	testutil.EqualDiff(t, expected, actual, path)
}

func createLexerOutput(f *os.File, out io.Writer) (err error) {
	var lex lexer
	lex.reader.reset(f)
	for {
		lex.next()
		if lex.err != nil {
			_, err = fmt.Fprintf(out, "\nERROR: %s", lex.err.Error())
			break
		}
		if lex.eof() {
			break
		}
		switch lex.tok.kind {
		case tkString:
			_, err = fmt.Fprintf(out, "%q", lex.tok.value)
		case tkNumber:
			_, err = fmt.Fprint(out, lex.tok.value)
		default:
			_, err = fmt.Fprint(out, lex.tok.kind)
		}
		if err != nil {
			panic(err)
		}
	}
	return
}
