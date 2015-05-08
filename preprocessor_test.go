package vfmd

import (
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
		expected: []Chunk{{bb('a', 0xEF, 0xBB, 0xBF), ChunkUnchangedBytes}},
		pending:  nil,
	}, {
		input:    bb(0xEF, 0xBB),
		expected: nil,
		pending:  bb(0xEF, 0xBB),
	}, {
		input:    bb(0xEF, 0xB0),
		expected: []Chunk{{bb(0xEF, 0xB0), ChunkUnchangedBytes}},
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
