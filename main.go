package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func main() {
	cmd := exec.Command("git clone --depth=1 https://github.com/gorilla/mux.git /Users/supershabam/Code/supershabam/vengo/vendor/github.com/gorilla/mux")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	out := buf.Bytes()
	if err != nil {
		panic(err)
	}
	fmt.Print(out)
}
