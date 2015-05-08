package vfmd

import (
	"bytes"
	"unicode/utf8"
)
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
	preproCR
)

type Chunk struct {
	Bytes []byte
	Type  ChunkType
}

type ChunkType int

const (
	ChunkIgnoredBOM ChunkType = iota
	ChunkUnchangedRunes
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
	const (
		CR = '\r'
		LF = '\n'
	)

	// ignore Byte-Order-Mark [#document]
	// TODO(akavel): add more tests
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

	// Flush old state if pending
	if b != LF && p.state == preproCR {
		p.state = preproNormal
		p.normalChunk(CR)
	}

	// Detect invalid UTF-8 and assume it's ISO-8859-1 then. [#document]
	// TODO(akavel): add tests
	if b > utf8.RuneSelf {
		// May be an UTF-8 encoded rune; or assume ISO-8859-1 if invalid.
		p.Pending = append(p.Pending, b)
		if !utf8.FullRune(p.Pending) {
			// Still cannot say if valid or invalid rune
			return nil
		}
		buf := p.Pending
		p.Pending = nil
		r, _ := utf8.DecodeRune(buf)
		if r == utf8.RuneError {
			// TODO(akavel): add chunk marking start of invalid utf8
			p.Write(iso2utf(buf...))
			// TODO(akavel): add chunk marking end of invalid utf8
		} else {
			p.normalChunk(buf...)
		}
		return nil
	}
	// If any bytes are pending, they failed to build a complete UTF-8 rune; flush them
	if len(p.Pending) > 0 {
		buf := p.Pending
		p.Pending = nil
		// TODO(akavel): add chunk marking start of invalid utf8
		p.Write(iso2utf(buf...))
		// TODO(akavel): add chunk marking end of invalid utf8
	}

	// TODO(akavel): convert CRLF to LF [#characters]
	switch {
	case b == LF && p.state == preproCR:
		// CRLF detected
		p.state = preproNormal
		p.Chunks = append(p.Chunks, Chunk{
			Bytes: []byte{LF},
			Type:  ChunkNormalizedCRLF,
		})
		return nil
	case b == CR:
		p.state = preproCR
		return nil
	}

	// TODO(akavel): expand tabs to spaces [#lines]

	// TODO(akavel): WIP
	p.normalChunk(b)
	return nil
}

func (p *Preprocessor) normalChunk(b ...byte) {
	n := len(p.Chunks)
	if n == 0 || p.Chunks[n-1].Type != ChunkUnchangedRunes {
		p.Chunks = append(p.Chunks, Chunk{Type: ChunkUnchangedRunes})
		n++
	}
	p.Chunks[n-1].Bytes = append(p.Chunks[n-1].Bytes, b...)
}

func iso2utf(buf ...byte) []byte {
	out := make([]byte, 0, 2*len(buf))
	for _, b := range buf {
		r := rune(b)
		n := utf8.RuneLen(r) // 1 or 2
		pos := len(out)
		out = out[:pos+n]
		utf8.EncodeRune(out[pos:], r)
	}
	return out
}
