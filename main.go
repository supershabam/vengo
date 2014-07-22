package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	// base = "github.com/supershabam/vengo/"
	base   string
	target string
)

func init() {
	flag.StringVar(&base, "base", "", "name of this package e.g. github.com/supershabam/vengo")
	flag.StringVar(&target, "target", "", "name of package to install e.g. github.com/gorilla/mux")
	flag.Parse()

	if base == "" {
		log.Fatalf("base must be provided")
	}
	if target == "" {
		log.Fatalf("target must be provided")
	}
}

func isStdlib(path string) bool {
	return !strings.Contains(path, ".")
}

func ensurePrefix(path, base string) string {
	if strings.Contains(path, base) {
		return path
	}
	return strings.Replace(path, "\"", fmt.Sprintf("\"%s/", base), 1)
}

func rebase(dir, base string) (first error) {
	walkFn := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			// skip the .git directory
			if strings.Contains(path, "/.git") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.Contains(path, ".go") {
			rewrite(path, base)
		}
		return nil
	}
	return filepath.Walk(dir, walkFn)
}

func rewrite(path, base string) {
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
		s.Path.Value = ensurePrefix(s.Path.Value, base)
	}

	file, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	printer.Fprint(file, fset, f)
}

func vengo(target, base string) (first error) {
	gitURL := fmt.Sprintf("https://%s.git", target)
	vendir := fmt.Sprintf("./vendor/%s", target)
	// ensure vendir
	if err := os.MkdirAll(vendir, 0777); err != nil {
		return err
	}
	// clean vendir
	cmd := exec.Command("rm", "-fr", vendir)
	if err := cmd.Run(); err != nil {
		return err
	}
	// clone into vendir
	cmd = exec.Command("git", strings.Split(fmt.Sprintf("clone --depth=1 %s %s", gitURL, vendir), " ")...)
	if err := cmd.Run(); err != nil {
		return err
	}
	// un-gitify
	cmd = exec.Command("rm", "-fr", fmt.Sprintf("%s/.git", vendir))
	if err := cmd.Run(); err != nil {
		return err
	}
	// rewrite cloned files
	return rebase(vendir, base)
}

func main() {
	if err := vengo(target, base); err != nil {
		panic(err)
	}
}
