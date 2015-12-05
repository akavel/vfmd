package span

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kylelemons/godebug/diff"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/utils"
)

func bb(s string) []byte { return []byte(s) }

type spanCase struct {
	fname  string
	buf    []byte
	blocks [][]byte
	spans  []Span
}

type spans []Span

const dir = `../testdata/tests/span_level/`

type opt func([]byte) []byte

func head(nlines int) opt {
	return func(buf []byte) []byte {
		lines := bytes.Split(buf, []byte("\n"))[:nlines]
		return bytes.Join(lines, []byte("\n"))
	}
}

func lines(filename string, spans spans, opts ...opt) spanCase {
	buf, err := ioutil.ReadFile(dir + filename)
	if err != nil {
		panic(err)
	}
	for _, opt := range opts {
		buf = opt(buf)
	}
	return spanCase{
		fname: filename,
		buf:   buf,
		blocks: bytes.FieldsFunc(buf, func(r rune) bool {
			return r == '\n' || r == '\r'
		}),
		spans: spans,
	}
}
func blocks(filename string, spans spans, opts ...opt) spanCase {
	buf, err := ioutil.ReadFile(dir + filename)
	if err != nil {
		panic(err)
	}
	buf = bytes.Replace(buf, bb("\r"), bb(""), -1)
	for _, opt := range opts {
		buf = opt(buf)
	}
	return spanCase{
		fname:  filename,
		buf:    buf,
		blocks: bytes.Split(buf, []byte("\n\n")),
		spans:  spans,
	}
}

func emB(tag string) Span { return Span{bb(tag), md.Emphasis{len(tag)}} }
func emE(tag string) Span { return Span{bb(tag), md.End{}} }

