package span

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/akavel/vfmd-go/utils"
)

type Detector interface {
	Detect(*Splitter) (consumed int)
}

var DefaultDetectors = []Detector{
	EscapedChar{},
	LinkTags{},
	EmphasisTags{},
	CodeTags{},
	ImageTags{},
	AutomaticLinks{},
	// TODO(akavel): HTMLTags{},
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
	reClosingTagWithoutAngle = regexp.MustCompile(`^\]\s*\(\s*([^\(\)<>\` + "`" + `\s]+)([\)\s][\s\S]*)$`)
	// e.g.: "] ( <http://example.net/?q=)>"...
	reClosingTagWithAngle = regexp.MustCompile(`^\]\s*\(\s*<([^<>\` + "`" + `]*)>([\)\s][\s\S]*)$`)

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
			ReferenceID: utils.Simplify(m[1]),
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
		linkURL := utils.DelWhites(string(m[1]))
		residual := m[2]
		title := ""
		t := reJustClosingParen.FindSubmatch(residual)
		if t == nil {
			t = reTitleAndClosingParen.FindSubmatch(residual)
		}
		if t != nil {
			if len(t) > 1 {
				attribs := t[1]
				unquoted := attribs[1 : len(attribs)-1]
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
		ReferenceID: utils.Simplify(s.Buf[begin.LinkStart:s.Pos]),
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
	i := 1
	for i < len(rest) && isEmph(rest[i]) {
		i++
	}
	indicator := rest[:i]
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
			s.Openings.Push(MaybeOpening{
				Tag:      tag,
				NodeType: EmphasisNode,
			})
		}
		return len(indicator)
	}

	// right-flanking; maybe a closing tag
	closingEmphasisTags(s, tags)
	return len(indicator)
}

func closingEmphasisTags(s *Splitter, tags [][]byte) {
	for len(tags) > 0 {
		// find topmost opening of matching emphasis type
		tag := tags[0]
		i := len(s.Openings) - 1
		for i >= 0 && (s.Openings[i].NodeType != EmphasisNode || !bytes.HasPrefix(s.Openings[i].Tag, tag[:1])) {
			i--
		}
		// no opening node of this type, try next tag
		if i == -1 {
			tags = tags[1:]
			continue
		}
		// cancel any unclosed openings inside the emphasis span
		s.Openings = s.Openings[:i+1]
		// "procedure for matching emphasis tag strings"
		tags[0] = matchEmphasisTag(s, tag)
		if len(tags[0]) == 0 {
			tags = tags[1:]
		}
	}
}
func matchEmphasisTag(s *Splitter, tag []byte) []byte {
	top := s.Openings.Peek()
	if len(top.Tag) > len(tag) {
		n := len(tag)
		s.Emit(top.Tag[len(top.Tag)-n:], EmphasisBegin{Level: n})
		s.Emit(tag, EmphasisEnd{Level: n})
		top.Tag = top.Tag[:len(top.Tag)-n]
		return nil
	}
	// now len(top.Tag) <= len(tag)
	n := len(top.Tag)
	s.Emit(top.Tag, EmphasisBegin{Level: n})
	s.Emit(tag[:n], EmphasisEnd{Level: n})
	s.Openings.Pop()
	return tag[n:]
}

func isEmph(c byte) bool { return c == '*' || c == '_' }

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

type EmphasisBegin struct{ Level int }
type EmphasisEnd struct{ Level int }

type CodeTags struct{}

func (CodeTags) Detect(s *Splitter) (consumed int) {
	rest := s.Buf[s.Pos:]
	if rest[0] != '`' {
		return 0
	}
	i := 1
	for i < len(rest) && rest[i] == '`' {
		i++
	}
	opening := rest[:i]
	// try to find a sequence of '`' with length exactly equal to 'opening'
	for {
		pos := bytes.Index(rest[i:], opening)
		if pos == -1 {
			return len(opening)
		}
		i = i + pos + len(opening)
		if i >= len(rest) || rest[i] != '`' {
			// found closing tag!
			code := rest[len(opening) : i-len(opening)]
			code = bytes.Trim(code, utils.Whites)
			s.Emit(rest[:i], Code{Code: code})
			return i
		}
		for i < len(rest) && rest[i] == '`' {
			// too many '`' character to match the opening; consume
			// them as code contents and search again further
			i++
		}
	}
}

type Code struct{ Code []byte }

type ImageTags struct{}

func (ImageTags) Detect(s *Splitter) (consumed int) {
	rest := s.Buf[s.Pos:]
	if !bytes.HasPrefix(rest, []byte(`![`)) {
		return 0
	}
	m := reImageTagStarter.FindSubmatch(rest)
	if m == nil {
		return 2
	}
	altText, residual := m[1], m[3]

	// e.g.: "] [ref id]" ?
	r := reImageRef.FindSubmatch(residual)
	if r != nil {
		tag := rest[:len(rest)-len(residual)+len(r[0])]
		refID := utils.Simplify(r[1])
		// NOTE(akavel): below refID resolution seems not in spec, but expected according to testdata/test/span_level/image{/expected,}/link_text_with_newline.* and makes sense to me as such.
		// TODO(akavel): send fix for this to the spec
		if refID == "" {
			refID = utils.Simplify(altText)
		}
		s.Emit(tag, Image{
			AltText:     altText,
			ReferenceID: refID,
		})
		return len(tag)
	}

	// e.g.: "] ("...
	consumed = imageParen(s, altText, len(rest)-len(residual), residual)
	if consumed > 0 {
		return consumed
	}

	// "neither of the above conditions"
	closing := residual[:1]
	r = reImageEmptyRef.FindSubmatch(residual)
	if r != nil {
		closing = r[0]
	}
	tag := rest[:len(rest)-len(residual)+len(closing)]
	s.Emit(tag, Image{
		ReferenceID: utils.Simplify(altText),
		AltText:     altText,
	})
	return len(tag)
}

var (
	reImageTagStarter = regexp.MustCompile(`^!\[(([^\\\[\]\` + "`" + `]|\\.)*)(\][\s\S]*)$`)
	reImageRef        = regexp.MustCompile(`^\]\s*\[(([^\\\[\]\` + "`" + `]|\\.)*)\]`)

	reImageParen           = regexp.MustCompile(`^\]\s*\(`)
	reImageURLWithoutAngle = regexp.MustCompile(`^\]\s*\(\s*([^\(\)<>\` + "`" + `\s]+)([\)\s][\s\S]*)$`)
	// NOTE(akavel): below regexp was in spec, but fixed so that final
	// capture matches the above pattern, and passes
	// "image/link_with_parenthesis" test case.
	// reImageURLWithAngle    = regexp.MustCompile(`^\]\s*\(\s*<([^<>\` + "`" + `]*)>([\)][\s\S]+)$`)
	// TODO(akavel): send below fix to the spec.
	reImageURLWithAngle = regexp.MustCompile(`^\]\s*\(\s*<([^<>\` + "`" + `]*)>([\)\s][\s\S]*)$`)

	reImageAttrParen = regexp.MustCompile(`^\s*\)`)
	reImageAttrTitle = regexp.MustCompile(`^\s*("(([^"\\\` + "`" + `]|\\.)*)"|'(([^'\\\` + "`" + `]|\\.)*)')\s*\)`)

	reImageEmptyRef = regexp.MustCompile(`^(\]\s*\[\s*\])`)
)

func imageParen(s *Splitter, altText []byte, prefix int, residual []byte) (consumed int) {
	// fmt.Println("imageParen @", s.Pos, string(s.Buf[s.Pos:s.Pos+prefix]))
	// e.g.: "] ("
	if !reImageParen.Match(residual) {
		// fmt.Println("no ](")
		return 0
	}
	// fmt.Println("yes ](")

	r := reImageURLWithoutAngle.FindSubmatch(residual)
	if r == nil {
		r = reImageURLWithAngle.FindSubmatch(residual)
	}
	if r == nil {
		// fmt.Println("no imgurl")
		return 0
	}
	// fmt.Println("yes imgurl")
	unprocessedSrc, attrs := r[1], r[2]

	a := reImageAttrParen.FindSubmatch(attrs)
	if a == nil {
		a = reImageAttrTitle.FindSubmatch(attrs)
	}
	if a == nil {
		// fmt.Println("no imgattr")
		return 0
	}
	// fmt.Println("yes imgattr")
	title := ""
	if len(a) >= 2 {
		unprocessedTitle := a[1][1 : len(a[1])-1]
		title = strings.Replace(string(unprocessedTitle), "\x0a", "", -1)
	}

	rest := s.Buf[s.Pos:]
	tag := rest[:prefix+(len(residual)-len(attrs))+len(a[0])]
	s.Emit(tag, Image{
		// TODO(akavel): keep only raw slices as fields, add methods to return processed strings
		URL:     utils.DelWhites(string(unprocessedSrc)),
		Title:   title,
		AltText: altText,
	})
	return len(tag)
}

type Image struct {
	ReferenceID string
	URL         string
	Title       string
	AltText     []byte
}

type AutomaticLinks struct{}

func (AutomaticLinks) Detect(s *Splitter) (consumed int) {
	rest := s.Buf[s.Pos:]
	if s.Pos > 0 && rest[0] != '<' {
		r, _ := utf8.DecodeLastRune(s.Buf[:s.Pos])
		if r == utf8.RuneError || !isWordSep(r) {
			return 0
		}
	}
	// "potential-auto-link-start-position"
	// e.g. "<http://example.net>"
	m := reURLWithinAngle.FindSubmatch(rest)
	// e.g. "<mailto:someone@example.net?subject=Hi+there>"
	if m == nil {
		m = reMailtoURLWithinAngle.FindSubmatch(rest)
	}
	if m != nil {
		url := utils.DelWhites(string(m[1]))
		s.Emit(m[0], AutoLink{
			URL:  url,
			Text: url,
		})
		return len(m[0])
	}

	// e.g.: "<someone@example.net>"
	m = reMailWithinAngle.FindSubmatch(rest)
	if m != nil {
		s.Emit(m[0], AutoLink{
			URL:  "mailto:" + string(m[1]),
			Text: string(m[1]),
		})
		return len(m[0])
	}

	// e.g.: "http://example.net"
	m = reURLWithoutAngle.FindSubmatch(rest)
	if m == nil {
		m = reMailtoURLWithoutAngle.FindSubmatch(rest)
	}
	if m != nil {
		scheme := m[1]
		tag := m[0]
		// remove any trailing "speculative-url-end" characters
		for len(tag) > 0 {
			r, n := utf8.DecodeLastRune(tag)
			if !isSpeculativeURLEnd(r) {
				break
			}
			tag = tag[:len(tag)-n]
		}
		if len(tag) <= len(scheme) {
			return len(scheme)
		}
		s.Emit(tag, AutoLink{
			URL:  string(tag),
			Text: string(tag),
		})
		return len(tag)
	}
	return 0
}

var (
	// NOTE(akavel): "(?i)" is non-capturing - it's a flag enabling
	// case-insensitive matching
	reURLWithinAngle        = regexp.MustCompile(`^(?i)<([a-z0-9\+\.\-]+:\/\/[^<> \` + "`" + `]+)>`)
	reMailtoURLWithinAngle  = regexp.MustCompile(`^(?i)<(mailto:[^<> \` + "`" + `]+)>`)
	reMailWithinAngle       = regexp.MustCompile(`^<([^\(\)\<\>\[\]\:\'\@\\\,\"\s\` + "`" + `]+@[^\(\)\<\>\[\]\:\'\@\\\,\"\s\` + "`" + `\.]+\.[^\(\)\<\>\[\]\:\'\@\\\,\"\s\` + "`" + `]+)>`)
	reURLWithoutAngle       = regexp.MustCompile(`^(?i)([a-z0-9\+\.\-]+:\/\/)[^<>\` + "`" + `\s]+`)
	reMailtoURLWithoutAngle = regexp.MustCompile(`^(?i)(mailto:)[^<>\` + "`" + `\s]+`)
)

func isWordSep(r rune) bool {
	return unicode.In(r,
		unicode.Zs, unicode.Zl, unicode.Zp,
		unicode.Pc, unicode.Pd, unicode.Ps, unicode.Pe, unicode.Pi, unicode.Pf, unicode.Po,
		unicode.Cc, unicode.Cf)
}
func isSpeculativeURLEnd(r rune) bool {
	return r != '\u002f' && isWordSep(r)
}

type AutoLink struct{ URL, Text string }
