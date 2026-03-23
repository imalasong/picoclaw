//go:build ignore

package main

import (
	"log"
	"os"
	"path/filepath"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	src := filepath.Clean(filepath.Join(wd, "..", "..", "..", "..", "workspace"))
	st, err := os.Stat(src)
	if err != nil || !st.IsDir() {
		log.Fatalf("workspace source missing or not a directory: %s (%v)", src, err)
	}
	if err := os.RemoveAll("workspace"); err != nil {
		log.Fatal(err)
	}
	if err := os.CopyFS("workspace", os.DirFS(src)); err != nil {
		log.Fatal(err)
	}
}
