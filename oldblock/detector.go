package block // import "gopkg.in/akavel/vfmd.v0/oldblock"

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/akavel/vfmd.v0/utils"
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
	// This function is guaranteed to be run after the line has been
	// reported by Detect or Continue as consumed, and after all
	// preceding lines were PostProcessed, and before any subsequent
	// lines were PostProcessed. After last line, additional call with nil
	// argument is done.
	PostProcess(Line)
}

var (
	reUnorderedList  = regexp.MustCompile(`^( *[\*\-\+] +)[^ ]`)
	reOrderedList    = regexp.MustCompile(`^( *([0-9]+)\. +)[^ ]`)
	reHorizontalRule = regexp.MustCompile(`^ *((\* *\* *\* *[\* ]*)|(\- *\- *\- *[\- ]*)|(_ *_ *_ *[_ ]*))$`)
)

type Spans [][]byte

func (s Spans) GetSpans() Spans { return s }

type Blocks []Block

func (b Blocks) GetBlocks() Blocks { return b }

type NeverContinue struct{}

func (NeverContinue) Continue([]Line, Line) (consume, pause int) { return 0, 0 }

type NoPostProcess struct{}

func (NoPostProcess) PostProcess(Line) {}

// DefaultDetectors contains the list of default detectors in order in which
// they should be normally applied.
var DefaultDetectors []Detector = []Detector{
	Null{},
	&ReferenceResolution{},
	&SetextHeader{},
	&Code{},
	&AtxHeader{},
	&Quote{},
	HorizontalRule{},
	&UnorderedList{},
	&OrderedList{},
	&Paragraph{},
}

type Null struct {
	NeverContinue
	NoPostProcess
}

func (Null) Detect(start, second Line) (consume, pause int) {
	if start.isBlank() {
		return 1, 0
	}
	return 0, 0
}

type ReferenceResolution struct {
	NeverContinue
	NoPostProcess

	ReferenceID    string
	LinkURL        string
	TitleContainer string
	LinkTitle      string

	UnprocessedReferenceID        []byte
	RefValueSequence              []byte
	UnprocessedURL                []byte
	RefDefinitionTrailingSequence []byte
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
	b.UnprocessedReferenceID = m[1]
	b.ReferenceID = utils.Simplify(b.UnprocessedReferenceID)
	b.RefValueSequence = m[9] // TODO(akavel): verify if right one
	re = regexp.MustCompile(`^ *([^ \<\>]+|\<[^\<\>]*\>)( .*)?$`)
	m = re.FindSubmatch(b.RefValueSequence)
	if len(m) == 0 {
		return 0, 0
	}
	b.UnprocessedURL = m[1]
	{
		tmp := make([]byte, 0, len(b.UnprocessedURL))
		for _, c := range b.UnprocessedURL {
			if c != ' ' && c != '<' && c != '>' {
				tmp = append(tmp, c)
			}
		}
		b.LinkURL = string(tmp)
	}
	b.RefDefinitionTrailingSequence = m[2]

	// Detected ok. Now check if 1 or 2 lines.
	var nlines int
	re = regexp.MustCompile(`^ +("(([^"\\]|\\.)*)"|'(([^'\\]|\\.)*)'|\(([^\\\(\)]|\\.)*\)) *$`)
	if bytes.IndexAny(b.RefDefinitionTrailingSequence, " ") == -1 &&
		second != nil &&
		re.Match(second) {
		nlines = 2
		b.TitleContainer = string(second)
	} else {
		nlines = 1
		b.TitleContainer = string(b.RefDefinitionTrailingSequence)
	}

	re = regexp.MustCompile(`^\((([^\\\(\)]|\\.)*)\)`)
	if m := re.FindStringSubmatch(b.TitleContainer); len(m) != 0 {
		b.LinkTitle = m[1]
	}
	if s := HasQuotedStringPrefix(b.TitleContainer); s != "" {
		b.LinkTitle = s[1 : len(s)-1]
	}

	return nlines, 0
}

type SetextHeader struct {
	NeverContinue

	Level int
	Spans
}

func (s *SetextHeader) Detect(start, second Line) (consume, pause int) {
	if second == nil {
		return 0, 0
	}
	re := regexp.MustCompile(`^(-+|=+) *$`)
	if re.Match(second) {
		switch second[0] {
		case '=':
			s.Level = 1
		case '-':
			s.Level = 2
		}
		return 2, 0
	}
	return 0, 0
}

func (s *SetextHeader) PostProcess(line Line) {
	if line == nil {
		return
	}
	s.Spans = Spans{bytes.Trim(line, utils.Whites)}
}

type Code struct {
	Spans
}

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
func (c *Code) PostProcess(line Line) {
	line = trimLeftN(line, utils.Whites, 4)
	c.Spans = append(c.Spans, line)
}

func trimLeftN(s []byte, cutset string, nmax int) []byte {
	for nmax > 0 && len(s) > 0 && strings.IndexByte(cutset, s[0]) != -1 {
		nmax--
		s = s[1:]
	}
	return s
}

type AtxHeader struct {
	NeverContinue

	Level int
	Spans
}

