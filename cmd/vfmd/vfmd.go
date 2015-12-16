package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"gopkg.in/akavel/vfmd.v0"
	"gopkg.in/akavel/vfmd.v0/mdblock"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		in  = flag.String("i", "-", "path to input Markdown document, or - for standard input")
		out = flag.String("o", "-", "path to output HTML document, or - for standard output")
		// TODO(akavel): tmpl = flag.String("t", "<!doctype html><html lang=en><head><meta charset=utf-8><title></title></head><body>\n{{.}}\n</body></html>", "template for the output HTML document") // see: http://www.brucelawson.co.uk/2010/a-minimal-html5-document/
	)
	var err error
	inf, outf := os.Stdin, os.Stderr
	if *in != "-" {
		inf, err = os.Open(*in)
		if err != nil {
			return err
		}
		defer inf.Close()
	}
	if *out != "-" {
		outf, err = os.Create(*out)
		if err != nil {
			return err
		}
		defer outf.Close()
	}

	prep, err := vfmd.QuickPrep(inf)
	if err != nil {
		return err
	}
	blocks, err := mdblock.QuickParse(bytes.NewReader(prep), mdblock.BlocksAndSpans, nil, nil)
	if err != nil {
		return err
	}
	err = vfmd.QuickHTML(outf, blocks)
	if err != nil {
		return err
	}
	return nil
}
