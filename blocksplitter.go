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

type BlockDetector interface {
	// Detect checks if the provided start line and optionally second line
	// signify start of the particular block kind.  If unsuccessful, 0 and
	// 0 should be returned.  If successful, at least one of the returned
	// numbers should be positive.  Consume is number of lines that sure
	// belong to the block, and won't be needed in any subsequent calls to
	// Continue.  Pause is number of lines that may be still needed in
	// subsequent calls to Continue, and/or aren't yet fully confirmed to
	// belong to the block.
	//
	// Note: it is not allowed for Detect to report 0 lines to consume and
	// then for Continue to reject all the paused lines.
	//
	// Note: second==nil means end of file/stream
	Detect(start, second Line) (consume, pause int)
	// Continue checks if the specified paused lines and next line may
	// belong to the block, as reported started by Detect.  If any of the
	// lines is detected to be of a next block, Continue should report:
	// consume <= len(paused), and nothing to pause.  Otherwise, Continue
	// must report: consume+pause == len(paused)+1.
	//
	// Number of paused lines passed to Continue will always be equal to
	// value reported from previous Detect or Continue.
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

// static assertion of BlockDetector interface implementation by the listed types
var _ []BlockDetector = []BlockDetector{
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

func (b *NullBlock) Detect(start, second Line) (consume, pause int) {
	if start.isBlank() {
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

func (b *ReferenceResolutionBlock) Detect(start, second Line) (consume, pause int) {
	if start.hasFourSpacePrefix() {
		return 0, 0
	}
	// TODO(akavel): move the regexp out of function, for speed (or cache it?)
	re := regexp.MustCompile(`^ *\[(([^\\\[\]\!]|\\.|\![^\[])*((\!\[([^\\\[\]]|\\.)*\](\[([^\\\[\]]|\\.)*\])?)?([^\\\[\]]|\\.)*)*)\] *:(.*)$`)
	m := re.FindSubmatch(start)
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
		second != nil &&
		re.Match(second) {
		return 2, 0
	} else {
		return 1, 0
	}
}

type SetextHeaderBlock struct{ BlockNeverContinue }

func (b *SetextHeaderBlock) Detect(start, second Line) (consume, pause int) {
	if second == nil {
		return 0, 0
	}
	re := regexp.MustCompile(`^(-+|=+) *$`)
	if re.Match(second) {
		return 2, 0
	}
	return 0, 0
}

type CodeBlock struct{}

func (b *CodeBlock) Detect(start, second Line) (consume, pause int) {
	if start.hasFourSpacePrefix() {
		return 1, 0
	}
	return 0, 0
}
func (b *CodeBlock) Continue(paused []Line, next Line) (consume, pause int) {
	// FIXME(akavel): handle next==nil !!!
	if next == nil {
		return 0, 0
		// note: len(paused)==1 if prev was blank, so we can ditch it anyway
	}
	// TODO(akavel): verify it's coded ok, it was converted from a different approach
	switch {
	// previous was blank, next is not tab-indented. Reject both.
	case len(paused) == 1 && !next.hasFourSpacePrefix():
		return 0, 0
	case next.isBlank():
		return len(paused), 1 // note: only case where we pause a line
	case next.hasFourSpacePrefix():
		return len(paused) + 1, 0
	// next not blank & not indented. End the block.
	default:
		return len(paused), 0
	}
}

type AtxHeaderBlock struct{ BlockNeverContinue }

func (b *AtxHeaderBlock) Detect(start, second Line) (consume, pause int) {
	if bytes.HasPrefix(start, []byte("#")) {
		return 1, 0
	}
	return 0, 0
}

type QuoteBlock struct{}

func (b *QuoteBlock) Detect(start, second Line) (consume, pause int) {
	ltrim := bytes.TrimLeft(start, " ")
	if len(ltrim) > 0 && ltrim[0] == '>' {
		return 0, 1
	}
	return 0, 0
}
func (b *QuoteBlock) Continue(paused []Line, next Line) (consume, pause int) {
	// TODO(akavel): verify it's coded ok, it was converted from a different approach
	if next == nil {
		return len(paused), 0
	}
	if paused[0].isBlank() {
		if next.isBlank() ||
			next.hasFourSpacePrefix() ||
			bytes.TrimLeft(next, " ")[0] != '>' {
			return len(paused), 0
		}
	} else if !next.hasFourSpacePrefix() &&
		reHorizontalRule.Match(next) {
		return len(paused), 0
	}
	return len(paused), 1
}

type HorizontalRuleBlock struct{ BlockNeverContinue }

func (b *HorizontalRuleBlock) Detect(start, second Line) (consume, pause int) {
	if reHorizontalRule.Match(start) {
		return 1, 0
	}
	return 0, 0
}

type UnorderedListBlock struct {
	Starter []byte
}

func (b *UnorderedListBlock) Detect(start, second Line) (consume, pause int) {
	m := reUnorderedList.FindSubmatch(start)
	if m == nil {
		return 0, 0
	}
	b.Starter = m[1]
	return 0, 1
}
func (b *UnorderedListBlock) Continue(paused []Line, next Line) (consume, pause int) {
	if next == nil {
		return len(paused), 0
	}

	if paused[0].isBlank() {
		if next.isBlank() {
			return len(paused), 0
		}
		if !bytes.HasPrefix(next, b.Starter) &&
			// FIXME(akavel): spec refers to runes ("characters"), not bytes; fix this everywhere
			next.hasNonSpaceInPrefix(len(b.Starter)) {
			return len(paused), 0
		}
	} else {
		if !bytes.HasPrefix(next, b.Starter) &&
			next.hasNonSpaceInPrefix(len(b.Starter)) &&
			!next.hasFourSpacePrefix() &&
			(reUnorderedList.Match(next) ||
				reOrderedList.Match(next) ||
				reHorizontalRule.Match(next)) {
			return len(paused), 0
		}
	}
	return len(paused), 1
}

type OrderedListBlock struct {
	Starter []byte
}

func (b *OrderedListBlock) Detect(start, second Line) (consume, pause int) {
	m := reOrderedList.FindSubmatch(start)
	if m == nil {
		return 0, 0
	}
	b.Starter = m[1]
	return 0, 1
}
func (b *OrderedListBlock) Continue(paused []Line, next Line) (consume, pause int) {
	if next == nil {
		return len(paused), 0
	}

	if paused[0].isBlank() {
		if next.isBlank() {
			return len(paused), 0
		}
		if !reOrderedList.Match(next) &&
			next.hasNonSpaceInPrefix(len(b.Starter)) {
			return len(paused), 0
		}
	} else {
		if !reOrderedList.Match(next) &&
			next.hasNonSpaceInPrefix(len(b.Starter)) &&
			!next.hasFourSpacePrefix() &&
			(reUnorderedList.Match(next) ||
				reHorizontalRule.Match(next)) {
			return len(paused), 0
		}
	}
	return len(paused), 1
}

type ParagraphBlock struct {
	// NOTE: below fields must be set appropriately when creating a ParagraphBlock
	IsInBlockquote bool
	IsInList       bool
}

func (b *ParagraphBlock) Detect(start, second Line) (consume, pause int) {
	return 0, 1
}
func (b *ParagraphBlock) Continue(paused []Line, next Line) (consume, pause int) {
	if next == nil {
		return len(paused), 0
	}
	// TODO(akavel): support HTML parser & related interactions [#paragraph-line-sequence]
	if paused[0].isBlank() {
		return len(paused), 0
	}
	if !next.hasFourSpacePrefix() {
		if reHorizontalRule.Match(next) ||
			(b.IsInBlockquote && bytes.HasPrefix(bytes.TrimLeft(next, " "), []byte(">"))) ||
			(b.IsInList && reOrderedList.Match(next)) ||
			(b.IsInList && reUnorderedList.Match(next)) {
			return len(paused), 0
		}
	}
	return len(paused), 1
}
