package block

import (
	"bytes"
	"regexp"
)

// TODO(akavel): add tests for blocks

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

type Detector interface {
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
	//
	// FIXME(mateuszc): make this comment complete and sane
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

type NeverContinue struct{}

func (NeverContinue) Continue([]Line, Line) (consume, pause int) { return 0, 0 }

// DefaultDetectors contains the list of default detectors in order in which
// they should be normally applied.
var DefaultDetectors []Detector = []Detector{
	Null{},
	&ReferenceResolution{},
	SetextHeader{},
	Code{},
	AtxHeader{},
	Quote{},
	HorizontalRule{},
	&UnorderedList{},
	&OrderedList{},
	Paragraph{},
}

type Null struct{ NeverContinue }

func (Null) Detect(start, second Line) (consume, pause int) {
	if start.isBlank() {
		return 1, 0
	}
	return 0, 0
}

type ReferenceResolution struct {
	NeverContinue
	unprocessedReferenceID        []byte
	refValueSequence              []byte
	unprocessedUrl                []byte
	refDefinitionTrailingSequence []byte
}

func (b *ReferenceResolution) Detect(start, second Line) (consume, pause int) {
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

type SetextHeader struct{ NeverContinue }

func (SetextHeader) Detect(start, second Line) (consume, pause int) {
	if second == nil {
		return 0, 0
	}
	re := regexp.MustCompile(`^(-+|=+) *$`)
	if re.Match(second) {
		return 2, 0
	}
	return 0, 0
}

type Code struct{}

func (Code) Detect(start, second Line) (consume, pause int) {
	if start.hasFourSpacePrefix() {
		return 1, 0
	}
	return 0, 0
}
func (Code) Continue(paused []Line, next Line) (consume, pause int) {
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

type AtxHeader struct{ NeverContinue }

func (AtxHeader) Detect(start, second Line) (consume, pause int) {
	if bytes.HasPrefix(start, []byte("#")) {
		return 1, 0
	}
	return 0, 0
}

type Quote struct{}

func (Quote) Detect(start, second Line) (consume, pause int) {
	ltrim := bytes.TrimLeft(start, " ")
	if len(ltrim) > 0 && ltrim[0] == '>' {
		return 0, 1
	}
	return 0, 0
}
func (Quote) Continue(paused []Line, next Line) (consume, pause int) {
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

type HorizontalRule struct{ NeverContinue }

func (HorizontalRule) Detect(start, second Line) (consume, pause int) {
	if reHorizontalRule.Match(start) {
		return 1, 0
	}
	return 0, 0
}

type UnorderedList struct {
	Starter []byte
}

func (b *UnorderedList) Detect(start, second Line) (consume, pause int) {
	m := reUnorderedList.FindSubmatch(start)
	if m == nil {
		return 0, 0
	}
	b.Starter = m[1]
	return 0, 1
}
func (b *UnorderedList) Continue(paused []Line, next Line) (consume, pause int) {
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

type OrderedList struct {
	Starter []byte
}

func (b *OrderedList) Detect(start, second Line) (consume, pause int) {
	m := reOrderedList.FindSubmatch(start)
	if m == nil {
		return 0, 0
	}
	b.Starter = m[1]
	return 0, 1
}
func (b *OrderedList) Continue(paused []Line, next Line) (consume, pause int) {
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

type Paragraph struct {
	// NOTE: below fields must be set appropriately when creating a Paragraph
	InQuote bool
	InList  bool
}

func (Paragraph) Detect(start, second Line) (consume, pause int) {
	return 0, 1
}
func (b Paragraph) Continue(paused []Line, next Line) (consume, pause int) {
	if next == nil {
		return len(paused), 0
	}
	// TODO(akavel): support HTML parser & related interactions [#paragraph-line-sequence]
	if paused[0].isBlank() {
		return len(paused), 0
	}
	if !next.hasFourSpacePrefix() {
		if reHorizontalRule.Match(next) ||
			(b.InQuote && bytes.HasPrefix(bytes.TrimLeft(next, " "), []byte(">"))) ||
			(b.InList && reOrderedList.Match(next)) ||
			(b.InList && reUnorderedList.Match(next)) {
			return len(paused), 0
		}
	}
	return len(paused), 1
}

/*
NOTES:

AtxHeader
 -> text-span-sequence

SetextHeader
 -> text-span-sequence

Quote
 -> process lines (strip certain prefix bytes)
  -> detect[defaults..., Paragraph{InQuote=1}]

UnorderedList
 -> detect[UnorderedItem]
  -> process lines (strip certain prefix bytes)
   -> detect[defaults..., Paragraph{InList=1}]

OrderedList
 -> detect[OrderedItem]
  -> process lines (strip certain prefix bytes)
   -> detect[defaults..., Paragraph{InList=1}]

Paragraph
 -> join & trim
  -> process as text-span-sequence

Block{
	block.Detector
	NLines int
	Nested []Block
}

// This function is guaranteed to be run after the line has been reported by
// Detect or Continue as consumed, and after all preceding lines were
// PostProcessed, and before any subsequent lines were PostProcessed.
PostProcess(Line)
*/
