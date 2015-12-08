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
	Pos RegionPos
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
	Pos      RegionPos
	Openings OpeningsStack
	tags     []positionedTag
}

func (c *Context) reader() regionReader {
	return regionReader{c.Region, c.Pos}
}
func (c *Context) readFull(b []byte) bool {
	n, _ := io.ReadFull(c.reader(), b)
	return n == len(b)
}

type RegionPos struct {
	Run, Byte int
}

type positionedTag struct {
	tag md.Tag
	pos RegionPos
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

type sortedTags []positionedTag

func (s sortedTags) Len() int      { return len(s) }
func (s sortedTags) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortedTags) Less(i, j int) bool {
	if s[i].pos.Run != s[j].pos.Run {
		return s[i].pos.Run < s[j].pos.Run
	}
	return s[i].pos.Byte < s[j].pos.Byte
}
func (s *Context) Emit(pos RegionPos, tag md.Tag) {
	s.tags = append(s.tags, positionedTag{pos: pos, tag: tag})
}
