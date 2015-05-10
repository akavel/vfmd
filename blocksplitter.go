package vfmd

import (
	"bytes"
	"regexp"
)

// TODO(akavel): add tests for blocks

type BlockSplitter struct {
}

func (b *BlockSplitter) WriteLine(line []byte) error {
	// TODO(akavel): NIY
	return nil
}

func (b *BlockSplitter) Close() error {
	// TODO(akavel): NIY
	return nil
}

type Block interface {
	// Detect checks if startLine and optionally secondLine signify start
	// of the particular Block kind. If negative, 0 should be returned.
	// If positive, number of lines needed to perform the detection
	// should be returned (1 if startLine was enough for detection, 2 if
	// secondLine was used too and belongs to the block). The further
	// lines after the returned number should be passed to Contine().
	//
	// Note: secondLine==nil means end of file/stream
	Detect(startLine, secondLine []byte) (consumedLines int)
	// Continue checks if the specified line may belong to the block, as
	// reported started by Detect. If the line is detected to be of a
	// next block, this line should be returned, preceded by any earlier
	// consumed lines that are now confirmed to be of the next block too.
	//
	// Note: line==nil means end of file/stream
	Continue(line []byte) (refluxLines [][]byte)
	Lines() [][]byte
}

const (
	reUnorderedList  = `^( *[\*\-\+] +)[^ ]`
	reOrderedList    = `^( *([0-9]+)\. +)[^ ]`
	reHorizontalRule = `^ *((\* *\* *\* *[\* ]*)|(\- *\- *\- *[\- ]*)|(_ *_ *_ *[_ ]*))$`
)

func isBlank(line []byte) bool {
	return len(bytes.Trim(line, " \t")) == 0
}

type NullBlock struct{ L [][]byte }

func (b *NullBlock) Detect(line, secondLine []byte) (consumed int) {
	if isBlank(line) {
		b.L = [][]byte{line}
		return 1
	}
	return 0
}
func (b *NullBlock) Continue(line []byte) (reflux [][]byte) {
	return [][]byte{line}
}

type ReferenceResolutionBlock struct {
	unprocessedReferenceID        []byte
	refValueSequence              []byte
	unprocessedUrl                []byte
	refDefinitionTrailingSequence []byte
	L                             [][]byte
}

func (b *ReferenceResolutionBlock) Detect(line, secondLine []byte) (consumed int) {
	if bytes.HasPrefix(line, []byte("    ")) {
		return 0
	}
	// TODO(akavel): move the regexp out of function, for speed (or cache it?)
	re := regexp.MustCompile(`^ *\[(([^\\\[\]\!]|\\.|\![^\[])*((\!\[([^\\\[\]]|\\.)*\](\[([^\\\[\]]|\\.)*\])?)?([^\\\[\]]|\\.)*)*)\] *:(.*)$`)
	m := re.FindSubmatch(line)
	if len(m) == 0 {
		return 0
	}
	b.unprocessedReferenceID = m[1]
	b.refValueSequence = m[9] // TODO(akavel): verify if right one
	re = regexp.MustCompile(`^ *([^ \<\>]+|\<[^\<\>]*\>)( .*)?$`)
	m = re.FindSubmatch(b.refValueSequence)
	if len(m) == 0 {
		return 0
	}
	b.unprocessedUrl = m[1]
	b.refDefinitionTrailingSequence = m[2]

	// Detected ok. Now check if 1 or 2 lines.
	b.L = [][]byte{line}
	re = regexp.MustCompile(`^ +("(([^"\\]|\\.)*)"|'(([^'\\]|\\.)*)'|\(([^\\\(\)]|\\.)*\)) *$`)
	if bytes.IndexAny(b.refDefinitionTrailingSequence, " ") == -1 &&
		secondLine != nil &&
		re.Match(secondLine) {
		b.L = [][]byte{line, secondLine}
	} else {
		b.L = [][]byte{line}
	}
	return len(b.L)
}
func (b *ReferenceResolutionBlock) Continue(line []byte) (reflux [][]byte) {
	return [][]byte{line}
}

type SetextHeaderBlock struct {
	L [][]byte
}

func (b *SetextHeaderBlock) Detect(line, secondLine []byte) (consumed int) {
	if secondLine == nil {
		return 0
	}
	re := regexp.MustCompile(`^(-+|=+) *$`)
	if re.Match(secondLine) {
		b.L = [][]byte{line, secondLine}
		return 2
	}
	return 0
}
func (b *SetextHeaderBlock) Continue(line []byte) (reflux [][]byte) {
	return [][]byte{line}
}

type CodeBlock struct {
	L        [][]byte
	maybeEnd bool
}

func (b *CodeBlock) Detect(line, secondLine []byte) (consumed int) {
	if bytes.HasPrefix(line, []byte("    ")) {
		b.L = [][]byte{line}
		return 1
	}
	return 0
}
func (b *CodeBlock) Continue(line []byte) (reflux [][]byte) {
	if b.maybeEnd && !bytes.HasPrefix(line, []byte("    ")) {
		prev := b.L[len(b.L)-1]
		b.L = b.L[:len(b.L)-1]
		return [][]byte{prev, line}
	}
	blank := isBlank(line)
	if !blank && !bytes.HasPrefix(line, []byte("    ")) {
		return [][]byte{line}
	}
	b.maybeEnd = blank
	b.L = append(b.L, line)
	return nil
}

type AtxHeaderBlock struct{}
type QuoteBlock struct{}
type HorizontalRuleBlock struct{}
type UnorderedListBlock struct{}
type OrderedListBlock struct{}
type ParagraphBlock struct{}
