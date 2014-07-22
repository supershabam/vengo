package main

import (
	"fmt"
	"go/parser"
	"path/filepath"
	// "go/printer"
	"go/token"
	"os"
	"strings"
)

func main() {
	fset := token.NewFileSet() // positions are relative to fset

	if false {
		pkgm, err := parser.ParseDir(fset, "../../../../digitalocean/metrics-collector.git/*", nil, 0)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%+v\n", pkgm)
	}

	wf := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if strings.Contains(path, "/.git") {
				fmt.Printf("skipping %s\n", path)
				return filepath.SkipDir
			}
			return nil
		}
		if strings.Contains(path, ".go") {
			fmt.Println(path)
		}
		return nil
	}
	filepath.Walk("../../../../digitalocean/", wf)
}
