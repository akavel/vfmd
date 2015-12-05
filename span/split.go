package span

import (
	"io"
	"sort"

	"gopkg.in/akavel/vfmd.v0/md"
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

type Context struct {
	Region   md.Region
	Pos      int
	Openings OpeningsStack
	tags     []PositionedTag
}

func (c *Context) Reader() io.Reader {
	return c.Region.Reader(c.Pos)
}

type PositionedTag struct {
	Tag md.Tag
	Pos int
}

func Parse(buf []byte, detectors []Detector) []md.Tag {
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
	sort.Sort(sortedTags(s.tags))
	return s.Spans
}

type sortedTags []PositionedTag

func (s sortedTags) Len() int           { return len(s) }
func (s sortedTags) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortedTags) Less(i, j int) bool { return s[i].pos < s[j].pos }

func (s *Context) Emit(pos int, tag md.Tag) {
	s.tags = append(s.tags, PositionedTag{Pos: pos, Tag: tag})
}
