package mdhtml

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"strings"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/mdutils"
)

func QuickRender(w io.Writer, blocks []md.Tag) error {
	opt := Opt{
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

type Spaner interface {
	Span(Context, Opt) ([]md.Tag, error)
}
type Blocker interface {
	Block(Context, Opt) ([]md.Tag, error)
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
type Opt struct {
	refs                            map[string]htmlLinkInfo
	topPackedForP, bottomPackedForP bool
	itemEndForP                     int
}

func (opt Opt) fillRef(refID string, ref *htmlLinkInfo) bool {
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

type Context struct {
	W    io.Writer
	Tags []md.Tag
	Err  error
}

func (c *Context) Printf(format string, args ...interface{}) {
	if c.Err != nil {
		return
	}
	_, c.Err = fmt.Fprintf(c.W, format, args...)
}
func (c *Context) Spans(tags []md.Tag, opt Opt) {
	if c.Err != nil {
		return
	}
	c.Tags, c.Err = htmlSpans(tags, c.W, opt)
	if c.Err == nil {
		c.Err = chkmoved(tags, c.Tags)
	}
}
func (c *Context) Blocks(tags []md.Tag, opt Opt) {
	if c.Err != nil {
		return
	}
	c.Tags, c.Err = htmlBlocks(tags, c.W, opt)
	if c.Err == nil {
		c.Err = chkmoved(tags, c.Tags)
	}
}
func (c *Context) items(tags []md.Tag, parentRegion md.Raw, opt Opt) {
	if c.Err != nil {
		return
	}
	c.Tags, c.Err = htmlItems(tags, c.W, parentRegion, opt)
	if c.Err == nil {
		c.Err = chkmoved(tags, c.Tags)
	}
}
func (c *Context) write(buf []byte) {
	if c.Err != nil {
		return
	}
	_, c.Err = c.W.Write(buf)
}

func htmlBlock(tags []md.Tag, w io.Writer, opt Opt) ([]md.Tag, error) {
	c := Context{W: w, Tags: tags}
	switch t := tags[0].(type) {
	case md.AtxHeaderBlock:
		c.Printf("<h%d>", t.Level)
		c.Spans(tags[1:], opt)
		c.Printf("</h%d>\n", t.Level)
		return c.Tags, c.Err
	case md.SetextHeaderBlock:
		c.Printf("<h%d>", t.Level)
		c.Spans(tags[1:], opt)
		c.Printf("</h%d>\n", t.Level)
		return c.Tags, c.Err
	case md.NullBlock:
		// TODO(akavel): don't print the empty line?
		c.Printf("\n")
		return c.Tags[2:], c.Err
	case md.QuoteBlock:
		c.Printf("<blockquote>\n  ")
		c.Blocks(tags[1:], Opt{refs: opt.refs})
		c.Printf("</blockquote>\n")
		return c.Tags, c.Err
	case md.ParagraphBlock:
		n := len(t.Raw)
		no_p := opt.topPackedForP ||
			(opt.bottomPackedForP && t.Raw[n-1].Line == opt.itemEndForP)
		if !no_p {
			c.Printf("<p>")
		}
		c.Spans(tags[1:], opt)
		if !no_p {
			c.Printf("</p>\n")
		}
		return c.Tags, c.Err
	case md.CodeBlock:
		c.Printf("<pre><code>")
		for _, r := range t.Prose {
			c.Printf("%s", html.EscapeString(string(r.Bytes)))
		}
		c.Printf("</code></pre>\n")
		return c.Tags[2:], c.Err
	case md.HorizontalRuleBlock:
		c.Printf("<hr />\n")
		return c.Tags[2:], c.Err
	case md.OrderedListBlock:
		var i int
		fmt.Sscanf(string(t.Starter.Bytes), "%d", &i)
		if i != 1 {
			c.Printf("<ol start=\"%d\">\n", i)
		} else {
			c.Printf("<ol>\n")
		}
		c.items(tags[1:], t.Raw, opt)
		c.Printf("</ol>\n")
		return c.Tags, c.Err
	case md.UnorderedListBlock:
		c.Printf("<ul>\n")
		c.items(tags[1:], t.Raw, opt)
		c.Printf("</ul>\n")
		return c.Tags, c.Err
	case md.ReferenceResolutionBlock:
		return c.Tags[2:], nil
	default:
		b, ok := t.(Blocker)
		if !ok {
			// TODO(akavel): return error's context (e.g. remaining tags?)
			return tags, fmt.Errorf("vfmd: block type %T not supported yet", t)
		}
		return b.Block(c, opt)
	}
}

func isBlank(line md.Run) bool {
	return len(bytes.Trim(line.Bytes, " \t\n")) == 0
}

func htmlItems(tags []md.Tag, w io.Writer, parentRegion md.Raw, opt Opt) ([]md.Tag, error) {
	c := Context{W: w, Tags: tags}
	for {
		if (c.Tags[0] == md.End{}) {
			return c.Tags[1:], nil
		}

		t := c.Tags[0].(md.ItemBlock)
		opt := Opt{refs: opt.refs}
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

		c.Printf("<li>")
		c.Blocks(c.Tags[1:], opt)
		c.Printf("</li>\n")
		if c.Err != nil {
			return c.Tags, c.Err
		}
	}
}

func htmlBlocks(tags []md.Tag, w io.Writer, opt Opt) ([]md.Tag, error) {
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

func htmlSpans(tags []md.Tag, w io.Writer, opt Opt) ([]md.Tag, error) {
	c := Context{W: w, Tags: tags}
	for {
		oldtags := c.Tags
		switch t := c.Tags[0].(type) {
		case md.End:
			return c.Tags[1:], nil

		case md.Prose:
			for _, r := range t {
				c.Printf("%s", html.EscapeString(string(r.Bytes)))
			}
			c.Tags = c.Tags[1:]
		case md.Emphasis:
			c.Printf("%s", map[int]string{
				1: "<em>",
				2: "<strong>",
				3: "<strong><em>",
			}[t.Level])
			c.Spans(c.Tags[1:], opt)
			c.Printf("%s", map[int]string{
				1: "</em>",
				2: "</strong>",
				3: "</em></strong>",
			}[t.Level])
		case md.AutomaticLink:
			c.Printf(`<a href="%s">%s</a>`,
				// FIXME(akavel): fully correct escaping
				t.URL, html.EscapeString(t.Text))
			c.Tags = c.Tags[2:]
		case md.Code:
			c.Printf(`<code>%s</code>`,
				html.EscapeString(string(t.Code)))
			c.Tags = c.Tags[2:]
		case md.Link:
			ref := htmlLinkInfo{URL: t.URL, Title: t.Title}
			found := ref.URL != ""
			if !found {
				found = opt.fillRef(t.ReferenceID, &ref)
			}
			if found {
				// FIXME(akavel): fully correct escaping
				// FIXME(akavel): using URL below allows for "javascript:"; provide some way to protect against this (only whitelisted URL schemes?)
				if c.Err == nil {
					c.Err = tmplLink.Execute(w, map[string]interface{}{
						"Title": ref.Title,
						"URL":   template.URL(ref.URL),
					})
				}
			} else {
				c.Printf(`[`)
			}
			c.Spans(c.Tags[1:], opt)
			if found {
				c.Printf(`</a>`)
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
				if c.Err == nil {
					c.Err = tmplImage.Execute(w, map[string]interface{}{
						"Title": ref.Title,
						"alt":   alt,
						"URL":   template.URL(ref.URL),
					})
				}
			} else {
				c.Printf(`![%s`, alt)
				rawEnd := mdutils.DeEscapeProse(md.Prose(t.RawEnd))
				for _, r := range rawEnd {
					c.write(r.Bytes)
				}
			}
			c.Tags = c.Tags[2:]

		default:
			s, ok := t.(Spaner)
			if !ok {
				// TODO(akavel): return error's context (e.g. remaining tags?)
				return c.Tags, fmt.Errorf("vfmd: span type %T not supported, missing Span method", t)
			}
			c.Tags, c.Err = s.Span(c, opt)
		}
		if c.Err == nil {
			c.Err = chkmoved(oldtags, c.Tags)
		}
		if c.Err != nil {
			return c.Tags, c.Err
		}
	}
}
