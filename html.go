package vfmd

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"regexp"
	"strings"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/mdutils"
)

func QuickHTML(w io.Writer, blocks []md.Tag) error {
	opt := htmlOpt{
		refs: htmlRefs(blocks),
	}
	tags := blocks
	for len(tags) > 0 {
		newtags, err := htmlBlock(tags, w, opt)
		if err != nil {
			return err
		}
		err = chkmoved(tags, newtags)
		if err != nil {
			return err
		}
		tags = newtags
	}
	return nil
}

func chkmoved(oldtags, newtags []md.Tag) error {
	if len(oldtags) == 0 || len(newtags) == 0 {
		return nil
	}
	if &oldtags[0] != &newtags[0] {
		return nil
	}
	return fmt.Errorf("vfmd: parsing failed to move over %T (%d tags remaining)",
		newtags[0], len(newtags))
}

type htmlLinkInfo struct {
	URL, Title string
}
type htmlOpt struct {
	refs                            map[string]htmlLinkInfo
	topPackedForP, bottomPackedForP bool
	itemEndForP                     int
}

func (opt htmlOpt) fillRef(refID string, ref *htmlLinkInfo) bool {
	newref, found := opt.refs[strings.ToLower(refID)]
	if !found {
		return false
	}
	ref.URL = newref.URL
	ref.Title = newref.Title
	return true
}

func htmlRefs(tags []md.Tag) map[string]htmlLinkInfo {
	m := map[string]htmlLinkInfo{}
	for _, t := range tags {
		r, ok := t.(md.ReferenceResolutionBlock)
		if !ok {
			continue
		}
		// TODO(akavel): make that properly case-insensitive for various languages (Turkish etc.)
		id := strings.ToLower(r.ReferenceID)
		_, found := m[id]
		if found {
			continue
		}
		m[id] = htmlLinkInfo{URL: r.URL, Title: r.Title}
	}
	return m
}

type htmlContext struct {
	w    io.Writer
	tags []md.Tag
	err  error
}

func (c *htmlContext) printf(format string, args ...interface{}) {
	if c.err != nil {
		return
	}
	_, c.err = fmt.Fprintf(c.w, format, args...)
}
func (c *htmlContext) spans(tags []md.Tag, opt htmlOpt) {
	if c.err != nil {
		return
	}
	c.tags, c.err = htmlSpans(tags, c.w, opt)
	if c.err == nil {
		c.err = chkmoved(tags, c.tags)
	}
}
func (c *htmlContext) blocks(tags []md.Tag, opt htmlOpt) {
	if c.err != nil {
		return
	}
	c.tags, c.err = htmlBlocks(tags, c.w, opt)
	if c.err == nil {
		c.err = chkmoved(tags, c.tags)
	}
}
func (c *htmlContext) items(tags []md.Tag, parentRegion md.Raw, opt htmlOpt) {
	if c.err != nil {
		return
	}
	c.tags, c.err = htmlItems(tags, c.w, parentRegion, opt)
	if c.err == nil {
		c.err = chkmoved(tags, c.tags)
	}
}
func (c *htmlContext) write(buf []byte) {
	if c.err != nil {
		return
	}
	_, c.err = c.w.Write(buf)
}

