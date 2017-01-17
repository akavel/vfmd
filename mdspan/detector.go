package mdspan // import "gopkg.in/akavel/vfmd.v0/mdspan"

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/mdutils"
)

type Detector interface {
	Detect(*Context) (consumed int)
}

type DetectorFunc func(*Context) (consumed int)

func (d DetectorFunc) Detect(s *Context) (consumed int) { return d(s) }

var DefaultDetectors = []Detector{
	DetectorFunc(DetectEscapedChar),
	DetectorFunc(DetectLink),
	DetectorFunc(DetectEmphasis),
	DetectorFunc(DetectCode),
	DetectorFunc(DetectImage),
	DetectorFunc(DetectAutomaticLink),
	// TODO(akavel): DetectHTML,
}

func DetectEscapedChar(s *Context) (consumed int) {
	if mdutils.HasPrefix(s.Suffix, []byte{'\\'}) && mdutils.Len(s.Suffix) >= 2 {
		return 2
	} else {
		return 0
	}
}

func DetectLink(s *Context) (consumed int) {
	// [#procedure-for-identifying-link-tags]
	// "opening link tag"?
	if mdutils.HasPrefix(s.Suffix, []byte{'['}) {
		s.Openings.Push(MaybeOpening{
			Tag: "[",
			Pos: copyReg(s.Suffix, 0, 1),
		})
		return 1
	}
	if !mdutils.HasPrefix(s.Suffix, []byte{']'}) {
		return 0
	}
	// "closing link tag", c==']'
	return closingLinkTag(s)
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

func closingLinkTag(s *Context) (consumed int) {
	// TODO(akavel): refactor to use s.PopTo (or not, for efficiency?)
	if s.Openings.NullTopmostTagged("[") {
		return 1 // consume the ']'
	}

	// e.g.: "] [ref id]" ?
	m := mdutils.FindSubmatch(s.Suffix, reClosingTagRef)
	if m != nil {
		// cancel all unclosed spans inside the link
		for s.Openings.Peek().Tag != "[" {
			s.Openings.Pop()
		}
		// emit a link
		opening := s.Openings.Peek()
		s.Emit(opening.Pos, md.Link{
			ReferenceID: mdutils.SimplifyReg(m[1]),
			RawEnd:      md.Raw(m[0]),
		}, false)
		s.Emit(m[0], md.End{}, false)
		s.Openings.Pop()
		// cancel all unclosed links
		s.Openings.deleteLinks()
		return mdutils.Len(m[0])
	}

	// e.g.: "] (http://www.example.net"... ?
	m = mdutils.FindSubmatch(s.Suffix, reClosingTagWithoutAngle)
	if m == nil {
		// e.g.: "] ( <http://example.net/?q=)>"... ?
		m = mdutils.FindSubmatch(s.Suffix, reClosingTagWithAngle)
	}
	if m != nil {
		linkURL := mdutils.DelWhites(mdutils.String(m[1]))
		residual := m[2]
		title := ""
		t := mdutils.FindSubmatch(residual, reJustClosingParen)
		if t == nil {
			t = mdutils.FindSubmatch(residual, reTitleAndClosingParen)
		}
		if t != nil {
			if len(t) > 1 {
				attribs := mdutils.String(t[1])
				unquoted := attribs[1 : len(attribs)-1]
				title = strings.Replace(unquoted, "\u000a", "", -1)
			}
			// cancel all unclosed spans inside the link
			for s.Openings.Peek().Tag != "[" {
				s.Openings.Pop()
			}
			// emit a link
			opening := s.Openings.Peek()
			closingReg := copyReg(s.Suffix, 0,
				mdutils.Len(s.Suffix)-mdutils.Len(residual)+mdutils.Len(t[0]))
			s.Emit(opening.Pos, md.Link{
				URL:    linkURL,
				Title:  mdutils.DeEscape(title),
				RawEnd: md.Raw(closingReg),
			}, false)
			s.Emit(closingReg, md.End{}, false)
			s.Openings.Pop()
			// cancel all unclosed links
			s.Openings.deleteLinks()
			return mdutils.Len(closingReg)
		}
	}

	// e.g.: "] []" ?
	m = mdutils.FindSubmatch(s.Suffix, reEmptyRef)
	if m == nil {
		// just: "]"
		m = []md.Region{copyReg(s.Suffix, 0, 1)}
	}
	// cancel all unclosed spans inside the link
	for s.Openings.Peek().Tag != "[" {
		s.Openings.Pop()
	}
	// emit a link
	begin := s.Openings.Peek()
	// TODO(akavel): refIDReg := mdutils.Copy(s.Prefix)
	// TODO(akavel): mdutils.Skip(&refIDReg, begin.Pos+len(begin.Tag))
	s.Emit(begin.Pos, md.Link{
		// TODO(akavel): ReferenceID: mdutils.SimplifyReg(refIDReg),
		RawEnd: md.Raw(m[0]),
	}, false)
	s.Emit(m[0], md.End{}, false)
	s.Openings.Pop()
	// cancel all unclosed links
	s.Openings.deleteLinks()
	return mdutils.Len(m[0])
}

var reEmphasis = regexp.MustCompile(`^([_*]+)(.*)$`)

func DetectEmphasis(s *Context) (consumed int) {
	m := mdutils.FindSubmatch(s.Suffix, reEmphasis)
	if m == nil {
		return 0
	}
	indicator := m[1]
	// "right-fringe-mark"
	r, _ := mdutils.DecodeRune(m[2])
	rightFringe := emphasisFringeRank(r)
	r, _ = mdutils.DecodeLastRune(s.Prefix)
	leftFringe := emphasisFringeRank(r)
	// <0 means "left-flanking", >0 "right-flanking", 0 "non-flanking"
	flanking := leftFringe - rightFringe
	if flanking == 0 {
		return mdutils.Len(indicator)
	}
	// split into "emphasis-tag-strings" - subslices of the same char
	tags := []md.Region{}
	{
		prev, remaining := rune(0), mdutils.Copy(indicator)
		for !mdutils.Empty(remaining) {
			r, sz := mdutils.DecodeRune(remaining)
			if r != prev {
				tags = append(tags, md.Region{})
			}
			prev = r
			mdutils.Move(&tags[len(tags)-1], &remaining, sz)
		}
	}
	// left-flanking? if yes, add some openings
	if flanking < 0 {
		for _, tag := range tags {
			s.Openings.Push(MaybeOpening{
				Tag: mdutils.String(tag),
				Pos: tag,
			})
		}
		return mdutils.Len(indicator)
	}

	// right-flanking; maybe a closing tag
	closingEmphasisTags(s, tags)
	return mdutils.Len(indicator)
}

func closingEmphasisTags(s *Context, tags []md.Region) {
	// TODO(akavel): refactor to use s.PopTo
	for len(tags) > 0 {
		// find topmost opening of matching emphasis type
		tag := tags[0]
		i := len(s.Openings) - 1
		for i >= 0 {
			r, _ := mdutils.DecodeRune(tag)
			if strings.HasPrefix(s.Openings[i].Tag, string(r)) {
				break
			}
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
		if mdutils.Empty(tags[0]) {
			tags = tags[1:]
		}
	}
}
func matchEmphasisTag(s *Context, tag md.Region) md.Region {
	top := s.Openings.Peek()
	if len(top.Tag) > mdutils.Len(tag) {
		n := mdutils.Len(tag)
		prefix, suffix := md.Region{}, top.Pos
		mdutils.Move(&prefix, &suffix, len(top.Tag)-n)
		top.Pos = prefix
		top.Tag = mdutils.String(prefix)
		s.Emit(suffix, md.Emphasis{Level: n}, false)
		s.Emit(tag, md.End{}, false)
		return nil
	} else { // len(top.Tag) <= len(tag)
		n := len(top.Tag)
		prefix, suffix := md.Region{}, tag
		mdutils.Move(&prefix, &suffix, n)
		s.Emit(top.Pos, md.Emphasis{Level: n}, false)
		s.Emit(prefix, md.End{}, false)
		s.Openings.Pop()
		return suffix
	}
}

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

var reCode = regexp.MustCompile("^(`+)(.*)$")

func DetectCode(s *Context) (consumed int) {
	m := mdutils.FindSubmatch(s.Suffix, reCode)
	if m == nil {
		return 0
	}
	opening := m[1]
	// try to find a sequence of '`' with length exactly equal to 'opening'
	re := regexp.MustCompile("(.*?(" + mdutils.String(opening) + "))([^`]|$)")
	m = mdutils.FindSubmatch(m[2], re)
	if m == nil {
		return mdutils.Len(opening)
	}
	// TODO(akavel): s.Emit(..., md.Code{...}, true)
	var (
		nopen = mdutils.Len(opening)
		n     = nopen + mdutils.Len(m[1])
		all   = mdutils.CopyN(s.Suffix, n)
		code  = mdutils.String(all)[nopen : n-nopen]
	)
	s.Emit(all, md.Code{Code: []byte(code)}, true)
	return n
}

func DetectImage(s *Context) (consumed int) {
	if !mdutils.HasPrefix(s.Suffix, []byte(`![`)) {
		return 0
	}
	m := mdutils.FindSubmatch(s.Suffix, reImageTagStarter)
	if m == nil {
		return 2
	}
	altText, residual := m[1], m[3]

	// e.g.: "] [ref id]" ?
	r := mdutils.FindSubmatch(residual, reImageRef)
	if r != nil {
		// FIXME(akavel): optimize below 2 lines
		tag := mdutils.Copy(s.Suffix)
		mdutils.Limit(&tag, mdutils.Len(s.Suffix)-mdutils.Len(residual)+mdutils.Len(r[0]))
		refID := mdutils.SimplifyReg(r[1])
		// NOTE(akavel): below refID resolution seems not in spec, but expected according to testdata/test/span_level/image{/expected,}/link_text_with_newline.* and makes sense to me as such.
		// TODO(akavel): send fix for this to the spec
		if refID == "" {
			refID = mdutils.SimplifyReg(altText)
		}
		s.Emit(tag, md.Image{
			AltText:     mdutils.DeEscape(mdutils.String(altText)),
			ReferenceID: refID,
			RawEnd:      md.Raw(r[0]),
		}, true)
		return mdutils.Len(tag)
	}

	// e.g.: "] ("...
	consumed = imageParen(s, altText, mdutils.Len(s.Suffix)-mdutils.Len(residual), residual)
	if consumed > 0 {
		return consumed
	}

	// "neither of the above conditions"
	closing := residual[:1]
	r = mdutils.FindSubmatch(residual, reImageEmptyRef)
	if r != nil {
		closing = r[0]
	}
	// FIXME(akavel): optimize below 2 lines
	tag := mdutils.Copy(s.Suffix)
	mdutils.Limit(&tag, mdutils.Len(s.Suffix)-mdutils.Len(residual)+mdutils.Len(closing))
	s.Emit(tag, md.Image{
		ReferenceID: mdutils.SimplifyReg(altText),
		AltText:     mdutils.DeEscape(mdutils.String(altText)),
		RawEnd:      md.Raw(closing),
	}, true)
	return mdutils.Len(tag)
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

func imageParen(s *Context, altText md.Region, prefix int, residual md.Region) (consumed int) {
	// fmt.Println("imageParen @", s.Pos, string(s.Buf[s.Pos:s.Pos+prefix]))
	// e.g.: "] ("
	if !mdutils.Match(residual, reImageParen) {
		// fmt.Println("no ](")
		return 0
	}
	// fmt.Println("yes ](")

	r := mdutils.FindSubmatch(residual, reImageURLWithoutAngle)
	if r == nil {
		r = mdutils.FindSubmatch(residual, reImageURLWithAngle)
	}
	if r == nil {
		// fmt.Println("no imgurl")
		return 0
	}
	// fmt.Println("yes imgurl")
	unprocessedSrc, attrs := r[1], r[2]

	a := mdutils.FindSubmatch(attrs, reImageAttrParen)
	if a == nil {
		a = mdutils.FindSubmatch(attrs, reImageAttrTitle)
	}
	if a == nil {
		// fmt.Println("no imgattr")
		return 0
	}
	// fmt.Println("yes imgattr")
	title := ""
	if len(a) >= 2 {
		raw := mdutils.String(a[1])
		unprocessedTitle := raw[1 : len(raw)-1]
		title = strings.Replace(unprocessedTitle, "\x0a", "", -1)
	}

	tag := mdutils.Copy(s.Suffix)
	mdutils.Limit(&tag, prefix+(mdutils.Len(residual)-mdutils.Len(attrs))+mdutils.Len(a[0]))
	rawEnd := mdutils.Copy(tag)
	mdutils.Skip(&rawEnd, prefix)
	s.Emit(tag, md.Image{
		// TODO(akavel): keep only raw slices as fields, add methods to return processed strings
		URL:     mdutils.DelWhites(mdutils.String(unprocessedSrc)),
		Title:   mdutils.DeEscape(title),
		AltText: mdutils.DeEscape(mdutils.String(altText)),
		RawEnd:  md.Raw(rawEnd),
	}, true)
	return mdutils.Len(tag)
}

func DetectAutomaticLink(s *Context) (consumed int) {
	if !mdutils.Empty(s.Prefix) && !mdutils.HasPrefix(s.Suffix, []byte{'<'}) {
		r, _ := mdutils.DecodeLastRune(s.Prefix)
		if r == utf8.RuneError || !isWordSep(r) {
			return 0
		}
	}
	// "potential-auto-link-start-position"
	// e.g. "<http://example.net>"
	m := mdutils.FindSubmatch(s.Suffix, reURLWithinAngle)
	// e.g. "<mailto:someone@example.net?subject=Hi+there>"
	if m == nil {
		m = mdutils.FindSubmatch(s.Suffix, reMailtoURLWithinAngle)
	}
	if m != nil {
		url := mdutils.DelWhites(mdutils.String(m[1]))
		s.Emit(m[0], md.AutomaticLink{
			URL:  url,
			Text: url,
		}, true)
		return mdutils.Len(m[0])
	}

	// e.g.: "<someone@example.net>"
	m = mdutils.FindSubmatch(s.Suffix, reMailWithinAngle)
	if m != nil {
		s.Emit(m[0], md.AutomaticLink{
			URL:  "mailto:" + mdutils.String(m[1]),
			Text: mdutils.String(m[1]),
		}, true)
		return mdutils.Len(m[0])
	}

	// e.g.: "http://example.net"
	m = mdutils.FindSubmatch(s.Suffix, reURLWithoutAngle)
	if m == nil {
		m = mdutils.FindSubmatch(s.Suffix, reMailtoURLWithoutAngle)
	}
	if m != nil {
		scheme := m[1]
		tag := mdutils.Copy(m[0])
		// remove any trailing "speculative-url-end" characters
		for mdutils.Len(tag) > 0 {
			r, n := mdutils.DecodeLastRune(tag)
			if !isSpeculativeURLEnd(r) {
				break
			}
			mdutils.Limit(&tag, mdutils.Len(tag)-n)
		}
		if mdutils.Len(tag) <= mdutils.Len(scheme) {
			return mdutils.Len(scheme)
		}
		s.Emit(tag, md.AutomaticLink{
			URL:  mdutils.String(tag),
			Text: mdutils.String(tag),
		}, true)
		return mdutils.Len(tag)
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

// equivalent of concat(r[*].Bytes)[off:][:n]
func copyReg(r md.Region, off, n int) md.Region {
	// TODO(akavel): optimize
	r = mdutils.Copy(r)
	mdutils.Skip(&r, off)
	mdutils.Limit(&r, n)
	return r
}
