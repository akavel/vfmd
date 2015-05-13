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

type Line []byte

func (line Line) isBlank() bool {
	return len(bytes.Trim(line, " \t")) == 0
}
func (line Line) hasNonSpaceInPrefix(n int) bool {
	for i := 0; i < n && i < len(line); i++ {
		if line[i] != ' ' {
			return true
		}
	}
	return false
}
func (line Line) hasFourSpacePrefix() bool {
	return bytes.HasPrefix(line, []byte("    "))
}

type Block interface {
	// Detect checks if the provided start line and optionally second line
	// signify start of the particular Block kind. If unsuccessful, 0 and 0
	// should be returned.  If successful, at least one of the returned
	// numbers should be positive.  Consume is number of lines that sure
	// belong to the block, and won't be needed in any subsequent calls to
	// Continue. Pause is number of lines that may be still needed in
	// subsequent calls to Continue, and/or aren't yet fully confirmed to
	// belong to the block.
	//
	// Note: it is not allowed for Detect to report 0 lines to consume and
	// then for Continue to reject all the paused lines.
	//
	// Note: secondLine==nil means end of file/stream
	Detect(start, second Line) (consume, pause int)
	// Continue checks if the specified paused lines and next line may belong
	// to the block, as reported started by Detect. If any of the lines is
	// detected to be of a next block, Continue should report:
	// consume <= len(paused), and nothing to pause. Otherwise, Continue
	// must report: consume+pause == len(paused)+1.
	//
	// Note: next==nil means end of file/stream; however, Continue will
	// never be called with next==nil if previous call to
	// Detect/Continue didn't report any lines to pause.
	Continue(paused []Line, next Line) (consume, pause int)
}

var (
	reUnorderedList  = regexp.MustCompile(`^( *[\*\-\+] +)[^ ]`)
	reOrderedList    = regexp.MustCompile(`^( *([0-9]+)\. +)[^ ]`)
	reHorizontalRule = regexp.MustCompile(`^ *((\* *\* *\* *[\* ]*)|(\- *\- *\- *[\- ]*)|(_ *_ *_ *[_ ]*))$`)
)

type BlockNeverContinue struct{}

func (BlockNeverContinue) Continue([]Line, Line) (consume, pause int) { return 0, 0 }

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

