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
	// Detect checks if the provided startLine and optionally secondLine
	// signify start of the particular Block kind. If unsuccessful, 0 and 0
	// should be returned.  If successful, at least one of the returned
	// numbers should be positive.  Consumed is number of lines that sure
	// belong to the block, and won't be needed in any subsequent calls to
	// Continue. Paused is number of lines that may be still needed in
	// subsequent calls to Continue, and/or aren't yet fully confirmed to
	// belong to the block.
	//
	// Note: it is not allowed for Detect to report 0 consumed lines and
	// then for Continue to reject all the paused lines.
	//
	// Note: secondLine==nil means end of file/stream
	Detect(startLine, secondLine []byte) (consumed, paused int)
	// Continue checks if the specified pausedLines and newLine may belong
	// to the block, as reported started by Detect. If any of the lines is
	// detected to be of a next block, Continue should report:
	// consumed <= len(pausedLines), and no paused. Otherwise, Continue
	// must report: consumed+paused == len(pausedLines)+1.
	//
	// Note: newLine==nil means end of file/stream; however, Continue will
	// never be called with newLine==nil if previous call to
	// Detect/Continue didn't report any paused lines.
	Continue(pausedLines [][]byte, newLine []byte) (consumed, paused int)
}

var (
	reUnorderedList  = regexp.MustCompile(`^( *[\*\-\+] +)[^ ]`)
	reOrderedList    = regexp.MustCompile(`^( *([0-9]+)\. +)[^ ]`)
	reHorizontalRule = regexp.MustCompile(`^ *((\* *\* *\* *[\* ]*)|(\- *\- *\- *[\- ]*)|(_ *_ *_ *[_ ]*))$`)
)

func isBlank(line []byte) bool {
	return len(bytes.Trim(line, " \t")) == 0
}
func hasNonSpaceInPrefix(line []byte, n int) bool {
	for i := 0; i < n && i < len(line); i++ {
		if line[i] != ' ' {
			return true
		}
	}
	return false
}
func hasFourSpacePrefix(line []byte) bool {
	return bytes.HasPrefix(line, []byte("    "))
}

type BlockNeverContinue struct{}

func (BlockNeverContinue) Continue([][]byte, []byte) (consumed, paused int) { return 0, 0 }

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

func (b *NullBlock) Detect(line, secondLine []byte) (consumed, paused int) {
	if isBlank(line) {
		return 1, 0
	}
	return 0, 0
}

type ReferenceResolutionBlock struct {
	BlockNeverContinue
	unprocessedReferenceID        []byte
	refValueSequence              []byte
	unprocessedUrl                []byte
	refDefinitionTrailingSequence []byte
}

func (b *ReferenceResolutionBlock) Detect(line, secondLine []byte) (consumed, paused int) {
	if hasFourSpacePrefix(line) {
		return 0, 0
	}
	// TODO(akavel): move the regexp out of function, for speed (or cache it?)
	re := regexp.MustCompile(`^ *\[(([^\\\[\]\!]|\\.|\![^\[])*((\!\[([^\\\[\]]|\\.)*\](\[([^\\\[\]]|\\.)*\])?)?([^\\\[\]]|\\.)*)*)\] *:(.*)$`)
	m := re.FindSubmatch(line)
	if len(m) == 0 {
		return 0, 0
	}
	b.unprocessedReferenceID = m[1]
	b.refValueSequence = m[9] // TODO(akavel): verify if right one
	re = regexp.MustCompile(`^ *([^ \<\>]+|\<[^\<\>]*\>)( .*)?$`)
	m = re.FindSubmatch(b.refValueSequence)
	if len(m) == 0 {
		return 0, 0
	}
	b.unprocessedUrl = m[1]
	b.refDefinitionTrailingSequence = m[2]

	// Detected ok. Now check if 1 or 2 lines.
	re = regexp.MustCompile(`^ +("(([^"\\]|\\.)*)"|'(([^'\\]|\\.)*)'|\(([^\\\(\)]|\\.)*\)) *$`)
	if bytes.IndexAny(b.refDefinitionTrailingSequence, " ") == -1 &&
		secondLine != nil &&
		re.Match(secondLine) {
		return 2, 0
	} else {
		return 1, 0
	}
}

type SetextHeaderBlock struct{ BlockNeverContinue }

func (b *SetextHeaderBlock) Detect(line, secondLine []byte) (consumed, paused int) {
	if secondLine == nil {
		return 0, 0
	}
	re := regexp.MustCompile(`^(-+|=+) *$`)
	if re.Match(secondLine) {
		return 2, 0
	}
	return 0, 0
}

type CodeBlock struct{}

func (b *CodeBlock) Detect(line, secondLine []byte) (consumed, paused int) {
	if hasFourSpacePrefix(line) {
		return 1, 0
	}
	return 0, 0
}
func (b *CodeBlock) Continue(pausedLines [][]byte, newLine []byte) (consumed, paused int) {
	// FIXME(akavel): handle newLine==nil !!!
	// TODO(akavel): verify it's coded ok, it was converted from a different approach
	switch {
	// previous blank, current is not tab-indented
	case len(pausedLines) > 0 && !hasFourSpacePrefix(newLine):
		return 0, 0
	case isBlank(newLine):
		return len(pausedLines), 1
	case hasFourSpacePrefix(newLine):
		return len(pausedLines) + 1, 0
	// current not blank & not indented. End the block.
	default:
		return len(pausedLines), 0
	}
}

