package vfmd

import (
	"fmt"
	"io"

	"github.com/davecgh/go-spew/spew"

	"gopkg.in/akavel/vfmd.v0/block"
	"gopkg.in/akavel/vfmd.v0/span"
)

func HTML(blocks []Block, w io.Writer) {
	for _, b := range blocks {
		htmlbb(b, w)
	}
}

func htmlb(b Block, w io.Writer) {
	for _, b := range b.Children {
		htmlbb(b, w)
	}
}

func htmlbb(b Block, w io.Writer) {
	wr := func(s string) { w.Write([]byte(s)) }
	switch d := b.Detector.(type) {
	case *block.Paragraph:
		wr(`<p>`)
		htmls(b, w)
		wr(`</p>`)
		_ = d
	case *block.UnorderedList:
		wr(`<ul>`)
		htmlb(b, w)
		spew.Dump(b)
		wr(`</ul>`)
	default:
		panic(fmt.Sprintf("unsupported block type: %T", b.Detector))
	}
}

func htmls(b Block, w io.Writer) {
	for _, s := range b.Spans {
		htmlss(s, w)
	}
}

func htmlss(s span.Span, w io.Writer) {
	wr := func(s string) { w.Write([]byte(s)) }
	switch t := s.Tag.(type) {
	case span.Emphasis:
		switch t.Level {
		case 1:
			wr("<em>")
		default:
			wr("<strong>")
		}
	default:
		panic(fmt.Sprintf("unsupported tag type: %T", s.Tag))
	}
}