func (b *NullBlock) Detect(line, secondLine Line) (consumed, paused int) {
	if line.isBlank() {
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

func (b *ReferenceResolutionBlock) Detect(line, secondLine Line) (consumed, paused int) {
	if line.hasFourSpacePrefix() {
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

func (b *SetextHeaderBlock) Detect(line, secondLine Line) (consumed, paused int) {
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

func (b *CodeBlock) Detect(line, secondLine Line) (consumed, paused int) {
	if line.hasFourSpacePrefix() {
		return 1, 0
	}
	return 0, 0
}
func (b *CodeBlock) Continue(pausedLines []Line, newLine Line) (consumed, paused int) {
	// FIXME(akavel): handle newLine==nil !!!
	// TODO(akavel): verify it's coded ok, it was converted from a different approach
	switch {
	// previous blank, current is not tab-indented
	case len(pausedLines) > 0 && !newLine.hasFourSpacePrefix():
		return 0, 0
	case newLine.isBlank():
		return len(pausedLines), 1
	case newLine.hasFourSpacePrefix():
		return len(pausedLines) + 1, 0
	// current not blank & not indented. End the block.
	default:
		return len(pausedLines), 0
	}
}

type AtxHeaderBlock struct{ BlockNeverContinue }

func (b *AtxHeaderBlock) Detect(line, secondLine Line) (consumed, paused int) {
	if bytes.HasPrefix(line, []byte("#")) {
		return 1, 0
	}
	return 0, 0
}

type QuoteBlock struct{}

func (b *QuoteBlock) Detect(line, secondLine Line) (consumed, paused int) {
	ltrim := bytes.TrimLeft(line, " ")
	if len(ltrim) > 0 && ltrim[0] == '>' {
		return 0, 1
	}
	return 0, 0
}
func (b *QuoteBlock) Continue(pausedLines []Line, newLine Line) (consumed, paused int) {
	// TODO(akavel): verify it's coded ok, it was converted from a different approach
	if newLine == nil {
		return len(pausedLines), 0
	}
	if len(pausedLines) != 1 {
		panic("len(pausedLines)!=1")
	}
	if pausedLines[0].isBlank() {
		if newLine.isBlank() ||
			newLine.hasFourSpacePrefix() ||
			bytes.TrimLeft(newLine, " ")[0] != '>' {
			return len(pausedLines), 0
		}
	} else if !newLine.hasFourSpacePrefix() &&
		reHorizontalRule.Match(newLine) {
		return len(pausedLines), 0
	}
	return len(pausedLines), 1
}

type HorizontalRuleBlock struct{ BlockNeverContinue }

func (b *HorizontalRuleBlock) Detect(line, secondLine Line) (consumed, paused int) {
	if reHorizontalRule.Match(line) {
		return 1, 0
	}
	return 0, 0
}

type UnorderedListBlock struct {
	Starter []byte
}

func (b *UnorderedListBlock) Detect(line, secondLine Line) (consumed, paused int) {
	m := reUnorderedList.FindSubmatch(line)
	if m == nil {
		return 0, 0
	}
	b.Starter = m[1]
	return 0, 1
}
func (b *UnorderedListBlock) Continue(pausedLines []Line, newLine Line) (consumed, paused int) {
	if newLine == nil {
		return len(pausedLines), 0
	}
	if len(pausedLines) != 1 {
		panic("len(pausedLines)!=1")
	}

	if pausedLines[0].isBlank() {
		if newLine.isBlank() {
			return len(pausedLines), 0
		}
		if !bytes.HasPrefix(newLine, b.Starter) &&
			// FIXME(akavel): spec refers to runes ("characters"), not bytes; fix this everywhere
			newLine.hasNonSpaceInPrefix(len(b.Starter)) {
			return len(pausedLines), 0
		}
	} else {
		if !bytes.HasPrefix(newLine, b.Starter) &&
			newLine.hasNonSpaceInPrefix(len(b.Starter)) &&
			!newLine.hasFourSpacePrefix() &&
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

func (b *OrderedListBlock) Detect(line, secondLine Line) (consumed, paused int) {
	m := reOrderedList.FindSubmatch(line)
	if m == nil {
		return 0, 0
	}
	b.Starter = m[1]
	return 0, 1
}
func (b *OrderedListBlock) Continue(pausedLines []Line, newLine Line) (consumed, paused int) {
	if newLine == nil {
		return len(pausedLines), 0
	}
	if len(pausedLines) != 1 {
		panic("len(pausedLines)!=1")
	}

	if pausedLines[0].isBlank() {
		if newLine.isBlank() {
			return len(pausedLines), 0
		}
		if !reOrderedList.Match(newLine) &&
			newLine.hasNonSpaceInPrefix(len(b.Starter)) {
			return len(pausedLines), 0
		}
	} else {
		if !reOrderedList.Match(newLine) &&
			newLine.hasNonSpaceInPrefix(len(b.Starter)) &&
			!newLine.hasFourSpacePrefix() &&
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

func (b *ParagraphBlock) Detect(line, secondLine Line) (consumed, paused int) {
	return 0, 1
}
func (b *ParagraphBlock) Continue(pausedLines []Line, newLine Line) (consumed, paused int) {
	if newLine == nil {
		return len(pausedLines), 0
	}
	if len(pausedLines) != 1 {
		panic("len(pausedLines)!=1")
	}
	// TODO(akavel): support HTML parser & related interactions [#paragraph-line-sequence]
	if pausedLines[0].isBlank() {
		return len(pausedLines), 0
	}
	if !newLine.hasFourSpacePrefix() {
		if reHorizontalRule.Match(newLine) ||
			(b.IsInBlockquote && bytes.HasPrefix(bytes.TrimLeft(newLine, " "), []byte(">"))) ||
			(b.IsInList && reOrderedList.Match(newLine)) ||
			(b.IsInList && reUnorderedList.Match(newLine)) {
			return len(pausedLines), 0
		}
	}
	return len(pausedLines), 1
}
