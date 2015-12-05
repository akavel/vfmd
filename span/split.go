package span

import "sort"

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
	// Pos is a subslice of the original input buffer
	Pos []byte
	Tag interface{}
}

type Context struct {
	Buf      []byte
	Pos      int
	Openings OpeningsStack
	Spans    []Span
}

func Parse(buf []byte, detectors []Detector) []Span {
	if detectors == nil {
		detectors = DefaultDetectors
	}
	s := Context{Buf: buf}
walk:
	for s.Pos < len(s.Buf) {
		for _, d := range detectors {
			consumed := d.Detect(&s)
			if consumed > 0 {
				// fmt.Printf("DBG %T consumed %v at %v\t[%q]\n",
				// 	d, consumed, s.Pos, string(s.Buf[s.Pos:s.Pos+consumed]))
				// FIXME(akavel): if new spans emitted, verify no errors on span.OffsetIn(buf)
				s.Pos += consumed
				continue walk
			}
		}
		s.Pos++
	}
	sort.Sort(sortedSpans(s.Spans))
	return s.Spans
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

func (s *Context) Emit(slice []byte, tag interface{}) {
	s.Spans = append(s.Spans, Span{slice, tag})
}