func htmlBlock(tags []md.Tag, w io.Writer, opt htmlOpt) ([]md.Tag, error) {
	c := htmlContext{w: w, tags: tags}
	switch t := tags[0].(type) {
	case md.AtxHeaderBlock:
		c.printf("<h%d>", t.Level)
		c.spans(tags[1:], opt)
		c.printf("</h%d>\n", t.Level)
		return c.tags, c.err
	case md.SetextHeaderBlock:
		c.printf("<h%d>", t.Level)
		c.spans(tags[1:], opt)
		c.printf("</h%d>\n", t.Level)
		return c.tags, c.err
	case md.NullBlock:
		c.printf("\n")
		return c.tags[2:], c.err
	case md.QuoteBlock:
		c.printf("<blockquote>\n  ")
		c.blocks(tags[1:], htmlOpt{refs: opt.refs})
		c.printf("</blockquote>\n")
		return c.tags, c.err
	case md.ParagraphBlock:
		n := len(t.Raw)
		no_p := opt.topPackedForP ||
			(opt.bottomPackedForP && t.Raw[n-1].Line == opt.itemEndForP)
		if !no_p {
			c.printf("<p>")
		}
		c.spans(tags[1:], opt)
		if !no_p {
			c.printf("</p>\n")
		}
		return c.tags, c.err
	case md.CodeBlock:
		c.printf("<pre><code>")
		for _, r := range t.Prose {
			c.printf("%s", html.EscapeString(string(r.Bytes)))
		}
		c.printf("</code></pre>\n")
		return c.tags[2:], c.err
	case md.HorizontalRuleBlock:
		c.printf("<hr />\n")
		return c.tags[2:], c.err
	case md.OrderedListBlock:
		var i int
		fmt.Sscanf(string(t.Starter.Bytes), "%d", &i)
		if i != 1 {
			c.printf("<ol start=\"%d\">\n", i)
		} else {
			c.printf("<ol>\n")
		}
		c.items(tags[1:], t.Raw, opt)
		c.printf("</ol>\n")
		return c.tags, c.err
	case md.UnorderedListBlock:
		c.printf("<ul>\n")
		c.items(tags[1:], t.Raw, opt)
		c.printf("</ul>\n")
		return c.tags, c.err
	case md.ReferenceResolutionBlock:
		return c.tags[2:], nil
	default:
		// TODO(akavel): return error's context (e.g. remaining tags?)
		return tags, fmt.Errorf("vfmd: block type %T not supported yet", t)
	}
}

func isBlank(line md.Run) bool {
	return len(bytes.Trim(line.Bytes, " \t\n")) == 0
}

func htmlItems(tags []md.Tag, w io.Writer, parentRegion md.Raw, opt htmlOpt) ([]md.Tag, error) {
	c := htmlContext{w: w, tags: tags}
	for {
		if (c.tags[0] == md.End{}) {
			return c.tags[1:], nil
		}

		t := c.tags[0].(md.ItemBlock)
		opt := htmlOpt{refs: opt.refs}
		// top-packed?
		n, m := len(t.Raw), len(parentRegion)
		ifirst, ilast := t.Raw[0].Line, t.Raw[n-1].Line
		lfirst, llast := parentRegion[0].Line, parentRegion[m-1].Line
		if n == m {
			opt.topPackedForP = true
		} else if ifirst == lfirst && !isBlank(t.Raw[n-1]) {
			opt.topPackedForP = true
		} else if ifirst > lfirst && !isBlank(parentRegion[ifirst-lfirst-1]) {
			opt.topPackedForP = true
		}
		// bottom-packed?
		if n == m {
			opt.bottomPackedForP = true
		} else if ilast == llast && !isBlank(parentRegion[ifirst-lfirst-1]) {
			opt.bottomPackedForP = true
		} else if ilast < llast && !isBlank(t.Raw[n-1]) {
			opt.bottomPackedForP = true
		}
		opt.itemEndForP = t.Raw[n-1].Line

		c.printf("<li>")
		c.blocks(c.tags[1:], opt)
		c.printf("</li>\n")
		if c.err != nil {
			return c.tags, c.err
		}
	}
}

func htmlBlocks(tags []md.Tag, w io.Writer, opt htmlOpt) ([]md.Tag, error) {
	var err error
	for i := 0; len(tags) > 0; i++ {
		if (tags[0] == md.End{}) {
			return tags[1:], nil
		}
		opt := opt
		if i != 0 {
			// top-packedness disables <p> only if 1st element
			opt.topPackedForP = false
		}
		if i == 1 {
			// bottom-packedness doesn't disable <p> for 2nd element
			opt.bottomPackedForP = false
		}
		tags, err = htmlBlock(tags, w, opt)
		if err != nil {
			return tags, err
		}
	}
	return tags, nil
}

var (
	tmplImage = template.Must(template.New("vfmd.<img>").Parse(
		`<img src="{{.URL}}"` +
			`{{if not (eq .alt "")}} alt="{{.alt}}"{{end}}` +
			`{{if not (eq .Title "")}} title="{{.Title}}"{{end}}` +
			` />`))
	tmplLink = template.Must(template.New("vfmd.<a href>").Parse(
		`<a href="{{.URL}}"` +
			`{{if not (eq .Title "")}} title="{{.Title}}"{{end}}` +
			`>`))
)

