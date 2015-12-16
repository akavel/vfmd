package vfmd

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/davecgh/go-spew/spew"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/utils"
)

func QuickHTML(blocks []md.Tag) []byte {
	buf := bytes.NewBuffer(nil)
	opt := htmlOpt{
		refs: htmlRefs(blocks),
	}
	var err error
	tags := blocks
	for len(tags) > 0 {
		tags, err = htmlBlock(tags, buf, opt)
		if err != nil {
			i := len(blocks) - len(tags)
			fmt.Fprintf(os.Stderr, "%s\n%s\n%s\n",
				spew.Sdump(blocks[:i]), err, spew.Sdump(blocks[i:]))
			panic(err)
		}
	}
	return buf.Bytes()
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

func htmlBlock(tags []md.Tag, w io.Writer, opt htmlOpt) ([]md.Tag, error) {
	var err error
	switch t := tags[0].(type) {
	case md.AtxHeaderBlock:
		fmt.Fprintf(w, "<h%d>", t.Level)
		tags, err = htmlSpans(tags[1:], w, opt)
		fmt.Fprintf(w, "</h%d>\n", t.Level)
		return tags, err
	case md.SetextHeaderBlock:
		fmt.Fprintf(w, "<h%d>", t.Level)
		tags, err = htmlSpans(tags[1:], w, opt)
		fmt.Fprintf(w, "</h%d>\n", t.Level)
		return tags, err
	case md.NullBlock:
		fmt.Fprintln(w)
		return tags[2:], nil
	case md.QuoteBlock:
		fmt.Fprintf(w, "<blockquote>\n  ")
		tags, err = htmlBlocks(tags[1:], w, htmlOpt{refs: opt.refs})
		fmt.Fprintf(w, "</blockquote>\n")
		return tags, err
	case md.ParagraphBlock:
		n := len(t.Raw)
		no_p := opt.topPackedForP ||
			(opt.bottomPackedForP && t.Raw[n-1].Line == opt.itemEndForP)
		if !no_p {
			fmt.Fprintf(w, "<p>")
		}
		tags, err = htmlSpans(tags[1:], w, opt)
		if !no_p {
			fmt.Fprintf(w, "</p>\n")
		}
		return tags, err
	case md.CodeBlock:
		fmt.Fprintf(w, "<pre><code>")
		for _, r := range t.Prose {
			fmt.Fprint(w, html.EscapeString(string(r.Bytes)))
		}
		fmt.Fprintf(w, "</code></pre>\n")
		return tags[2:], nil
	case md.HorizontalRuleBlock:
		fmt.Fprintf(w, "<hr />\n")
		return tags[2:], nil
	case md.OrderedListBlock:
		var i int
		fmt.Sscanf(string(t.Starter.Bytes), "%d", &i)
		if i != 1 {
			fmt.Fprintf(w, "<ol start=\"%d\">\n", i)
		} else {
			fmt.Fprintf(w, "<ol>\n")
		}
		tags, err = htmlItems(tags[1:], w, t.Raw, opt)
		fmt.Fprintf(w, "</ol>\n")
		return tags, err
	case md.UnorderedListBlock:
		fmt.Fprintf(w, "<ul>\n")
		tags, err = htmlItems(tags[1:], w, t.Raw, opt)
		fmt.Fprintf(w, "</ul>\n")
		return tags, err
	case md.ReferenceResolutionBlock:
		return tags[2:], nil
	default:
		return tags, fmt.Errorf("block type %T not supported yet", t)
	}
}

func isBlank(line md.Run) bool {
	return len(bytes.Trim(line.Bytes, " \t\n")) == 0
}

func htmlItems(tags []md.Tag, w io.Writer, parentRegion md.Raw, opt htmlOpt) ([]md.Tag, error) {
	var err error
	for {
		if (tags[0] == md.End{}) {
			return tags[1:], nil
		}

		t := tags[0].(md.ItemBlock)
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

		fmt.Fprintf(w, "<li>")
		tags, err = htmlBlocks(tags[1:], w, opt)
		fmt.Fprintf(w, "</li>\n")
		if err != nil {
			return tags, err
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
	tmplImage = template.Must(template.New("").Parse(
		`<img src="{{.URL}}"` +
			`{{if not (eq .alt "")}} alt="{{.alt}}"{{end}}` +
			`{{if not (eq .Title "")}} title="{{.Title}}"{{end}}` +
			` />`))
	tmplLink = template.Must(template.New("").Parse(
		`<a href="{{.URL}}"` +
			`{{if not (eq .Title "")}} title="{{.Title}}"{{end}}` +
			`>`))
)

func htmlSpans(tags []md.Tag, w io.Writer, opt htmlOpt) ([]md.Tag, error) {
	var err error
	for {
		switch t := tags[0].(type) {
		case md.Prose:
			for _, r := range t {
				fmt.Fprint(w, html.EscapeString(string(r.Bytes)))
			}
			tags = tags[1:]
		case md.Emphasis:
			fmt.Fprint(w, map[int]string{
				1: "<em>",
				2: "<strong>",
				3: "<strong><em>",
			}[t.Level])
			tags, err = htmlSpans(tags[1:], w, opt)
			fmt.Fprint(w, map[int]string{
				1: "</em>",
				2: "</strong>",
				3: "</em></strong>",
			}[t.Level])
		case md.AutomaticLink:
			fmt.Fprintf(w, `<a href="%s">%s</a>`,
				// FIXME(akavel): fully correct escaping
				t.URL, html.EscapeString(t.Text))
			tags = tags[1:]
		case md.Code:
			fmt.Fprintf(w, `<code>%s</code>`,
				html.EscapeString(string(t.Code)))
			tags = tags[1:]
		case md.Link:
			ref := htmlLinkInfo{URL: t.URL, Title: t.Title}
			found := ref.URL != ""
			if !found {
				found = opt.fillRef(t.ReferenceID, &ref)
			}
			if found {
				// FIXME(akavel): fully correct escaping
				// FIXME(akavel): do something nice with err
				// FIXME(akavel): using URL below allows for "javascript:"; provide some way to protect against this (only whitelisted URL schemes?)
				err := tmplLink.Execute(w, map[string]interface{}{
					"Title": ref.Title,
					"URL":   template.URL(ref.URL),
				})
				if err != nil {
					panic(err)
				}
			} else {
				fmt.Fprintf(w, `[`)
			}
			tags, err = htmlSpans(tags[1:], w, opt)
			if found {
				fmt.Fprintf(w, `</a>`)
			} else {
				rawEnd := utils.DeEscapeProse(md.Prose(t.RawEnd))
				for _, r := range rawEnd {
					w.Write(r.Bytes)
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
				// FIXME(akavel): do something nice with err
				// FIXME(akavel): using URL below allows for "javascript:"; provide some way to protect against this (only whitelisted URL schemes?)
				err := tmplImage.Execute(w, map[string]interface{}{
					"Title": ref.Title,
					"alt":   alt,
					"URL":   template.URL(ref.URL),
				})
				if err != nil {
					panic(err)
				}
			} else {
				fmt.Fprintf(w, `![%s`, alt)
				rawEnd := utils.DeEscapeProse(md.Prose(t.RawEnd))
				for _, r := range rawEnd {
					w.Write(r.Bytes)
				}
			}
			tags = tags[1:]

		case md.End:
			return tags[1:], nil
		default:
			return tags, fmt.Errorf("span type %T not supported yet", t)
		}
		if err != nil {
			return tags, err
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
