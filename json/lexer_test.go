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
	errlex, err := createLexerOutput(f, &builder)
	if err != nil {
		t.Fatal(err)
	}
	if positive && errlex != nil {
		t.Fatal(errlex)
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

func createLexerOutput(f *os.File, out io.Writer) (errlex error, err error) {
	var lex lexer
	lex.reader.reset(f)
	for {
		errlex = lex.next()
		if errlex != nil {
			if errlex == io.EOF {
				errlex = nil
			} else {
				_, err = fmt.Fprintf(out, "\nERROR: %s", errlex.Error())
			}
			break
		}
		switch lex.tok.kind {
		case tokString:
			_, err = fmt.Fprintf(out, "%q", lex.tok.string)
		case tokNumber:
			_, err = fmt.Fprint(out, lex.tok.number)
		case tokBool:
			_, err = fmt.Fprintf(out, "%t", lex.tok.bool)
		case tokCrBrOpen:
			_, err = fmt.Fprint(out, "{")
		case tokCrBrClose:
			_, err = fmt.Fprint(out, "}")
		case tokSqBrOpen:
			_, err = fmt.Fprint(out, "[")
		case tokSqBrClose:
			_, err = fmt.Fprint(out, "]")
		case tokColon:
			_, err = fmt.Fprint(out, ":")
		case tokComma:
			_, err = fmt.Fprint(out, ",")
		case tokNull:
			_, err = fmt.Fprint(out, "null")
		}
		if err != nil {
			break
		}
	}
	return
}
