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

var (
	reUnorderedList  = regexp.MustCompile(`^( *[\*\-\+] +)[^ ]`)
	reOrderedList    = regexp.MustCompile(`^( *([0-9]+)\. +)[^ ]`)
	reHorizontalRule = regexp.MustCompile(`^ *((\* *\* *\* *[\* ]*)|(\- *\- *\- *[\- ]*)|(_ *_ *_ *[_ ]*))$`)
)

func isBlank(line []byte) bool {
	return len(bytes.Trim(line, " \t")) == 0
}

type BlockBase struct{ L [][]byte }

func (b BlockBase) Lines() [][]byte { return b.L }
func (b BlockBase) LastLine() []byte {
	if len(b.L) == 0 {
		return nil
	}
	return b.L[len(b.L)-1]
}

type BlockNeverContinue struct{ BlockBase }

func (BlockNeverContinue) Continue(line []byte) (reflux [][]byte) {
	return [][]byte{line}
}

// static assertion of Block interface implementation by the listed types
var _ []Block = []Block{
	&NullBlock{},
	&ReferenceResolutionBlock{},
	&SetextHeaderBlock{},
	&CodeBlock{},
	&AtxHeaderBlock{},
	&QuoteBlock{},
	&HorizontalRuleBlock{},
	&UnorderedListBlock{},
	&OrderedListBlock{},
	&ParagraphBlock{},
}

type NullBlock struct{ BlockNeverContinue }

func (b *NullBlock) Detect(line, secondLine []byte) (consumed int) {
	if isBlank(line) {
		b.L = [][]byte{line}
		return 1
	}
	return 0
}

type ReferenceResolutionBlock struct {
	BlockNeverContinue
	unprocessedReferenceID        []byte
	refValueSequence              []byte
	unprocessedUrl                []byte
	refDefinitionTrailingSequence []byte
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

type SetextHeaderBlock struct {
	BlockNeverContinue
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

type CodeBlock struct {
	BlockBase
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
		prev := b.LastLine()
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

type AtxHeaderBlock struct{ BlockNeverContinue }

func (b *AtxHeaderBlock) Detect(line, secondLine []byte) (consumed int) {
	if bytes.HasPrefix(line, []byte("#")) {
		b.L = [][]byte{line}
		return 1
	}
	return 0
}

type QuoteBlock struct{ BlockBase }

func (b *QuoteBlock) Detect(line, secondLine []byte) (consumed int) {
	ltrim := bytes.TrimLeft(line, " ")
	if len(ltrim) > 0 && ltrim[0] == '>' {
		b.L = [][]byte{line}
		return 1
	}
	return 0
}
func (b *QuoteBlock) Continue(line []byte) (reflux [][]byte) {
	if line == nil {
		return [][]byte{line}
	}
	if isBlank(b.LastLine()) {
		if isBlank(line) ||
			bytes.HasPrefix(line, []byte("    ")) ||
			bytes.TrimLeft(line, " ")[0] != '>' {
			return [][]byte{line}
		}
	} else if !bytes.HasPrefix(line, []byte("    ")) &&
		reHorizontalRule.Match(line) {
		return [][]byte{line}
	}
	b.L = append(b.L, line)
	return nil
}

type HorizontalRuleBlock struct{ BlockNeverContinue }

func (b *HorizontalRuleBlock) Detect(line, secondLine []byte) (consumed int) {
	if reHorizontalRule.Match(line) {
		b.L = [][]byte{line}
		return 1
	}
	return 0
}

type UnorderedListBlock struct {
	BlockBase
	Starter []byte
}

func (b *UnorderedListBlock) Detect(line, secondLine []byte) (consumed int) {
	m := reUnorderedList.FindSubmatch(line)
	if m == nil {
		return 0
	}
	b.L = [][]byte{line}
	b.Starter = m[1]
	return 1
}
func (b *UnorderedListBlock) Continue(line []byte) (reflux [][]byte) {
	if line == nil {
		return [][]byte{line}
	}

	prefix := len(b.Starter)
	if len(line) < prefix {
		prefix = len(line)
	}
	if isBlank(b.LastLine()) {
		if isBlank(line) {
			return [][]byte{line}
		}
		if !bytes.HasPrefix(line, b.Starter) &&
			// FIXME(akavel): spec refers to runes ("characters"), not bytes; fix this everywhere
			// has non-space chars in first prefix characters
			len(bytes.Trim(line[:prefix], " ")) > 0 {
			return [][]byte{line}
		}
	} else {
		if !bytes.HasPrefix(line, b.Starter) &&
			len(bytes.Trim(line[:prefix], " ")) > 0 &&
			!bytes.HasPrefix(line, []byte("    ")) {
			if reUnorderedList.Match(line) ||
				reOrderedList.Match(line) ||
				reHorizontalRule.Match(line) {
				return [][]byte{line}
			}
		}
	}
	b.L = append(b.L, line)
	return nil
}

type OrderedListBlock struct {
	BlockBase
	Starter []byte
}

func (b *OrderedListBlock) Detect(line, secondLine []byte) (consumed int) {
	m := reOrderedList.FindSubmatch(line)
	if m == nil {
		return 0
	}
	b.L = [][]byte{line}
	b.Starter = m[1]
	return 1
}
func (b *OrderedListBlock) Continue(line []byte) (reflux [][]byte) {
	if line == nil {
		return [][]byte{line}
	}

	prefix := len(b.Starter)
	if len(line) < prefix {
		prefix = len(line)
	}
	if isBlank(b.LastLine()) {
		if isBlank(line) {
			return [][]byte{line}
		}
		if !reOrderedList.Match(line) &&
			// TODO(akavel): extract below pattern to hasNonSpaceInFirstChars(line, n)
			len(bytes.Trim(line[:prefix], " ")) > 0 {
			return [][]byte{line}
		}
	} else {
		if !reOrderedList.Match(line) &&
			len(bytes.Trim(line[:prefix], " ")) > 0 &&
			!bytes.HasPrefix(line, []byte("    ")) {
			if reUnorderedList.Match(line) ||
				reHorizontalRule.Match(line) {
				return [][]byte{line}
			}
		}
	}
	b.L = append(b.L, line)
	return nil
}

type ParagraphBlock struct {
	BlockBase
	// NOTE: below fields must be set appropriately when creating a ParagraphBlock
	IsInBlockquote bool
	IsInList       bool
}

func (b *ParagraphBlock) Detect(line, secondLine []byte) (consumed int) {
	b.L = append(b.L, line)
	return 1
}
func (b *ParagraphBlock) Continue(line []byte) (reflux [][]byte) {
	if line == nil {
		return [][]byte{line}
	}
	// TODO(akavel): support HTML parser & related interactions [#paragraph-line-sequence]
	if isBlank(b.LastLine()) {
		return [][]byte{line}
	}
	if !bytes.HasPrefix(line, []byte("    ")) {
		if reHorizontalRule.Match(line) ||
			(b.IsInBlockquote && bytes.HasPrefix(bytes.TrimLeft(line, " "), []byte(">"))) ||
			(b.IsInList && reOrderedList.Match(line)) ||
			(b.IsInList && reUnorderedList.Match(line)) {
			return [][]byte{line}
		}
	}
	b.L = append(b.L, line)
	return nil
}
