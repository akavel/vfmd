package vfmd

import (
	"bytes"
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
		expected: []Chunk{{ChunkIgnoredBOM, nil}},
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
			{ChunkConvertedISO8859_1, bs("\u00ef\u00ff")},
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
	wrapISO := func(bytes []byte) []Chunk {
		return []Chunk{
			{ChunkConvertedISO8859_1, bytes},
		}
	}
	cases := []struct {
		input  []byte
		chunks []Chunk
	}{
		// Invalid UTF-8 should be converted as if it was ISO8859-1
		{bb(0x80), wrapISO(bs("\u0080"))},
		{bb(0xFF), wrapISO(bs("\u00FF"))},
		{bb(0xAA, 0xBB), wrapISO(bs("\u00AA\u00BB"))},
		// Correctly encoded UTF-8 should pass through unchanged
		{bs("żął"), []Chunk{{ChunkUnchangedRunes, bs("żął")}}},
		{bs("halo"), []Chunk{{ChunkUnchangedRunes, bs("halo")}}},
		// Mixed
		{append([]byte("ż"), 0xFF, 'a'), []Chunk{
			{ChunkUnchangedRunes, bs("ż")},
			{ChunkConvertedISO8859_1, bs("\u00FF")},
			{ChunkUnchangedRunes, bb('a')},
		}},
	}
	for _, c := range cases {
		p := Preprocessor{}
		p.Write(c.input)
		p.Close()

		if !reflect.DeepEqual(c.chunks, p.Chunks) {
			test.Errorf("case '% 2x' expected % 2x got % 2x",
				c.input, c.chunks, p.Chunks)
		}

		if len(p.Pending) > 0 {
			test.Errorf("case '% 2x' expected pending nil, got '% 2x'",
				c.input, p.Pending)
		}
	}
}

func TestTabExpansion(test *testing.T) {
	cases := []struct {
		input  []byte
		chunks []Chunk
	}{
		{bs("\t"), []Chunk{
			{ChunkExpandedTab, bs("    ")},
		}},
		{bs("\t\t"), []Chunk{
			{ChunkExpandedTab, bs("    ")},
			{ChunkExpandedTab, bs("    ")},
		}},

		{bs("a\t"), []Chunk{
			{ChunkUnchangedRunes, bs("a")},
			{ChunkExpandedTab, bs("   ")},
		}},
		{bs("ab\t"), []Chunk{
			{ChunkUnchangedRunes, bs("ab")},
			{ChunkExpandedTab, bs("  ")},
		}},
		{bs("abc\t"), []Chunk{
			{ChunkUnchangedRunes, bs("abc")},
			{ChunkExpandedTab, bs(" ")},
		}},
		{bs("abcd\t"), []Chunk{
			{ChunkUnchangedRunes, bs("abcd")},
			{ChunkExpandedTab, bs("    ")},
		}},

		{bs("ż\t"), []Chunk{
			{ChunkUnchangedRunes, bs("ż")},
			{ChunkExpandedTab, bs("   ")},
		}},
		{bb(0x80, '\t'), []Chunk{
			{ChunkConvertedISO8859_1, bs("\u0080")},
			{ChunkExpandedTab, bs("   ")},
		}},
	}
	for _, c := range cases {
		p := Preprocessor{}
		p.Write(c.input)

		if !reflect.DeepEqual(p.Chunks, c.chunks) {
			test.Errorf("case %q expected % 2x got % 2x",
				c.input, c.chunks, p.Chunks)
		}

		if len(p.Pending) > 0 {
			test.Errorf("case '% 2x' expected pending nil, got '% 2x'",
				c.input, p.Pending)
		}
	}
}

func TestCRLFPending(test *testing.T) {
	cases := []struct {
		input   []byte
		output  []Chunk
		pending []byte
	}{{
		input:   bs("\r"),
		output:  nil,
		pending: bs("\r"),
	}, {
		input:   bs("\r\n"),
		output:  []Chunk{{ChunkIgnoredCR, nil}, {ChunkUnchangedLF, bs("\n")}},
		pending: nil,
	}}

	for _, c := range cases {
		p := Preprocessor{}
		p.Write(c.input)

		if !reflect.DeepEqual(p.Chunks, c.output) {
			test.Errorf("case '% 2x' expected '% 2x' got '% 2x'",
				c.input, c.output, p.Chunks)
		}
		if !bytes.Equal(p.Pending, c.pending) {
			test.Errorf("case '% 2x' expected pending '% 2x' got '% 2x'",
				c.input, c.pending, p.Pending)
		}

	}

}
