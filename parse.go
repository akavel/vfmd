package vfmd

import (
	"io"

	"gopkg.in/akavel/vfmd.v0/block"
	"gopkg.in/akavel/vfmd.v0/span"
)

type Block struct {
	block.Block
	Spans    []span.Span
	Children []Block
}

func Parse(r io.Reader) ([]Block, error) {
	readbuf := make([]byte, 32*1024)
	lines := []block.Line{}
	line := []byte{}
	prep := Preprocessor{}
	blocks := block.Splitter{}
	result := []Block{}
	i := 0
	for {
		// Preprocess.
		n, err := r.Read(readbuf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		prep.Write(readbuf[:n])
		if err == io.EOF {
			prep.Close()
		}

		// Pass preprocessed input to block splitter, split into lines.
		for i < len(prep.Chunks) {
			if prep.Chunks[i].Type == ChunkUnchangedLF {
				line = buildLine(line, prep.Chunks[:i])
				lines = append(lines, line)
				prep.Chunks = prep.Chunks[i+1:]
				i = 0
				err2 := blocks.WriteLine(line)
				if err2 != nil {
					return nil, err2
				}
				continue
			}
			i++
		}
		if err == io.EOF {
			if i > 0 {
				line = buildLine(line, prep.Chunks)
				lines = append(lines, line)
				err2 := blocks.WriteLine(line)
				if err2 != nil {
					return nil, err2
				}
			}
			err2 := blocks.Close()
			if err2 != nil {
				return nil, err2
			}
		}

		// Extract detected blocks & post-process them.
		result = append(result, postProcess(blocks, lines)...)
		blocks.Blocks = nil

		if err == io.EOF {
			return result, nil
		}
	}
}

// FIXME(akavel): that looks too weird and hackish
type blockser interface {
	GetBlocks() block.Blocks
}
type spanser interface {
	GetSpans() block.Spans
}

func postProcess(b blockser, lines []block.Line) []Block {
	bs := b.GetBlocks()
	result := make([]Block, 0, len(bs))
	for _, b := range bs {
		b := Block{Block: b}
		if _, ok := b.Detector.(block.Null); ok {
			result = append(result, b)
			continue
		}
		// spew.Dump(b)
		// spew.Printf("%q\n", lines)
		for _, line := range lines[b.First : b.Last+1] {
			b.PostProcess(line)
		}
		b.PostProcess(nil)
		if bl, ok := b.Detector.(blockser); ok {
			b.Children = postProcess(bl, lines)
		}
		if sp, ok := b.Detector.(spanser); ok {
			buf := []byte{}
			for _, s := range sp.GetSpans() {
				buf = append(buf, s...)
				buf = append(buf, '\n')
			}
			b.Spans = span.Parse(buf, nil)
		}
		result = append(result, b)
	}
	return result
}

func buildLine(buf []byte, chunks []Chunk) []byte {
	if len(buf) > 0 {
		buf = buf[:0]
	}
	for _, c := range chunks {
		buf = append(buf, c.Bytes...)
	}
	return buf
}
