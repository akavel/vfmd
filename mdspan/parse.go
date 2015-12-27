package mdspan

import (
	"sort"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/mdutils"
)

type NodeType int

type OpeningsStack []MaybeOpening
type MaybeOpening struct {
	Tag string
	Pos int
	// TODO(akavel): HTMLTag
}

func (s OpeningsStack) NullTopmostTagged(tag string) bool {
	for _, node := range s {
		if node.Tag == tag {
			return false
		}
	}
	// TODO(akavel): HTML
	return true
}

func (s *OpeningsStack) Push(o MaybeOpening) {
	(*s) = append(*s, o)
}
func (s OpeningsStack) Peek() *MaybeOpening {
	if len(s) == 0 {
		return nil
	}
	return &s[len(s)-1]
}
func (s *OpeningsStack) Pop() {
	if len(*s) > 0 {
		(*s) = (*s)[:len(*s)-1]
	}
}
func (s *OpeningsStack) PopTo(f func(*MaybeOpening) bool) (MaybeOpening, bool) {
	for i := len(*s) - 1; i >= 0; i-- {
		if !f(&(*s)[i]) {
			continue
		}
		// found, pop and return
		o := (*s)[i]
		(*s) = (*s)[:i]
		return o, true
	}
	return MaybeOpening{Pos: -1}, false
}

// deleteLinks cancels all unclosed links
func (s *OpeningsStack) deleteLinks() {
	filtered := make(OpeningsStack, 0, len(*s))
	for _, o := range *s {
		if o.Tag != "[" {
			filtered = append(filtered, o)
		}
	}
	*s = filtered
}

type Span struct {
	// Pos is a Region of the original input buffer
	Pos       md.Region
	Tag       md.Tag
	SelfClose bool
}

type Context struct {
	Prefix, Suffix md.Region
	Openings       OpeningsStack
	Spans          []Span
}

func Parse(r md.Region, detectors []Detector) []md.Tag {
	if detectors == nil {
		detectors = DefaultDetectors
	}
	s := Context{
		Prefix: md.Region{},
		Suffix: Copy(r),
	}
walk:
	for len(s.Suffix) > 0 {
		if len(s.Suffix[0]) == 0 {
			s.Suffix = s.Suffix[1:]
			continue
		}
		consumed := 0
		for _, d := range detectors {
			consumed = d.Detect(&s)
			if consumed > 0 {
				// fmt.Printf("DBG %T consumed %v at %v\t[%q]\n",
				// 	d, consumed, s.Pos, string(s.Buf[s.Pos:s.Pos+consumed]))
				// FIXME(akavel): if new spans emitted, verify no errors on span.OffsetIn(buf)
				break
			}
		}
		if consumed == 0 {
			consumed = 1
		}
		_, err := mdutils.Move(&s.Prefix, &s.Suffix, consumed)
		if err != nil {
			panic(err.Error())
		}
	}
	sort.Sort(sortedSpans(s.Spans))
	tags := []md.Tag{}
	endOffset := 0
	for _, span := range s.Spans {
		offset, _ := mdutils.OffsetIn(buf, span.Pos)
		if offset > endOffset {
			// FIXME(akavel): for every "  \n" sequence, insert an md.HardBreak tag
			tags = append(tags, mdutils.DeEscapeProse(md.Prose{
				// FIXME(akavel): fix Line in md.Run
				md.Run{-1, buf[endOffset:offset]},
			}))
		}
		tags = append(tags, span.Tag)
		if span.SelfClose {
			tags = append(tags, md.End{})
		}
		endOffset = offset + len(span.Pos)
	}
	if endOffset < len(buf) {
		tags = append(tags, mdutils.DeEscapeProse(md.Prose{
			// FIXME(akavel): fix Line in md.Run
			md.Run{-1, buf[endOffset:]},
		}))
	}
	return tags
}

type sortedSpans []Span

func (s sortedSpans) Len() int      { return len(s) }
func (s sortedSpans) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortedSpans) Less(i, j int) bool {
	iext, jext := s[i].Pos, s[j].Pos
	iext, jext = iext[:cap(iext)], jext[:cap(jext)]
	if &iext[cap(iext)-1] != &jext[cap(jext)-1] {
		// TODO(akavel): panic
		return false
	}
	return len(iext) > len(jext)
}

func (s *Context) Emit(slice []byte, tag interface{}, selfClose bool) {
	s.Spans = append(s.Spans, Span{slice, tag, selfClose})
}
