package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
)

var (
	basepath string
)

func init() {
	flag.StringVar(&basepath, "basepath", "ass", "path to ensure exists prefixed to imports")
	flag.Parse()
}

func isStdlib(path string) bool {
	return !strings.Contains(path, ".")
}

func ensurePrefix(path string) string {
	if strings.Contains(path, basepath) {
		return path
	}
	return strings.Replace(path, "\"", fmt.Sprintf("\"%s", basepath), 1)
}

func main() {
	fset := token.NewFileSet() // positions are relative to fset

	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseFile(fset, "../../../../digitalocean/metrics-collector.git/cmd/collector/main.go", nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, s := range f.Imports {
		if isStdlib(s.Path.Value) {
			fmt.Printf("%s is stdlib\n", s.Path.Value)
			continue
		}
		s.Path.Value = ensurePrefix(s.Path.Value)
	}

	var buf bytes.Buffer
	printer.Fprint(&buf, fset, f)
	fmt.Println(buf.String())
}
