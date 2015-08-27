package span

import "errors"

type NodeType int

const (
	AsteriskEmphasisNode NodeType = iota
	UnderscoreEmphasisNode
	LinkNode
	// RawHTMLNode // TODO(akavel): HTML
)

type OpeningsStack []MaybeOpening
type MaybeOpening struct {
	Tag []byte
	NodeType
	LinkStart int
	// HTMLTag
}

func (s OpeningsStack) IsTopmostOfType(i int) bool {
	if i < 0 || i > len(s) {
		return false
	}
	for _, over := range s[i+1:] {
		if over.NodeType == s[i].NodeType {
			return false
		}
	}
	// TODO(akavel): HTML
	return true
}

func (s OpeningsStack) NullTopmostOfType(t NodeType) bool {
	for _, node := range s {
		if node.NodeType == t {
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
		if o.NodeType != LinkNode {
			filtered = append(filtered, o)
		}
	}
	*s = filtered
}

type Span struct {
	// Pos is a subslice of the original input buffer
	Pos []byte
}

func (span Span) OffsetIn(buf []byte) (int, error) {
	// one weird trick to check if one of two slices is subslice of the other
	extSpan, extBuf := span.Pos[:cap(span.Pos)], buf[:cap(buf)]
	if &extSpan[cap(extSpan)-1] != &extBuf[cap(extBuf)-1] {
		return -1, errors.New("vfmd-go: Span is not subslice of the buf provided to OffsetIn()")
	}
	return len(extBuf) - len(extSpan), nil
}

type Splitter struct {
	Buf      []byte
	Pos      int
	Openings OpeningsStack
	Spans    []Span
}

func (s *Splitter) Process(buf []byte) []Span {
	remaining := buf
	/*
		for len(remaining)>0 {
			consumed := 0
			for _, d := range detectors {
				if d.Detect(&consumed, remaining, buf) {
					// ...
				}
			}
			assert(consumed>0, consumed)
			remaining = remaining[consumed:]
		}
		// ...
	*/
}
