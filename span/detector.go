package span

import (
	"regexp"
	"strings"
)

type Detector interface {
	Detect(*Splitter) (consumed int)
}

var DefaultDetectors = []Detector{
	EscapedChar{},
	LinkTags{},
}

type EscapedChar struct{}

func (EscapedChar) Detect(s *Splitter) (consumed int) {
	rest := s.Buf[s.Pos:]
	if len(rest) >= 2 && rest[0] == '\\' {
		return 2
	} else {
		return 0
	}
}

type LinkTags struct{}

func (LinkTags) Detect(s *Splitter) (consumed int) {
	// [#procedure-for-identifying-link-tags]
	c := s.Buf[s.Pos]
	if c != '[' && c != ']' {
		return 0
	}
	// "opening link tag"?
	if c == '[' {
		s.Openings.Push(MaybeOpening{
			Tag:       s.Buf[s.Pos : s.Pos+1],
			NodeType:  LinkNode,
			LinkStart: s.Pos + 1,
		})
		return 1
	}
	// "closing link tag", c==']'
	return LinkTags{}.closingLinkTag(s)
}

var (
	// e.g.: "] [ref id]"
	reClosingTagRef = regexp.MustCompile(`^\]\s*\[(([^\\\[\]\` + "`" + `]|\\.)+)\]`)
	// e.g.: "] (http://www.example.net"...
	reClosingTagWithoutAngle = regexp.MustCompile(`^\]\s*\(\s*([^\(\)<>\` + "`" + `\s]+)([\)\s].*)$`)
	// e.g.: "] ( <http://example.net/?q=)>"...
	reClosingTagWithAngle = regexp.MustCompile(`^\]\s*\(\s*<([^<>\` + "`" + `]*)>([\)\s].*)$`)

	reJustClosingParen     = regexp.MustCompile(`^\s*\)`)
	reTitleAndClosingParen = regexp.MustCompile(`^\s*("(([^\\"\` + "`" + `]|\\.)*)"|'(([^\\'\` + "`" + `]|\\.)*)')\s*\)`)

	reEmptyRef = regexp.MustCompile(`^(\]\s*\[\s*\])`)
)

func (LinkTags) closingLinkTag(s *Splitter) (consumed int) {
	if s.Openings.NullTopmostOfType(LinkNode) {
		return 1 // consume the ']'
	}
	rest := s.Buf[s.Pos:]

	// e.g.: "] [ref id]" ?
	m := reClosingTagRef.FindSubmatch(rest)
	if m != nil {
		// cancel all unclosed spans inside the link
		for s.Openings.Peek().NodeType != LinkNode {
			s.Openings.Pop()
		}
		// emit a link
		s.Emit(s.Openings.Peek().Tag, LinkBegin{
			ReferenceID: Simplify(string(m[1])),
		})
		s.Emit(m[0], LinkEnd{})
		s.Openings.Pop()
		// cancel all unclosed links
		s.Openings.deleteLinks()
		return len(m[0])
	}

	// e.g.: "] (http://www.example.net"... ?
	m = reClosingTagWithoutAngle.FindSubmatch(rest)
	if m == nil {
		// e.g.: "] ( <http://example.net/?q=)>"... ?
		m = reClosingTagWithAngle.FindSubmatch(rest)
	}
	if m != nil {
		linkURL := DelWhitespace(string(m[1]))
		residual := m[2]
		title := ""
		t := reJustClosingParen.FindSubmatch(residual)
		if t == nil {
			t = reTitleAndClosingParen.FindSubmatch(residual)
		}
		if t != nil {
			if len(t) > 1 {
				attribs := t[1]
				unquoted := attribs[1 : len(attribs)-2]
				title = strings.Replace(string(unquoted), "\u000a", "", -1)
			}
			// cancel all unclosed spans inside the link
			for s.Openings.Peek().NodeType != LinkNode {
				s.Openings.Pop()
			}
			// emit a link
			s.Emit(s.Openings.Peek().Tag, LinkBegin{
				URL:   linkURL,
				Title: title,
			})
			s.Emit(t[0], LinkEnd{})
			s.Openings.Pop()
			// cancel all unclosed links
			s.Openings.deleteLinks()
			return len(rest) - len(residual) + len(t[0])
		}
	}

	m = reEmptyRef.FindSubmatch(rest)
	if m == nil {
		m = [][]byte{rest[:1]}
	}
	// cancel all unclosed spans inside the link
	for s.Openings.Peek().NodeType != LinkNode {
		s.Openings.Pop()
	}
	// emit a link
	begin := s.Openings.Peek()
	s.Emit(begin.Tag, LinkBegin{
		ReferenceID: Simplify(string(s.Buf[begin.LinkStart:s.Pos])),
	})
	s.Emit(m[0], LinkEnd{})
	s.Openings.Pop()
	// cancel all unclosed links
	s.Openings.deleteLinks()
	return len(m[0])
}

// FIXME(akavel): test if this works as expected
var whitespaceDeleter = strings.NewReplacer("\u0009", "",
	"\u000a", "",
	"\u000c", "",
	"\u000d", "",
	"\u0020", "")

func DelWhitespace(s string) string {
	return whitespaceDeleter.Replace(s)
}

type LinkBegin struct{ ReferenceID, URL, Title string }
type LinkEnd struct{}

func Simplify(s string) string {
	panic("NIY")
}