func htmlSpans(tags []md.Tag, w io.Writer, opt htmlOpt) ([]md.Tag, error) {
	c := htmlContext{w: w, tags: tags}
	for {
		oldtags := c.tags
		switch t := c.tags[0].(type) {
		case md.End:
			return c.tags[1:], nil

		case md.Prose:
			for _, r := range t {
				c.printf("%s", html.EscapeString(string(r.Bytes)))
			}
			c.tags = c.tags[1:]
		case md.Emphasis:
			c.printf("%s", map[int]string{
				1: "<em>",
				2: "<strong>",
				3: "<strong><em>",
			}[t.Level])
			c.spans(c.tags[1:], opt)
			c.printf("%s", map[int]string{
				1: "</em>",
				2: "</strong>",
				3: "</em></strong>",
			}[t.Level])
		case md.AutomaticLink:
			c.printf(`<a href="%s">%s</a>`,
				// FIXME(akavel): fully correct escaping
				t.URL, html.EscapeString(t.Text))
			c.tags = c.tags[1:]
		case md.Code:
			c.printf(`<code>%s</code>`,
				html.EscapeString(string(t.Code)))
			c.tags = c.tags[1:]
		case md.Link:
			ref := htmlLinkInfo{URL: t.URL, Title: t.Title}
			found := ref.URL != ""
			if !found {
				found = opt.fillRef(t.ReferenceID, &ref)
			}
			if found {
				// FIXME(akavel): fully correct escaping
				// FIXME(akavel): using URL below allows for "javascript:"; provide some way to protect against this (only whitelisted URL schemes?)
				if c.err == nil {
					c.err = tmplLink.Execute(w, map[string]interface{}{
						"Title": ref.Title,
						"URL":   template.URL(ref.URL),
					})
				}
			} else {
				c.printf(`[`)
			}
			c.spans(c.tags[1:], opt)
			if found {
				c.printf(`</a>`)
			} else {
				rawEnd := mdutils.DeEscapeProse(md.Prose(t.RawEnd))
				for _, r := range rawEnd {
					c.write(r.Bytes)
				}
			}
		case md.Image:
			ref := htmlLinkInfo{URL: t.URL, Title: t.Title}
			found := ref.URL != ""
			if !found {
				found = opt.fillRef(t.ReferenceID, &ref)
			}
			alt := string(t.AltText)
			if found {
				// FIXME(akavel): fully correct escaping
				// FIXME(akavel): using URL below allows for "javascript:"; provide some way to protect against this (only whitelisted URL schemes?)
				if c.err == nil {
					c.err = tmplImage.Execute(w, map[string]interface{}{
						"Title": ref.Title,
						"alt":   alt,
						"URL":   template.URL(ref.URL),
					})
				}
			} else {
				c.printf(`![%s`, alt)
				rawEnd := mdutils.DeEscapeProse(md.Prose(t.RawEnd))
				for _, r := range rawEnd {
					c.write(r.Bytes)
				}
			}
			c.tags = c.tags[1:]

		default:
			// TODO(akavel): return error's context (e.g. remaining tags?)
			return c.tags, fmt.Errorf("vfmd: span type %T not supported yet", t)
		}
		if c.err == nil {
			c.err = chkmoved(oldtags, c.tags)
		}
		if c.err != nil {
			return c.tags, c.err
		}
	}
}

var reSimplifyHtml = regexp.MustCompile(`>\s*<`)

// simplifyHtml performs a quick & dirty HTML unification in a similar way
// as the fallback approach in the "run_tests" script in testdata dir.
func simplifyHtml(buf []byte) []byte {
	buf = reSimplifyHtml.ReplaceAllLiteral(buf, []byte(">\n<"))
	buf = bytes.Replace(buf, []byte("<pre>\n<code>"), []byte("<pre><code>"), -1)
	buf = bytes.Replace(buf, []byte("</code>\n</pre>"), []byte("</code></pre>"), -1)
	buf = bytes.TrimSpace(buf)
	return buf
}
