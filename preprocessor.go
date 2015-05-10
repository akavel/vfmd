package vfmd

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"
)

type Preprocessor struct {
	Chunks  []Chunk
	Pending []byte
	state   int
	column  int
}

// Preprocessor states
const (
	preproMaybeBOM = iota
	preproNormal
	preproCR
	preproISO8859_1
)

type Chunk struct {
	Type  ChunkType
	Bytes []byte
}

// SourceLength returns length of the byte slice in the original stream that
// was converted into this Chunk.
func (c Chunk) SourceLength() int {
	switch c.Type {
	case ChunkIgnoredBOM:
		return 3
	case ChunkUnchangedRunes:
		return len(c.Bytes)
	case ChunkNormalizedCRLF:
		return 2
	case ChunkExpandedTab:
		return 1
	case ChunkConvertedISO8859_1:
		return utf8.RuneCount(c.Bytes)
	}
	panic(fmt.Sprintf("unknown vfmd.Chunk type %d [%q]", c.Type, c.Bytes))
}

type ChunkType int

const (
	ChunkIgnoredBOM ChunkType = iota
	ChunkUnchangedRunes
	ChunkNormalizedCRLF
	ChunkExpandedTab
	ChunkConvertedISO8859_1
)

// statically ensure that certain interfaces are implemented by Preprocessor
var _ io.Writer = &Preprocessor{}

func (p *Preprocessor) Write(buf []byte) (int, error) {
	for _, b := range buf {
		p.WriteByte(b)
	}
	return len(buf), nil
}

const (
	_CR = '\r'
	_LF = '\n'
)

func (p *Preprocessor) WriteByte(b byte) error {
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
			p.otherChunk(ChunkIgnoredBOM, p.Pending...)
			return nil
		default: // still not sure
			return nil
		}
	}

	if p.state == preproCR {
		p.Pending = nil
		if b == _LF {
			// CRLF detected
			p.otherChunk(ChunkNormalizedCRLF, _LF)
			return nil
		}
		// Flush the pending _CR
		p.state = preproNormal
		p.normalChunk(_CR)
	}

	// Detect invalid UTF-8 and assume it's ISO-8859-1 then. [#document]
	if b >= utf8.RuneSelf {
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
			p.writeAsISO8859_1(buf...)
		} else {
			p.normalChunk(buf...)
		}
		return nil
	}
	// If any bytes are pending, they failed to build a complete UTF-8 rune; flush them
	if len(p.Pending) > 0 {
		buf := p.Pending
		p.Pending = nil
		p.writeAsISO8859_1(buf...)
	}

	// Convert CRLF to LF. [#characters]
	if b == _CR {
		p.state = preproCR
		// Show to user that there's a byte pending, so processing is
		// not fully completed.
		p.Pending = []byte{_CR}
		return nil
	}

	// Expand tabs to spaces. [#lines]
	if b == '\t' {
		spaces := 4 - (p.column % 4)
		bufSpaces := []byte("    ")
		p.otherChunk(ChunkExpandedTab, bufSpaces[:spaces]...)
		return nil
	}

	p.normalChunk(b)
	return nil
}

func (p *Preprocessor) Close() error {
	// TODO(akavel): change to preproClosed state and panic on any later action
	if p.state == preproCR {
		p.Pending = nil
		p.normalChunk(_CR)
		return nil
	}
	if len(p.Pending) > 0 {
		buf := p.Pending
		p.Pending = nil
		p.writeAsISO8859_1(buf...)
	}
	return nil
}

func (p *Preprocessor) normalChunk(b ...byte) {
	typ := ChunkUnchangedRunes
	if p.state == preproISO8859_1 {
		typ = ChunkConvertedISO8859_1
	}

	p.calcColumn(b)
	n := len(p.Chunks)
	if n == 0 || p.Chunks[n-1].Type != typ {
		p.Chunks = append(p.Chunks, Chunk{Type: typ})
		n++
	}
	p.Chunks[n-1].Bytes = append(p.Chunks[n-1].Bytes, b...)
}

func (p *Preprocessor) otherChunk(typ ChunkType, b ...byte) {
	p.calcColumn(b)
	p.Chunks = append(p.Chunks, Chunk{
		Bytes: b,
		Type:  typ,
	})
	p.Pending = nil
	p.state = preproNormal
}

func (p *Preprocessor) calcColumn(added []byte) {
	i := bytes.LastIndex(added, []byte{_LF})
	if i >= 0 {
		p.column = 0
		added = added[i:]
	}
	p.column += utf8.RuneCount(added)
}

func (p *Preprocessor) writeAsISO8859_1(bytes ...byte) {
	p.state = preproISO8859_1

	// Unicode codepoints U+0000 to U+00ff correspond to ISO 8859-1, and
	// result in 1-2 bytes when encoded as UTF-8
	buf := make([]byte, 0, 2)
	for _, b := range bytes {
		r := rune(b)
		buf = buf[:utf8.RuneLen(r)] // 1 or 2
		utf8.EncodeRune(buf, r)
		p.Write(buf)
	}

	p.state = preproNormal
}
