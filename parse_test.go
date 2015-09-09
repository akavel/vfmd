package vfmd

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

var dirs = []string{
	`testdata/tests/span_level`,
	`testdata/tests/block_level`,
	`testdata/tests/text_processing`,
}

func TestParse(test *testing.T) {
	files := []string{}
	for _, dir := range dirs {
		m, err := filepath.Glob(filepath.FromSlash(dir + "/*/*.md"))
		if err != nil {
			panic(err)
		}
		files = append(files, m...)
	}
	for _, f := range files {
		fh, err := os.Open(f)
		if err != nil {
			panic(err)
		}
		blocks, err := Parse(fh)
		fh.Close()
		if err != nil {
			test.Error(err)
		}
		fmt.Println(blocks)
		return // TMP
	}
}
