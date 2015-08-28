package span

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Detector interface {
	Detect(*Splitter) (consumed int)
}

var DefaultDetectors = []Detector{
	EscapedChar{},
	LinkTags{},
	EmphasisTags{},
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

	// e.g.: "] []" ?
	m = reEmptyRef.FindSubmatch(rest)
	if m == nil {
		// just: "]"
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

type LinkBegin struct{ ReferenceID, URL, Title string }
type LinkEnd struct{}

type EmphasisTags struct{}

func (EmphasisTags) Detect(s *Splitter) (consumed int) {
	rest := s.Buf[s.Pos:]
	if !isEmph(rest[0]) {
		return 0
	}
	// find substring composed solely of '*' and '_'
	indicator := rest[:1]
	for i := len(indicator); i < len(rest); i++ {
		if !isEmph(rest[i]) {
			break
		}
		indicator = rest[:i+1]
	}
	// "right-fringe-mark"
	r, _ := utf8.DecodeRune(rest[len(indicator):])
	rightFringe := emphasisFringeRank(r)
	r, _ = utf8.DecodeLastRune(s.Buf[:s.Pos])
	leftFringe := emphasisFringeRank(r)
	// <0 means "left-flanking", >0 "right-flanking", 0 "non-flanking"
	flanking := leftFringe - rightFringe
	if flanking == 0 {
		return len(indicator)
	}
	// split into "emphasis-tag-strings" - subslices of the same char
	tags := [][]byte{}
	prev := 0
	for curr := 1; curr <= len(indicator); curr++ {
		if curr == len(indicator) || indicator[curr] != indicator[prev] {
			tags = append(tags, indicator[prev:curr])
			prev = curr
		}
	}
	// left-flanking? if yes, add some openings
	if flanking < 0 {
		for _, tag := range tags {
			// TODO(akavel): leave just one EmphasisNode value and compare specific subtypes based on Tag[0]
			typ := AsteriskEmphasisNode
			if tag[0] == '_' {
				typ = UnderscoreEmphasisNode
			}
			s.Openings.Push(MaybeOpening{
				Tag:      tag,
				NodeType: typ,
			})
		}
		return len(indicator)
	}

	panic("NIY")
}

func isEmph(c byte) { return c == '*' || c == '_' }

func emphasisFringeRank(r rune) int {
	switch {
	case r == utf8.RuneError:
		// NOTE(akavel): not sure if that's really correct
		return 0
	case unicode.In(r, unicode.Zs, unicode.Zl, unicode.Zp, unicode.Cc, unicode.Cf):
		return 0
	case unicode.In(r, unicode.Pc, unicode.Pd, unicode.Ps, unicode.Pe, unicode.Pi, unicode.Pf, unicode.Po, unicode.Sc, unicode.Sk, unicode.Sm, unicode.So):
		return 1
	default:
		return 2
	}
}