func (AtxHeader) Detect(start, second Line) (consume, pause int) {
	if bytes.HasPrefix(start, []byte("#")) {
		return 1, 0
	}
	return 0, 0
}
func (a *AtxHeader) PostProcess(line Line) {
	if line == nil {
		return
	}
	text := bytes.Trim(line, "#")
	if len(text) == 0 {
		a.Level = len(text)
	} else {
		a.Level, _ = utils.OffsetIn(line, text)
	}
	if a.Level > 6 {
		a.Level = 6
	}
	a.Spans = Spans{bytes.Trim(text, utils.Whites)}
}

type Quote struct {
	Detectors []Detector
	splitter  Splitter

	Blocks
}

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
func (q *Quote) PostProcess(line Line) {
	if line == nil {
		// FIXME(akavel): handle error
		_ = q.splitter.Close()
		q.Blocks = q.splitter.Blocks
		return
	}

	text := bytes.TrimLeft(line, " ")
	switch {
	case bytes.HasPrefix(text, []byte("> ")):
		text = text[2:]
	case bytes.HasPrefix(text, []byte(">")):
		text = text[1:]
	}

	if q.splitter.Detectors == nil {
		q.splitter.Detectors = q.Detectors
	}
	// FIXME(akavel): handle error
	// FIXME(akavel): ignore final line if "empty"
	_ = q.splitter.WriteLine(line)
	q.Blocks = q.splitter.Blocks
}

type HorizontalRule struct {
	NeverContinue
	NoPostProcess
}

func (HorizontalRule) Detect(start, second Line) (consume, pause int) {
	if reHorizontalRule.Match(start) {
		return 1, 0
	}
	return 0, 0
}

type UnorderedListItem struct {
	Blocks
	splitter Splitter
}

type UnorderedList struct {
	Detectors []Detector

	Starter []byte
	Items   []UnorderedListItem
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
func (b *UnorderedList) PostProcess(line Line) {
	var current *UnorderedListItem
	if len(b.Items) > 0 {
		current = &b.Items[len(b.Items)-1]
	}

	if line == nil {
		if current != nil {
			// FIXME(akavel): handle errors
			_ = current.splitter.Close()
			current.Blocks = current.splitter.Blocks
		}
		return
	}

	if bytes.HasPrefix(line, b.Starter) {
		if current != nil {
			// FIXME(akavel): handle errors
			_ = current.splitter.Close()
			current.Blocks = current.splitter.Blocks
		}
		b.Items = append(b.Items, UnorderedListItem{})
		current = &b.Items[len(b.Items)-1]
		current.splitter.Detectors = b.Detectors
		// FIXME(akavel): handle errors
		_ = current.splitter.WriteLine(line[len(b.Starter):])
		return
	}

	_ = current.splitter.WriteLine(trimLeftN(line, " ", len(b.Starter)))
}

type OrderedListItem struct {
	Blocks
	splitter Splitter
}

type OrderedList struct {
	Detectors []Detector

	Starter     []byte
	FirstNumber int
	Items       []OrderedListItem
}

func (b *OrderedList) Detect(start, second Line) (consume, pause int) {
	m := reOrderedList.FindSubmatch(start)
	if m == nil {
		return 0, 0
	}
	b.Starter = m[1]
	b.FirstNumber, _ = strconv.Atoi(string(m[2]))
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
func (b *OrderedList) PostProcess(line Line) {
	var current *OrderedListItem
	if len(b.Items) > 0 {
		current = &b.Items[len(b.Items)-1]
	}

	if line == nil {
		if current != nil {
			// FIXME(akavel): handle errors
			_ = current.splitter.Close()
			current.Blocks = current.splitter.Blocks
		}
		return
	}

	m := reOrderedList.FindSubmatch(line)
	if m != nil {
		text := bytes.TrimLeft(m[1], " ")
		spaces, _ := utils.OffsetIn(m[1], text)
		if spaces >= len(b.Starter) {
			m = nil
		}
	}
	if m != nil {
		if current != nil {
			// FIXME(akavel): handle errors
			_ = current.splitter.Close()
			current.Blocks = current.splitter.Blocks
		}
		b.Items = append(b.Items, OrderedListItem{})
		current = &b.Items[len(b.Items)-1]
		current.splitter.Detectors = b.Detectors
		// FIXME(akavel): handle errors
		_ = current.splitter.WriteLine(line[len(m[1]):])
		return
	}

	_ = current.splitter.WriteLine(trimLeftN(line, " ", len(b.Starter)))
}

type Paragraph struct {
	// FIXME(akavel): below fields must be set appropriately when creating a Paragraph
	InQuote bool
	InList  bool

	Spans
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
func (b *Paragraph) PostProcess(line Line) {
	if line == nil {
		if n := len(b.Spans); n > 0 {
			b.Spans[n-1] = bytes.TrimRight(b.Spans[n-1], utils.Whites)
		}
		return
	}

	if len(b.Spans) == 0 {
		line = bytes.TrimLeft(line, utils.Whites)
	}
	b.Spans = append(b.Spans, line)
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