func TestSpan(test *testing.T) {
	cases := []spanCase{
		lines(`automatic_links/angle_brackets_in_link.md`, spans{
			{bb("http://exampl"), md.AutomaticLink{URL: `http://exampl`, Text: `http://exampl`}},
			// TODO(akavel): below is expected by testdata/, but
			// invalid according to spec, because preceding "<" is
			// not a 'word-separator' character (it has unicode
			// general class Sm - "Symbol, math"); try to resolve
			// this with the spec author.
			// {bb("http://exampl"), md.AutomaticLink{URL: `http://exampl`, Text: `http://exampl`}},
		}),
		lines("automatic_links/ending_with_punctuation.md", spans{
			{bb("http://example.net"), md.AutomaticLink{URL: "http://example.net", Text: "http://example.net"}},
			{bb("http://example.net/"), md.AutomaticLink{URL: "http://example.net/", Text: "http://example.net/"}},
			{bb("http://example.net"), md.AutomaticLink{URL: "http://example.net", Text: "http://example.net"}},
			{bb("http://example.net/"), md.AutomaticLink{URL: "http://example.net/", Text: "http://example.net/"}},

			{bb("<http://example.net,>"), md.AutomaticLink{URL: "http://example.net,", Text: "http://example.net,"}},
			{bb("<http://example.net/,>"), md.AutomaticLink{URL: "http://example.net/,", Text: "http://example.net/,"}},
			{bb("<http://example.net)>"), md.AutomaticLink{URL: "http://example.net)", Text: "http://example.net)"}},
			{bb("<http://example.net/)>"), md.AutomaticLink{URL: "http://example.net/)", Text: "http://example.net/)"}},
		}),
		lines("automatic_links/mail_url_in_angle_brackets.md", spans{
			{bb("<mailto:someone@example.net>"), md.AutomaticLink{URL: "mailto:someone@example.net", Text: "mailto:someone@example.net"}},
			{bb("<someone@example.net>"), md.AutomaticLink{URL: "mailto:someone@example.net", Text: "someone@example.net"}},
		}),
		lines("automatic_links/mail_url_without_angle_brackets.md", spans{
			// NOTE(akavel): below line is unexpected according to
			// testdata/, but from spec this seems totally expected,
			// so I added it
			{bb("mailto:someone@example.net"), md.AutomaticLink{URL: "mailto:someone@example.net", Text: "mailto:someone@example.net"}},
		}),
		lines("automatic_links/url_schemes.md", spans{
			{bb("http://example.net"), md.AutomaticLink{URL: "http://example.net", Text: "http://example.net"}},
			{bb("<http://example.net>"), md.AutomaticLink{URL: "http://example.net", Text: "http://example.net"}},
			{bb("file:///tmp/tmp.html"), md.AutomaticLink{URL: "file:///tmp/tmp.html", Text: "file:///tmp/tmp.html"}},
			{bb("<file:///tmp/tmp.html>"), md.AutomaticLink{URL: "file:///tmp/tmp.html", Text: "file:///tmp/tmp.html"}},
			{bb("feed://example.net/rss.xml"), md.AutomaticLink{URL: "feed://example.net/rss.xml", Text: "feed://example.net/rss.xml"}},
			{bb("<feed://example.net/rss.xml>"), md.AutomaticLink{URL: "feed://example.net/rss.xml", Text: "feed://example.net/rss.xml"}},
			{bb("googlechrome://example.net/"), md.AutomaticLink{URL: "googlechrome://example.net/", Text: "googlechrome://example.net/"}},
			{bb("<googlechrome://example.net/>"), md.AutomaticLink{URL: "googlechrome://example.net/", Text: "googlechrome://example.net/"}},
			{bb("`<>`"), md.Code{bb("<>")}},
			// NOTE(akavel): below line is unexpected according to
			// testdata/, but from spec this seems totally expected,
			// so I added it
			{bb("mailto:me@example.net"), md.AutomaticLink{URL: "mailto:me@example.net", Text: "mailto:me@example.net"}},
			{bb("<mailto:me@example.net>"), md.AutomaticLink{URL: "mailto:me@example.net", Text: "mailto:me@example.net"}},
		}),
		lines("automatic_links/url_special_chars.md", spans{
			{bb(`http://example.net/*#$%^&\~/blah`), md.AutomaticLink{URL: `http://example.net/*#$%^&\~/blah`, Text: `http://example.net/*#$%^&\~/blah`}},
			{bb(`<http://example.net/*#$%^&\~)/blah>`), md.AutomaticLink{URL: `http://example.net/*#$%^&\~)/blah`, Text: `http://example.net/*#$%^&\~)/blah`}},
			// NOTE(akavel): testdata expects below commented entry,
			// but this seems wrong compared to spec; I've added
			// fixed entry
			// {bb(`http://example.net/blah/`), md.AutomaticLink{URL: `http://example.net/blah/`, Text: `http://example.net/blah/`}},
			{bb(`http://example.net/blah/*#$%^&\~`), md.AutomaticLink{URL: `http://example.net/blah/*#$%^&\~`, Text: `http://example.net/blah/*#$%^&\~`}},
			{bb(`<http://example.net/blah/*#$%^&\~)>`), md.AutomaticLink{URL: `http://example.net/blah/*#$%^&\~)`, Text: `http://example.net/blah/*#$%^&\~)`}},
		}),
		lines("automatic_links/web_url_in_angle_brackets.md", spans{
			{bb("<http://example.net/path/>"), md.AutomaticLink{URL: "http://example.net/path/", Text: "http://example.net/path/"}},
			{bb("<https://example.net/path/>"), md.AutomaticLink{URL: "https://example.net/path/", Text: "https://example.net/path/"}},
			{bb("<ftp://example.net/path/>"), md.AutomaticLink{URL: "ftp://example.net/path/", Text: "ftp://example.net/path/"}},
		}),
		lines("automatic_links/web_url_without_angle_brackets.md", spans{
			{bb("http://example.net/path/"), md.AutomaticLink{URL: "http://example.net/path/", Text: "http://example.net/path/"}},
			{bb("https://example.net/path/"), md.AutomaticLink{URL: "https://example.net/path/", Text: "https://example.net/path/"}},
			{bb("ftp://example.net/path/"), md.AutomaticLink{URL: "ftp://example.net/path/", Text: "ftp://example.net/path/"}},
		}),
		lines("code/end_of_codespan.md", spans{
			{bb("`code span`"), md.Code{bb("code span")}},
			{bb("``code span` ends``"), md.Code{bb("code span` ends")}},
			{bb("`code span`` ends`"), md.Code{bb("code span`` ends")}},
			{bb("````code span`` ``ends````"), md.Code{bb("code span`` ``ends")}},
			{bb("`code span\\`"), md.Code{bb(`code span\`)}},
		}),
		blocks("code/multiline.md", spans{
			{bb("`code span\ncan span multiple\nlines`"), md.Code{bb("code span\ncan span multiple\nlines")}},
		}),
		lines("code/vs_emph.md", spans{
			{bb("`code containing *em* text`"), md.Code{bb("code containing *em* text")}},
			{bb("`code containing **strong** text`"), md.Code{bb("code containing **strong** text")}},
			{bb("`code containing _em_ text`"), md.Code{bb("code containing _em_ text")}},
			{bb("`code containing __strong__ text`"), md.Code{bb("code containing __strong__ text")}},

			{bb("*"), md.Emphasis{Level: 1}},
			{bb("`code`"), md.Code{bb("code")}},
			{bb("*"), md.End{}},
			{bb("**"), md.Emphasis{Level: 2}},
			{bb("`code`"), md.Code{bb("code")}},
			{bb("**"), md.End{}},
			{bb("_"), md.Emphasis{Level: 1}},
			{bb("`code`"), md.Code{bb("code")}},
			{bb("_"), md.End{}},
			{bb("__"), md.Emphasis{Level: 2}},
			{bb("`code`"), md.Code{bb("code")}},
			{bb("__"), md.End{}},

			{bb("`code *intertwined`"), md.Code{bb("code *intertwined")}},
			{bb("`with em* text`"), md.Code{bb("with em* text")}},
			{bb("`code **intertwined`"), md.Code{bb("code **intertwined")}},
			{bb("`with strong** text`"), md.Code{bb("with strong** text")}},
			{bb("`code _intertwined`"), md.Code{bb("code _intertwined")}},
			{bb("`with em_ text`"), md.Code{bb("with em_ text")}},
			{bb("`code __intertwined`"), md.Code{bb("code __intertwined")}},
			{bb("`with strong__ text`"), md.Code{bb("with strong__ text")}},
		}),
		lines("code/vs_image.md", spans{
			{bb("`code containing ![image](url)`"), md.Code{bb("code containing ![image](url)")}},
			{bb("`code containing ![image][ref]`"), md.Code{bb("code containing ![image][ref]")}},
			{bb("`code containing ![ref]`"), md.Code{bb("code containing ![ref]")}},

			{bb("`containing code`"), md.Code{bb("containing code")}},
			{bb("`containing code`"), md.Code{bb("containing code")}},
			{bb("["), md.Link{ReferenceID: "ref"}},
			{bb("]"), md.End{}},
			{bb("`containing code`"), md.Code{bb("containing code")}},

			{bb("`code ![intertwined`"), md.Code{bb("code ![intertwined")}},
			{bb("`intertwined](with) image`"), md.Code{bb("intertwined](with) image")}},
			{bb("`code ![intertwined`"), md.Code{bb("code ![intertwined")}},
			{bb("["), md.Link{ReferenceID: "ref"}},
			{bb("]"), md.End{}},
			{bb("`intertwined with][ref] image`"), md.Code{bb("intertwined with][ref] image")}},
			{bb("`code ![intertwined`"), md.Code{bb("code ![intertwined")}},
			{bb("`with] image`"), md.Code{bb("with] image")}},
		}, head(30)),
		lines("code/vs_link.md", spans{
			{bb("`code containing [link](url)`"), md.Code{bb("code containing [link](url)")}},
			{bb("`code containing [link][ref]`"), md.Code{bb("code containing [link][ref]")}},
			{bb("`code containing [ref]`"), md.Code{bb("code containing [ref]")}},

			{bb("["), md.Link{URL: "url"}},
			{bb("`containing code`"), md.Code{bb("containing code")}},

			{bb("](url)"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref"}},
			{bb("`containing code`"), md.Code{bb("containing code")}},
			{bb("][ref]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "link `containing code`"}},
			{bb("`containing code`"), md.Code{bb("containing code")}},
			{bb("]"), md.End{}},

			{bb("`code [intertwined`"), md.Code{bb("code [intertwined")}},
			{bb("`intertwined](with) link`"), md.Code{bb("intertwined](with) link")}},
			{bb("`code [intertwined`"), md.Code{bb("code [intertwined")}},
			{bb("["), md.Link{ReferenceID: "ref"}},
			{bb("]"), md.End{}},
			{bb("`intertwined with][ref] link`"), md.Code{bb("intertwined with][ref] link")}},
			{bb("`code [intertwined`"), md.Code{bb("code [intertwined")}},
			{bb("`with] link`"), md.Code{bb("with] link")}},
		}, head(30)),
		lines("code/well_formed.md", spans{
			{bb("`code span`"), md.Code{bb("code span")}},
			{bb("``code ` span``"), md.Code{bb("code ` span")}},
			{bb("`` ` code span``"), md.Code{bb("` code span")}},
			{bb("``code span ` ``"), md.Code{bb("code span `")}},
			{bb("`` `code span` ``"), md.Code{bb("`code span`")}},
		}),
		lines("emphasis/emphasis_tag_combinations.md", spans{
			emB("*"), emB("__"), emE("__"), emE("*"),
			emB("_"), emB("**"), emE("**"), emE("_"),
			emB("***"), emE("***"),
			emB("**"), emB("*"), emE("*"), emE("**"),
			emB("*"), emB("*"), emB("*"), emE("*"), emE("*"), emE("*"),
			emB("*"), emB("**"), emE("**"), emE("*"),

			emB("_"), emB("__"), emE("__"), emE("_"),
			emB("_"), emB("_"), emB("_"), emE("_"), emE("_"), emE("_"),
			emB("__"), emB("_"), emE("_"), emE("__"),
		}),
		lines("emphasis/intertwined.md", spans{
			emB("*"), emE("*"),
			emB("**"), emE("**"),
			emB("*"), emB("*"), emB("*"), emE("*"), emE("*"), emE("*"),
			emB("*"), emB("*"), emB("*"), emE("*"), emE("*"), emE("*"),

			emB("_"), emE("_"),
			emB("__"), emE("__"),
			emB("_"), emB("_"), emB("_"), emE("_"), emE("_"), emE("_"),
			emB("_"), emB("_"), emB("_"), emE("_"), emE("_"), emE("_"),
		}),
		lines("emphasis/intraword.md", spans{}),
		lines("emphasis/nested_homogenous.md", spans{
			emB("*"), emB("*"), emE("*"), emE("*"),
			emB("**"), emB("**"), emE("**"), emE("**"),
			emB("_"), emB("_"), emE("_"), emE("_"),
			emB("__"), emB("__"), emE("__"), emE("__"),
		}),
		lines("emphasis/opening_and_closing_tags.md", spans{}),
		lines("emphasis/simple.md", spans{
			emB("*"), emE("*"), emB("**"), emE("**"),
			emB("_"), emE("_"), emB("__"), emE("__"),
		}),
		lines("emphasis/within_whitespace.md", spans{}),
		lines("emphasis/with_punctuation.md", spans{
			emB("*"), emE("*"),
			emB("*"), emE("*"),
			emB("*"), emE("*"),
			emB("*"), emE("*"),

			emB("_"), emE("_"), emB("_"), emE("_"),
			// NOTE(akavel): link below not expected in testdata
			// because it's not defined below; but we leave this to
			// user.
			{bb("["), md.Link{ReferenceID: "_"}},
			{bb("]"), md.End{}},
			emB("_"), emE("_"), emB("_"), emE("_"),
			// NOTE(akavel): link below not expected in testdata
			// because it's not defined below; but we leave this to
			// user.
			{bb("["), md.Link{ReferenceID: "_"}},
			{bb("]"), md.End{}},
		}),
		lines("image/direct_link.md", spans{
			{bb("![image](url)"), md.Image{AltText: bb("image"), URL: "url"}},
			{bb(`![image](url "title")`), md.Image{AltText: bb("image"), URL: "url", Title: "title"}},
		}),
		lines("image/direct_link_with_2separating_spaces.md", spans{
			{bb("![linking]  (/img.png)"), md.Image{AltText: bb("linking"), URL: "/img.png"}},
		}),
		blocks("image/direct_link_with_separating_newline.md", spans{
			{bb("![link]\n(/img.png)"), md.Image{AltText: bb("link"), URL: "/img.png"}},
		}),
		lines("image/direct_link_with_separating_space.md", spans{
			{bb("![link] (http://example.net/img.png)"), md.Image{AltText: bb("link"), URL: "http://example.net/img.png"}},
		}),
		lines("image/image_title.md", spans{
			{bb(`![link](url "title")`), md.Image{AltText: bb("link"), URL: "url", Title: `title`}},
			{bb(`![link](url 'title')`), md.Image{AltText: bb("link"), URL: "url", Title: `title`}},
			// TODO(akavel): unquote contents of Title when
			// processing? doesn't seem noted in spec, send fix for
			// spec?
			{bb(`![link](url "title 'with' \"quotes\"")`), md.Image{AltText: bb("link"), URL: "url", Title: `title 'with' \"quotes\"`}},
			{bb(`![link](url 'title \'with\' "quotes"')`), md.Image{AltText: bb("link"), URL: "url", Title: `title \'with\' "quotes"`}},
			{bb(`![link](url "title with (brackets)")`), md.Image{AltText: bb("link"), URL: "url", Title: `title with (brackets)`}},
			{bb(`![link](url 'title with (brackets)')`), md.Image{AltText: bb("link"), URL: "url", Title: `title with (brackets)`}},

			{bb("![ref id1]"), md.Image{ReferenceID: "ref id1", AltText: bb("ref id1")}},
			{bb("![ref id2]"), md.Image{ReferenceID: "ref id2", AltText: bb("ref id2")}},
			{bb("![ref id3]"), md.Image{ReferenceID: "ref id3", AltText: bb("ref id3")}},
			{bb("![ref id4]"), md.Image{ReferenceID: "ref id4", AltText: bb("ref id4")}},
			{bb("![ref id5]"), md.Image{ReferenceID: "ref id5", AltText: bb("ref id5")}},
			{bb("![ref id6]"), md.Image{ReferenceID: "ref id6", AltText: bb("ref id6")}},
			{bb("![ref id7]"), md.Image{ReferenceID: "ref id7", AltText: bb("ref id7")}},
			{bb("![ref id8]"), md.Image{ReferenceID: "ref id8", AltText: bb("ref id8")}},
			{bb("![ref id9]"), md.Image{ReferenceID: "ref id9", AltText: bb("ref id9")}},
			{bb("![ref id10]"), md.Image{ReferenceID: "ref id10", AltText: bb("ref id10")}},
			{bb("![ref id11]"), md.Image{ReferenceID: "ref id11", AltText: bb("ref id11")}},
			{bb("![ref id12]"), md.Image{ReferenceID: "ref id12", AltText: bb("ref id12")}},
		}, head(19)),
		lines("image/incomplete.md", spans{
			{bb("![ref undefined]"), md.Image{ReferenceID: "ref undefined", AltText: bb("ref undefined")}},
			{bb("![ref 1]"), md.Image{ReferenceID: "ref 1", AltText: bb("ref 1")}},
			{bb("![ref undefined]"), md.Image{ReferenceID: "ref undefined", AltText: bb("ref undefined")}},
			{bb("![ref 1]"), md.Image{ReferenceID: "ref 1", AltText: bb("ref 1")}},
		}, head(8)),
		blocks("image/link_text_with_newline.md", spans{
			{bb("![link\ntext](url1)"), md.Image{AltText: bb("link\ntext"), URL: "url1"}},
			{bb("![ref\nid][]"), md.Image{ReferenceID: "ref id", AltText: bb("ref\nid")}},
			{bb("![ref\nid]"), md.Image{ReferenceID: "ref id", AltText: bb("ref\nid")}},
		}, head(9)),
		lines("image/link_with_parenthesis.md", spans{
			{bb("![bad link](url)"), md.Image{URL: "url", AltText: bb("bad link")}},
			{bb(`![bad link](url\)`), md.Image{URL: `url\`, AltText: bb("bad link")}},
			{bb(`![link](<url)> "title")`), md.Image{URL: `url)`, AltText: bb("link"), Title: "title"}},
		}),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// lines("image/multiple_ref_id_definitions.md", spans{}),
		lines("image/nested_images.md", spans{
			{bb("![link2](url2)"), md.Image{AltText: bb("link2"), URL: "url2"}},
		}),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// lines("image/ref_case_sensitivity.md", spans{}),
		blocks("image/ref_id_matching.md", spans{
			{bb("![link][ref id]"), md.Image{ReferenceID: "ref id", AltText: bb("link")}},
			{bb("![link][ref   id]"), md.Image{ReferenceID: "ref id", AltText: bb("link")}},
			{bb("![link][  ref id  ]"), md.Image{ReferenceID: "ref id", AltText: bb("link")}},
			{bb("![link][ref\n   id]"), md.Image{ReferenceID: "ref id", AltText: bb("link")}},
			{bb("![ref id][]"), md.Image{ReferenceID: "ref id", AltText: bb("ref id")}},
			{bb("![ref   id][]"), md.Image{ReferenceID: "ref id", AltText: bb("ref   id")}},
			{bb("![  ref id  ][]"), md.Image{ReferenceID: "ref id", AltText: bb("  ref id  ")}},
			{bb("![ref\n   id][]"), md.Image{ReferenceID: "ref id", AltText: bb("ref\n   id")}},
			{bb("![ref id]"), md.Image{ReferenceID: "ref id", AltText: bb("ref id")}},
			{bb("![ref   id]"), md.Image{ReferenceID: "ref id", AltText: bb("ref   id")}},
			{bb("![  ref id  ]"), md.Image{ReferenceID: "ref id", AltText: bb("  ref id  ")}},
			{bb("![ref\n   id]"), md.Image{ReferenceID: "ref id", AltText: bb("ref\n   id")}},
		}, head(18)),
		// NOTE(akavel): below tests are not really interesting for us
		// here now.
		// lines("image/ref_link.md", spans{}),
		// lines("image/ref_link_empty.md", spans{}),
		// lines("image/ref_link_self.md", spans{}),
		lines("image/ref_link_with_2separating_spaces.md", spans{
			{bb("![link]  [ref]"), md.Image{AltText: bb("link"), ReferenceID: "ref"}},
		}, head(2)),
		blocks("image/ref_link_with_separating_newline.md", spans{
			{bb("![link]\n[ref]"), md.Image{AltText: bb("link"), ReferenceID: "ref"}},
		}, head(3)),
		lines("image/ref_link_with_separating_space.md", spans{
			{bb("![link] [ref]"), md.Image{AltText: bb("link"), ReferenceID: "ref"}},
		}, head(2)),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// lines("image/ref_resolution_within_other_blocks.md", spans{}),
		lines("image/square_brackets_in_link_or_ref.md", spans{
			{bb("["), md.Link{ReferenceID: "1"}},
			{bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "2"}},
			{bb("]"), md.End{}},
			{bb("![2]"), md.Image{ReferenceID: "2", AltText: bb("2")}},
			// TODO(akavel): make sure we handled escaping properly in cases below
			{bb(`![link\[1\]](url)`), md.Image{URL: "url", AltText: bb(`link\[1\]`)}},
			{bb(`![link\[2\]](url)`), md.Image{URL: "url", AltText: bb(`link\[2\]`)}},
			{bb(`![link!\[2\]](url)`), md.Image{URL: "url", AltText: bb(`link!\[2\]`)}},
			{bb("["), md.Link{ReferenceID: "2"}},
			{bb("]"), md.End{}},
			{bb("![link]"), md.Image{ReferenceID: "link", AltText: bb("link")}},
			{bb("["), md.Link{ReferenceID: "3"}},
			{bb("]"), md.End{}},
			{bb("![link]"), md.Image{ReferenceID: "link", AltText: bb("link")}},
			{bb("["), md.Link{ReferenceID: "4"}},
			{bb("]"), md.End{}},
			// TODO(akavel): make sure we handled escaping properly in cases below
			{bb(`![link][ref\[3\]]`), md.Image{ReferenceID: `ref\[3\]`, AltText: bb(`link`)}},
			{bb(`![link][ref\[4\]]`), md.Image{ReferenceID: `ref\[4\]`, AltText: bb(`link`)}},
			{bb("![link]"), md.Image{ReferenceID: "link", AltText: bb("link")}},
			{bb("["), md.Link{ReferenceID: "5"}},
			{bb("]"), md.End{}},
			{bb("![link][ref]"), md.Image{ReferenceID: "ref", AltText: bb("link")}},
			{bb(`![link][ref\[5]`), md.Image{ReferenceID: `ref\[5`, AltText: bb(`link`)}},
			{bb(`![link][ref\]6]`), md.Image{ReferenceID: `ref\]6`, AltText: bb(`link`)}},
		}, head(16)),
		lines("image/two_consecutive_refs.md", spans{
			{bb("![one][two]"), md.Image{ReferenceID: "two", AltText: bb("one")}},
			{bb("["), md.Link{ReferenceID: "three"}},
			{bb("]"), md.End{}},
			{bb("![one][four]"), md.Image{ReferenceID: "four", AltText: bb("one")}},
			{bb("["), md.Link{ReferenceID: "three"}},
			{bb("]"), md.End{}},
		}, head(4)),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// image/unused_ref.md
		lines("image/url_escapes.md", spans{
			{bb(`![link](url\_\:\$\?)`), md.Image{AltText: bb("link"), URL: `url\_\:\$\?`}},
			{bb(`![link](http://g&ouml;&ouml;gle.com)`), md.Image{AltText: bb("link"), URL: `http://g&ouml;&ouml;gle.com`}},
		}, head(2)),
		lines("image/url_in_angle_brackets.md", spans{
			{bb(`![link](<url>)`), md.Image{AltText: bb("link"), URL: "url"}},
			{bb(`![link](<url(>)`), md.Image{AltText: bb("link"), URL: "url("}},
			{bb(`![link](<url)>)`), md.Image{AltText: bb("link"), URL: "url)"}},
			{bb(`![link](<url)> "title")`), md.Image{AltText: bb("link"), URL: "url)", Title: "title"}},
		}, head(4)),
		lines("image/url_special_chars.md", spans{
			{bb(`![link](url*#$%^&\~)`), md.Image{AltText: bb("link"), URL: `url*#$%^&\~`}},
			{bb("![link][ref id1]"), md.Image{ReferenceID: "ref id1", AltText: bb("link")}},
			{bb("![ref id1]"), md.Image{ReferenceID: "ref id1", AltText: bb("ref id1")}},
			{bb("![link]"), md.Image{ReferenceID: "link", AltText: bb("link")}},
		}, head(8)),
		blocks("image/url_whitespace.md", spans{
			{bb("![link]"), md.Image{AltText: bb("link"), ReferenceID: "link"}},
			{bb("![link]"), md.Image{AltText: bb("link"), ReferenceID: "link"}},
			{bb("![link](<url 1>)"), md.Image{AltText: bb("link"), URL: "url1"}},
			{bb("![link](<url \n   1>)"), md.Image{AltText: bb("link"), URL: "url1"}},
		}, head(6)),
		lines("image/vs_code.md", spans{
			{bb("`code`"), md.Code{bb("code")}},
			{bb("`containing ![image](url)`"), md.Code{bb("containing ![image](url)")}},
			{bb("`containing ![image][ref]`"), md.Code{bb("containing ![image][ref]")}},
			{bb("["), md.Link{ReferenceID: "ref"}},
			{bb("]"), md.End{}},
			{bb("`intertwined](url) with code`"), md.Code{bb("intertwined](url) with code")}},
			{bb("`intertwined ![with code`"), md.Code{bb("intertwined ![with code")}},
		}),
		lines("image/vs_emph.md", spans{
			{bb("![image containing *em* text](url)"), md.Image{URL: "url", AltText: bb("image containing *em* text")}},
			{bb("![image containing **strong** text](url)"), md.Image{URL: "url", AltText: bb("image containing **strong** text")}},
			{bb("![image containing _em_ text](url)"), md.Image{URL: "url", AltText: bb("image containing _em_ text")}},
			{bb("![image containing __strong__ text](url)"), md.Image{URL: "url", AltText: bb("image containing __strong__ text")}},

			emB("*"),
			{bb("![image](url)"), md.Image{AltText: bb("image"), URL: "url"}},
			emE("*"),
			emB("**"),
			{bb("![image](url)"), md.Image{AltText: bb("image"), URL: "url"}},
			emE("**"),
			emB("_"),
			{bb("![image](url)"), md.Image{AltText: bb("image"), URL: "url"}},
			emE("_"),
			emB("__"),
			{bb("![image](url)"), md.Image{AltText: bb("image"), URL: "url"}},
			emE("__"),

			{bb("![image *intertwined](url)"), md.Image{URL: "url", AltText: bb("image *intertwined")}},
			{bb("![with em* text](url)"), md.Image{URL: "url", AltText: bb("with em* text")}},
			{bb("![image **intertwined](url)"), md.Image{URL: "url", AltText: bb("image **intertwined")}},
			{bb("![with strong** text](url)"), md.Image{URL: "url", AltText: bb("with strong** text")}},
			{bb("![image _intertwined](url)"), md.Image{URL: "url", AltText: bb("image _intertwined")}},
			{bb("![with em_ text](url)"), md.Image{URL: "url", AltText: bb("with em_ text")}},
			{bb("![image __intertwined](url)"), md.Image{URL: "url", AltText: bb("image __intertwined")}},
			{bb("![with strong__ text](url)"), md.Image{URL: "url", AltText: bb("with strong__ text")}},
		}),
		lines("image/within_link.md", spans{
			{bb("["), md.Link{ReferenceID: "![kitten 1]"}},
			{bb("![kitten 1]"), md.Image{ReferenceID: "kitten 1", AltText: bb("kitten 1")}},
			{bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "![kitten 2]"}},
			{bb("![kitten 2]"), md.Image{ReferenceID: "kitten 2", AltText: bb("kitten 2")}},
			{bb("]"), md.End{}},
		}, head(5)),
		lines("link/direct_link.md", spans{
			{bb("["), md.Link{URL: "url"}},
			{bb("](url)"), md.End{}},
			{bb("["), md.Link{URL: "url", Title: "title"}},
			{bb(`](url "title")`), md.End{}},
		}),
		lines("link/direct_link_with_2separating_spaces.md", spans{
			{bb("["), md.Link{URL: "/example.html"}},
			{bb("]  (/example.html)"), md.End{}},
		}),
		blocks("link/direct_link_with_separating_newline.md", spans{
			{bb("["), md.Link{URL: "/example.html"}},
			{bb("]\n(/example.html)"), md.End{}},
		}),
		lines("link/direct_link_with_separating_space.md", spans{
			{bb("["), md.Link{URL: "http://example.net"}},
			{bb("] (http://example.net)"), md.End{}},
		}),
		lines("link/incomplete.md", spans{
			{bb("["), md.Link{ReferenceID: "ref undefined"}},
			{bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref 1"}},
			{bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref undefined"}},
			{bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref 1"}},
			{bb("]"), md.End{}},
		}, head(8)),
		blocks("link/link_text_with_newline.md", spans{
			{bb("["), md.Link{URL: "url1"}},
			{bb("](url1)"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref id"}},
			{bb("][]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref id"}},
			{bb("]"), md.End{}},
		}, head(9)),
		lines("link/link_title.md", spans{
			{bb("["), md.Link{URL: "url", Title: "title"}},
			{bb(`](url "title")`), md.End{}},
			{bb("["), md.Link{URL: "url", Title: "title"}},
			{bb(`](url 'title')`), md.End{}},
			{bb("["), md.Link{URL: "url", Title: `title 'with' \"quotes\"`}},
			{bb(`](url "title 'with' \"quotes\"")`), md.End{}},
			{bb("["), md.Link{URL: "url", Title: `title \'with\' "quotes"`}},
			{bb(`](url 'title \'with\' "quotes"')`), md.End{}},
			{bb("["), md.Link{URL: "url", Title: "title with (brackets)"}},
			{bb(`](url "title with (brackets)")`), md.End{}},
			{bb("["), md.Link{URL: "url", Title: "title with (brackets)"}},
			{bb(`](url 'title with (brackets)')`), md.End{}},
		}, head(6)),
		lines("link/link_with_parenthesis.md", spans{
			{bb("["), md.Link{URL: "url"}},
			{bb("](url)"), md.End{}},
			{bb("["), md.Link{URL: `url\`}},
			{bb(`](url\)`), md.End{}},
			{bb("["), md.Link{URL: `url)`, Title: "title"}},
			{bb(`](<url)> "title")`), md.End{}},
		}),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// link/multiple_ref_id_definitions.md
		lines("link/nested_links.md", spans{
			{bb("["), md.Link{URL: "url2"}},
			{bb("](url2)"), md.End{}},
		}),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// link/ref_case_sensitivity.md
		blocks("link/ref_id_matching.md", spans{
			{bb("["), md.Link{ReferenceID: "ref id"}}, {bb("][ref id]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref id"}}, {bb("][ref   id]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref id"}}, {bb("][  ref id  ]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref id"}}, {bb("][ref\n   id]"), md.End{}},

			{bb("["), md.Link{ReferenceID: "ref id"}}, {bb("][]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref id"}}, {bb("][]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref id"}}, {bb("][]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref id"}}, {bb("][]"), md.End{}},

			{bb("["), md.Link{ReferenceID: "ref id"}}, {bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref id"}}, {bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref id"}}, {bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref id"}}, {bb("]"), md.End{}},
		}, head(18)),
		// NOTE(akavel): below tests are not really interesting for us
		// here now.
		// link/ref_link.md
		// link/ref_link_empty.md
		// link/ref_link_self.md
		lines("link/ref_link_with_2separating_spaces.md", spans{
			{bb("["), md.Link{ReferenceID: "ref"}},
			{bb("]  [ref]"), md.End{}},
		}, head(2)),
		blocks("link/ref_link_with_separating_newline.md", spans{
			{bb("["), md.Link{ReferenceID: "ref"}},
			{bb("]\n[ref]"), md.End{}},
		}, head(3)),
		lines("link/ref_link_with_separating_space.md", spans{
			{bb("["), md.Link{ReferenceID: "ref"}},
			{bb("] [ref]"), md.End{}},
		}, head(2)),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// link/ref_resolution_within_other_blocks.md
		lines("link/square_brackets_in_link_or_ref.md", spans{
			{bb("["), md.Link{ReferenceID: "1"}}, {bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "2"}}, {bb("]"), md.End{}},
			{bb("["), md.Link{URL: "url"}}, {bb("](url)"), md.End{}},
			{bb("["), md.Link{URL: "url"}}, {bb("](url)"), md.End{}},
			{bb("["), md.Link{ReferenceID: "link"}}, {bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "3"}}, {bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "link"}}, {bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "4"}}, {bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: `ref\[3\]`}}, {bb(`][ref\[3\]]`), md.End{}},
			{bb("["), md.Link{ReferenceID: `ref\[4\]`}}, {bb(`][ref\[4\]]`), md.End{}},
			{bb("["), md.Link{ReferenceID: "link"}}, {bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "5"}}, {bb("]"), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref"}}, {bb("][ref]"), md.End{}},
			{bb("["), md.Link{ReferenceID: `ref\[5`}}, {bb(`][ref\[5]`), md.End{}},
			{bb("["), md.Link{ReferenceID: `ref\]6`}}, {bb(`][ref\]6]`), md.End{}},
		}, head(13)),
		lines("link/two_consecutive_refs.md", spans{
			{bb("["), md.Link{ReferenceID: `two`}}, {bb(`][two]`), md.End{}},
			{bb("["), md.Link{ReferenceID: `three`}}, {bb(`]`), md.End{}},
			{bb("["), md.Link{ReferenceID: `four`}}, {bb(`][four]`), md.End{}},
			{bb("["), md.Link{ReferenceID: `three`}}, {bb(`]`), md.End{}},
		}, head(4)),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// link/unused_ref.md
		lines("link/url_escapes.md", spans{
			// TODO(akavel): make sure we handled escaping properly in cases below
			{bb("["), md.Link{URL: `url\_\:\$\?`}}, {bb(`](url\_\:\$\?)`), md.End{}},
			{bb("["), md.Link{URL: `http://g&ouml;&ouml;gle.com`}}, {bb(`](http://g&ouml;&ouml;gle.com)`), md.End{}},
		}, head(2)),
		lines("link/url_in_angle_brackets.md", spans{
			{bb("["), md.Link{URL: `url`}}, {bb(`](<url>)`), md.End{}},
			{bb("["), md.Link{URL: `url(`}}, {bb(`](<url(>)`), md.End{}},
			{bb("["), md.Link{URL: `url)`}}, {bb(`](<url)>)`), md.End{}},
			{bb("["), md.Link{URL: `url)`, Title: "title"}}, {bb(`](<url)> "title")`), md.End{}},
		}, head(4)),
		lines("link/url_special_chars.md", spans{
			{bb("["), md.Link{URL: `url*#$%^&\~`}}, {bb(`](url*#$%^&\~)`), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref id1"}}, {bb(`][ref id1]`), md.End{}},
			{bb("["), md.Link{ReferenceID: "ref id1"}}, {bb(`]`), md.End{}},
			{bb("["), md.Link{ReferenceID: "link"}}, {bb(`]`), md.End{}},
		}, head(8)),
		blocks("link/url_whitespace.md", spans{
			{bb("["), md.Link{ReferenceID: "link"}}, {bb(`]`), md.End{}},
			{bb("["), md.Link{ReferenceID: "link"}}, {bb(`]`), md.End{}},
			{bb("["), md.Link{URL: "url1"}}, {bb("](<url 1>)"), md.End{}},
			{bb("["), md.Link{URL: "url1"}}, {bb("](<url \n   1>)"), md.End{}},
		}, head(6)),
		lines("link/vs_code.md", spans{
			{bb("["), md.Link{URL: "url"}},
			{bb("`code`"), md.Code{bb("code")}},
			{bb("](url)"), md.End{}},
			{bb("`containing [link](url)`"), md.Code{bb("containing [link](url)")}},
			{bb("`containing [link][ref]`"), md.Code{bb("containing [link][ref]")}},
			{bb("["), md.Link{ReferenceID: "ref"}},
			{bb("]"), md.End{}},
			{bb("`intertwined](url) with code`"), md.Code{bb("intertwined](url) with code")}},
			{bb("`intertwined [with code`"), md.Code{bb("intertwined [with code")}},
		}),
		lines("link/vs_emph.md", spans{
			{bb("["), md.Link{URL: "url"}},
			emB("*"), emE("*"),
			{bb("](url)"), md.End{}},
			{bb("["), md.Link{URL: "url"}},
			emB("**"), emE("**"),
			{bb("](url)"), md.End{}},
			{bb("["), md.Link{URL: "url"}},
			emB("_"), emE("_"),
			{bb("](url)"), md.End{}},
			{bb("["), md.Link{URL: "url"}},
			emB("__"), emE("__"),
			{bb("](url)"), md.End{}},

			emB("*"),
			{bb("["), md.Link{URL: "url"}}, {bb("](url)"), md.End{}},
			emE("*"),
			emB("**"),
			{bb("["), md.Link{URL: "url"}}, {bb("](url)"), md.End{}},
			emE("**"),
			emB("_"),
			{bb("["), md.Link{URL: "url"}}, {bb("](url)"), md.End{}},
			emE("_"),
			emB("__"),
			{bb("["), md.Link{URL: "url"}}, {bb("](url)"), md.End{}},
			emE("__"),

			{bb("["), md.Link{URL: "url"}}, {bb("](url)"), md.End{}},
			emB("*"), emE("*"),
			{bb("["), md.Link{URL: "url"}}, {bb("](url)"), md.End{}},
			emB("**"), emE("**"),
			{bb("["), md.Link{URL: "url"}}, {bb("](url)"), md.End{}},
			emB("_"), emE("_"),
			{bb("["), md.Link{URL: "url"}}, {bb("](url)"), md.End{}},
			emB("__"), emE("__"),
		}),
		lines("link/vs_image.md", spans{
			{bb("["), md.Link{URL: "url"}},
			{bb("![image](/image.jpg)"), md.Image{AltText: bb("image"), URL: "/image.jpg"}},
			{bb("](url)"), md.End{}},
		}),
	}
	for _, c := range cases {
		fmt.Printf("\ncase %s\n", c.fname)
		spans := []Span{}
		for _, b := range c.blocks {
			spans = append(spans, Parse(b, nil)...)
		}
		if !reflect.DeepEqual(c.spans, spans) {
			test.Errorf("case %s expected:\n%s",
				c.fname, spew.Sdump(c.spans))
			test.Errorf("got:")
			for i, span := range spans {
				off, err := utils.OffsetIn(c.buf, span.Pos)
				test.Errorf("[%d] @ %d [%v]: %s",
					i, off, err, spew.Sdump(span))
			}
			test.Errorf("blocks:\n%s", spew.Sdump(c.blocks))
			test.Errorf("expected vs. got DIFF:\n%s",
				diff.Diff(spew.Sdump(c.spans), spew.Sdump(spans)))
		}
	}
}

func init() {
	spew.Config.Indent = "  "
}

/*
TODO(akavel): tests for HTML after HTML is implemented:

in ROOT/testdata/tests/span_level:

code/vs_html.md
emphasis/vs_html.md
image/vs_html.md
link/vs_html.md
*/
