package vfmd

import (
	"reflect"
	"testing"
)

func bb(bytes ...byte) []byte { return bytes }
func bs(s string) []byte      { return []byte(s) }

func TestBOM(test *testing.T) {
	BOM := bb(0xEF, 0xBB, 0xBF)
	cases := []struct {
		input    []byte
		expected []Chunk
		pending  []byte
	}{{
		input:    BOM,
		expected: []Chunk{{ChunkIgnoredBOM, BOM}},
		pending:  nil,
	}, {
		input:    bb('a', 0xEF, 0xBB, 0xBF),
		expected: []Chunk{{ChunkUnchangedRunes, bb('a', 0xEF, 0xBB, 0xBF)}},
		pending:  nil,
	}, {
		input:    bb(0xEF, 0xBB),
		expected: nil,
		pending:  bb(0xEF, 0xBB),
	}, {
		input: bb(0xEF, 0xFF),
		expected: []Chunk{
			{ChunkStartISO8859_1, nil},
			{ChunkUnchangedRunes, bs("\u00ef\u00ff")},
			{ChunkEndISO8859_1, nil},
		},
		pending: nil,
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

func TestISO8859_1(test *testing.T) {
	cases := []struct{ input, output []byte }{
		{bb(0x80), bs("\u0080")},
		{bb(0xFF), bs("\u00FF")},
		{bb(0xAA, 0xBB), bs("\u00AA\u00BB")},
	}
	for _, c := range cases {
		p := Preprocessor{}
		p.Write(c.input)
		p.Close()

		chunks := []Chunk{
			{ChunkStartISO8859_1, nil},
			{ChunkUnchangedRunes, c.output},
			{ChunkEndISO8859_1, nil},
		}
		if !reflect.DeepEqual(chunks, p.Chunks) {
			test.Errorf("case '% 2x' expected % 2x got % 2x",
				c.input, chunks, p.Chunks)
		}

		if len(p.Pending) > 0 {
			test.Errorf("case '% 2x' expected pending nil, got '% 2x'",
				c.input, p.Pending)
		}
	}
}
