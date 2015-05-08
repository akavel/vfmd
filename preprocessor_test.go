package vfmd

import (
	"bytes"
	"reflect"
	"testing"
)

func bb(bytes ...byte) []byte { return bytes }

func TestBOM(test *testing.T) {
	BOM := bb(0xEF, 0xBB, 0xBF)
	cases := []struct {
		input    []byte
		expected []Chunk
		pending  []byte
	}{{
		input:    BOM,
		expected: []Chunk{{BOM, ChunkIgnoredBOM}},
		pending:  nil,
	}, {
		input:    bb('a', 0xEF, 0xBB, 0xBF),
		expected: []Chunk{{bb('a', 0xEF, 0xBB, 0xBF), ChunkUnchangedRunes}},
		pending:  nil,
	}, {
		input:    bb(0xEF, 0xBB),
		expected: nil,
		pending:  bb(0xEF, 0xBB),
	}, {
		input:    bb(0xEF, 0xFF),
		expected: []Chunk{{iso2utf(0xEF, 0xFF), ChunkUnchangedRunes}},
		pending:  nil,
	}}

	for _, c := range cases {
		p := Preprocessor{}
		p.Write(c.input)
		if !reflect.DeepEqual(p.Chunks, c.expected) {
			test.Errorf("case '% 2x' expected % 2x got % 2x",
				c.input, c.expected, p.Chunks)
		}
		if !reflect.DeepEqual(p.Pending, c.pending) {
			test.Errorf("case '% 2x' expected pending '% 2x' got '% 2x'",
				c.input, c.pending, p.Pending)
		}
	}
}

func TestIso2utf(test *testing.T) {
	cases := []struct{ input, output []byte }{
		{bb(0x80), []byte("\u0080")},
		{bb(0xFF), []byte("\u00FF")},
		{bb(0xAA, 0xBB), []byte("\u00AA\u00BB")},
		{bb(0x01), []byte("\u0001")},
	}
	for _, c := range cases {
		output := iso2utf(c.input...)
		if !bytes.Equal(c.output, output) {
			test.Errorf("case '%c' expected '% 2x' got '% 2x'",
				c.input, c.output, output)
		}
	}
}
