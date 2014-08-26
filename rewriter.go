package vengo

import (
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

type Rewriter struct {
	Base string
}

func (r Rewriter) Rewrite(path string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return err
	}

	for _, s := range f.Imports {
		if isSTDLib(s.Path.Value) {
			continue
		}
		s.Path.Value = ensurePrefix(s.Path.Value, r.Base)
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	printer.Fprintf(file, fset, f)

	return nil
}

func isSTDLib(path string) bool {
	return !strings.Contains(path, ".")
}

func ensurePrefix(path, base string) string {
	if strings.Contains(path, base) {
		return path
	}
	return strings.Replace(path, "\"", fmt.Sprintf("\"%s/", base), 1)
}
