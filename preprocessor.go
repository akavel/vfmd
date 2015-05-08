package vfmd

import "bytes"
import "io"

type Preprocessor struct {
	Chunks  []Chunk
	Pending []byte
	state   int
}

// Preprocessor states
const (
	preproMaybeBOM = iota
	preproNormal
)

type Chunk struct {
	Bytes []byte
	Type  ChunkType
}

type ChunkType int

const (
	ChunkIgnoredBOM ChunkType = iota
	ChunkUnchangedBytes
)

// statically ensure that certain interfaces are implemented by Preprocessor
var _ io.Writer = &Preprocessor{}

func (p *Preprocessor) Write(buf []byte) (int, error) {
	for _, b := range buf {
		p.WriteByte(b)
	}
	return len(buf), nil
}

func (p *Preprocessor) WriteByte(b byte) error {
	// ignore Byte-Order-Mark [#document]
	// TODO(akavel): add test
	if p.state == preproMaybeBOM {
		BOM := []byte{0xEF, 0xBB, 0xBF}
		p.Pending = append(p.Pending, b)
		switch {
		case !bytes.HasPrefix(BOM, p.Pending):
			p.state = preproNormal
			buf := p.Pending
			p.Pending = nil
			p.Write(buf)
			return nil
		case len(p.Pending) == len(BOM):
			p.Chunks = append(p.Chunks, Chunk{
				Bytes: p.Pending,
				Type:  ChunkIgnoredBOM,
			})
			p.Pending = nil
			p.state = preproNormal
			return nil
		default: // still not sure
			return nil
		}
	}
	// TODO(akavel): convert ISO-8859-1 to UTF-8 [#document]
	// TODO(akavel): convert CRLF to LF [#characters]
	// TODO(akavel): expand tabs to spaces [#lines]

	// TODO(akavel): WIP
	n := len(p.Chunks)
	if n == 0 || p.Chunks[n-1].Type != ChunkUnchangedBytes {
		p.Chunks = append(p.Chunks, Chunk{Type: ChunkUnchangedBytes})
		n++
	}
	p.Chunks[n-1].Bytes = append(p.Chunks[n-1].Bytes, b)
	return nil
}