type AtxHeaderBlock struct{ BlockNeverContinue }

func (b *AtxHeaderBlock) Detect(line, secondLine []byte) (consumed, paused int) {
	if bytes.HasPrefix(line, []byte("#")) {
		return 1, 0
	}
	return 0, 0
}

type QuoteBlock struct{}

func (b *QuoteBlock) Detect(line, secondLine []byte) (consumed, paused int) {
	ltrim := bytes.TrimLeft(line, " ")
	if len(ltrim) > 0 && ltrim[0] == '>' {
		return 0, 1
	}
	return 0, 0
}
func (b *QuoteBlock) Continue(pausedLines [][]byte, newLine []byte) (consumed, paused int) {
	// TODO(akavel): verify it's coded ok, it was converted from a different approach
	if newLine == nil {
		return len(pausedLines), 0
	}
	if len(pausedLines) != 1 {
		panic("len(pausedLines)!=1")
	}
	if isBlank(pausedLines[0]) {
		if isBlank(newLine) ||
			hasFourSpacePrefix(newLine) ||
			bytes.TrimLeft(newLine, " ")[0] != '>' {
			return len(pausedLines), 0
		}
	} else if !hasFourSpacePrefix(newLine) &&
		reHorizontalRule.Match(newLine) {
		return len(pausedLines), 0
	}
	return len(pausedLines), 1
}

type HorizontalRuleBlock struct{ BlockNeverContinue }

func (b *HorizontalRuleBlock) Detect(line, secondLine []byte) (consumed, paused int) {
	if reHorizontalRule.Match(line) {
		return 1, 0
	}
	return 0, 0
}

type UnorderedListBlock struct {
	Starter []byte
}

func (b *UnorderedListBlock) Detect(line, secondLine []byte) (consumed, paused int) {
	m := reUnorderedList.FindSubmatch(line)
	if m == nil {
		return 0, 0
	}
	b.Starter = m[1]
	return 0, 1
}
func (b *UnorderedListBlock) Continue(pausedLines [][]byte, newLine []byte) (consumed, paused int) {
	if newLine == nil {
		return len(pausedLines), 0
	}
	if len(pausedLines) != 1 {
		panic("len(pausedLines)!=1")
	}

	if isBlank(pausedLines[0]) {
		if isBlank(newLine) {
			return len(pausedLines), 0
		}
		if !bytes.HasPrefix(newLine, b.Starter) &&
			// FIXME(akavel): spec refers to runes ("characters"), not bytes; fix this everywhere
			hasNonSpaceInPrefix(newLine, len(b.Starter)) {
			return len(pausedLines), 0
		}
	} else {
		if !bytes.HasPrefix(newLine, b.Starter) &&
			hasNonSpaceInPrefix(newLine, len(b.Starter)) &&
			!hasFourSpacePrefix(newLine) &&
			(reUnorderedList.Match(newLine) ||
				reOrderedList.Match(newLine) ||
				reHorizontalRule.Match(newLine)) {
			return len(pausedLines), 0
		}
	}
	return len(pausedLines), 1
}

type OrderedListBlock struct {
	Starter []byte
}

func (b *OrderedListBlock) Detect(line, secondLine []byte) (consumed, paused int) {
	m := reOrderedList.FindSubmatch(line)
	if m == nil {
		return 0, 0
	}
	b.Starter = m[1]
	return 0, 1
}
func (b *OrderedListBlock) Continue(pausedLines [][]byte, newLine []byte) (consumed, paused int) {
	if newLine == nil {
		return len(pausedLines), 0
	}
	if len(pausedLines) != 1 {
		panic("len(pausedLines)!=1")
	}

	if isBlank(pausedLines[0]) {
		if isBlank(newLine) {
			return len(pausedLines), 0
		}
		if !reOrderedList.Match(newLine) &&
			hasNonSpaceInPrefix(newLine, len(b.Starter)) {
			return len(pausedLines), 0
		}
	} else {
		if !reOrderedList.Match(newLine) &&
			hasNonSpaceInPrefix(newLine, len(b.Starter)) &&
			!hasFourSpacePrefix(newLine) &&
			(reUnorderedList.Match(newLine) ||
				reHorizontalRule.Match(newLine)) {
			return len(pausedLines), 0
		}
	}
	return len(pausedLines), 1
}

type ParagraphBlock struct {
	// NOTE: below fields must be set appropriately when creating a ParagraphBlock
	IsInBlockquote bool
	IsInList       bool
}

func (b *ParagraphBlock) Detect(line, secondLine []byte) (consumed, paused int) {
	return 0, 1
}
func (b *ParagraphBlock) Continue(pausedLines [][]byte, newLine []byte) (consumed, paused int) {
	if newLine == nil {
		return len(pausedLines), 0
	}
	if len(pausedLines) != 1 {
		panic("len(pausedLines)!=1")
	}
	// TODO(akavel): support HTML parser & related interactions [#paragraph-line-sequence]
	if isBlank(pausedLines[0]) {
		return len(pausedLines), 0
	}
	if !hasFourSpacePrefix(newLine) {
		if reHorizontalRule.Match(newLine) ||
			(b.IsInBlockquote && bytes.HasPrefix(bytes.TrimLeft(newLine, " "), []byte(">"))) ||
			(b.IsInList && reOrderedList.Match(newLine)) ||
			(b.IsInList && reUnorderedList.Match(newLine)) {
			return len(pausedLines), 0
		}
	}
	return len(pausedLines), 1
}
