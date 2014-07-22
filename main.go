package main

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	basepath = "github.com/supershabam/vengo/"
)

func isStdlib(path string) bool {
	return !strings.Contains(path, ".")
}

func ensurePrefix(path string) string {
	if strings.Contains(path, basepath) {
		return path
	}
	return strings.Replace(path, "\"", fmt.Sprintf("\"%s", basepath), 1)
}

func rewrite(path string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, s := range f.Imports {
		if isStdlib(s.Path.Value) {
			continue
		}
		s.Path.Value = ensurePrefix(s.Path.Value)
	}

	file, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	printer.Fprint(file, fset, f)
}

func main() {
	if false {
		cmd := exec.Command("git", strings.Split("clone --depth=1 https://github.com/gorilla/mux.git ./vendor/github.com/gorilla/mux", " ")...)
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err := cmd.Run()
		out := buf.Bytes()
		if err != nil {
			fmt.Printf("%s\n", out)
		}
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			// skip the .git directory
			if strings.Contains(path, "/.git") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.Contains(path, ".go") {
			rewrite(path)
		}
		return nil
	}
	filepath.Walk("./vendor/github.com/gorilla/mux", walkFn)
}
