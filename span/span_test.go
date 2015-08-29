package span

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
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

func diff(ok, bad []Span) string {
	for i := range ok {
		if i >= len(bad) {
			return fmt.Sprintf("ends abruptly at position %d, expected:\n%s",
				i, spew.Sdump(ok[i]))
		}
		if !reflect.DeepEqual(ok[i], bad[i]) {
			return fmt.Sprintf("position %d, expected:\n%sgot:\n%s",
				i, spew.Sdump(ok[i]), spew.Sdump(bad[i]))
		}
	}
	return fmt.Sprintf("too many nodes, starting at position %d:\n%s",
		len(ok), spew.Sdump(bad[len(ok)]))
}

func emB(tag string) Span { return Span{bb(tag), EmphasisBegin{len(tag)}} }
func emE(tag string) Span { return Span{bb(tag), EmphasisEnd{len(tag)}} }

func TestSpan(test *testing.T) {
	cases := []spanCase{
		lines(`automatic_links/angle_brackets_in_link.md`, spans{
			{bb("http://exampl"), AutoLink{URL: `http://exampl`, Text: `http://exampl`}},
			// TODO(akavel): below is expected by testdata/, but
			// invalid according to spec, because preceding "<" is
			// not a 'word-separator' character (it has unicode
			// general class Sm - "Symbol, math"); try to resolve
			// this with the spec author.
			// {bb("http://exampl"), AutoLink{URL: `http://exampl`, Text: `http://exampl`}},
		}),
		lines("automatic_links/ending_with_punctuation.md", spans{
			{bb("http://example.net"), AutoLink{URL: "http://example.net", Text: "http://example.net"}},
			{bb("http://example.net/"), AutoLink{URL: "http://example.net/", Text: "http://example.net/"}},
			{bb("http://example.net"), AutoLink{URL: "http://example.net", Text: "http://example.net"}},
			{bb("http://example.net/"), AutoLink{URL: "http://example.net/", Text: "http://example.net/"}},

			{bb("<http://example.net,>"), AutoLink{URL: "http://example.net,", Text: "http://example.net,"}},
			{bb("<http://example.net/,>"), AutoLink{URL: "http://example.net/,", Text: "http://example.net/,"}},
			{bb("<http://example.net)>"), AutoLink{URL: "http://example.net)", Text: "http://example.net)"}},
			{bb("<http://example.net/)>"), AutoLink{URL: "http://example.net/)", Text: "http://example.net/)"}},
		}),
		lines("automatic_links/mail_url_in_angle_brackets.md", spans{
			{bb("<mailto:someone@example.net>"), AutoLink{URL: "mailto:someone@example.net", Text: "mailto:someone@example.net"}},
			{bb("<someone@example.net>"), AutoLink{URL: "mailto:someone@example.net", Text: "someone@example.net"}},
		}),
		lines("automatic_links/mail_url_without_angle_brackets.md", spans{
			// NOTE(akavel): below line is unexpected according to
			// testdata/, but from spec this seems totally expected,
			// so I added it
			{bb("mailto:someone@example.net"), AutoLink{URL: "mailto:someone@example.net", Text: "mailto:someone@example.net"}},
		}),
		lines("automatic_links/url_schemes.md", spans{
			{bb("http://example.net"), AutoLink{URL: "http://example.net", Text: "http://example.net"}},
			{bb("<http://example.net>"), AutoLink{URL: "http://example.net", Text: "http://example.net"}},
			{bb("file:///tmp/tmp.html"), AutoLink{URL: "file:///tmp/tmp.html", Text: "file:///tmp/tmp.html"}},
			{bb("<file:///tmp/tmp.html>"), AutoLink{URL: "file:///tmp/tmp.html", Text: "file:///tmp/tmp.html"}},
			{bb("feed://example.net/rss.xml"), AutoLink{URL: "feed://example.net/rss.xml", Text: "feed://example.net/rss.xml"}},
			{bb("<feed://example.net/rss.xml>"), AutoLink{URL: "feed://example.net/rss.xml", Text: "feed://example.net/rss.xml"}},
			{bb("googlechrome://example.net/"), AutoLink{URL: "googlechrome://example.net/", Text: "googlechrome://example.net/"}},
			{bb("<googlechrome://example.net/>"), AutoLink{URL: "googlechrome://example.net/", Text: "googlechrome://example.net/"}},
			{bb("`<>`"), Code{bb("<>")}},
			// NOTE(akavel): below line is unexpected according to
			// testdata/, but from spec this seems totally expected,
			// so I added it
			{bb("mailto:me@example.net"), AutoLink{URL: "mailto:me@example.net", Text: "mailto:me@example.net"}},
			{bb("<mailto:me@example.net>"), AutoLink{URL: "mailto:me@example.net", Text: "mailto:me@example.net"}},
		}),
		lines("automatic_links/url_special_chars.md", spans{
			{bb(`http://example.net/*#$%^&\~/blah`), AutoLink{URL: `http://example.net/*#$%^&\~/blah`, Text: `http://example.net/*#$%^&\~/blah`}},
			{bb(`<http://example.net/*#$%^&\~)/blah>`), AutoLink{URL: `http://example.net/*#$%^&\~)/blah`, Text: `http://example.net/*#$%^&\~)/blah`}},
			// NOTE(akavel): testdata expects below commented entry,
			// but this seems wrong compared to spec; I've added
			// fixed entry
			// {bb(`http://example.net/blah/`), AutoLink{URL: `http://example.net/blah/`, Text: `http://example.net/blah/`}},
			{bb(`http://example.net/blah/*#$%^&\~`), AutoLink{URL: `http://example.net/blah/*#$%^&\~`, Text: `http://example.net/blah/*#$%^&\~`}},
			{bb(`<http://example.net/blah/*#$%^&\~)>`), AutoLink{URL: `http://example.net/blah/*#$%^&\~)`, Text: `http://example.net/blah/*#$%^&\~)`}},
		}),
		lines("automatic_links/web_url_in_angle_brackets.md", spans{
			{bb("<http://example.net/path/>"), AutoLink{URL: "http://example.net/path/", Text: "http://example.net/path/"}},
			{bb("<https://example.net/path/>"), AutoLink{URL: "https://example.net/path/", Text: "https://example.net/path/"}},
			{bb("<ftp://example.net/path/>"), AutoLink{URL: "ftp://example.net/path/", Text: "ftp://example.net/path/"}},
		}),
		lines("automatic_links/web_url_without_angle_brackets.md", spans{
			{bb("http://example.net/path/"), AutoLink{URL: "http://example.net/path/", Text: "http://example.net/path/"}},
			{bb("https://example.net/path/"), AutoLink{URL: "https://example.net/path/", Text: "https://example.net/path/"}},
			{bb("ftp://example.net/path/"), AutoLink{URL: "ftp://example.net/path/", Text: "ftp://example.net/path/"}},
		}),
		lines("code/end_of_codespan.md", spans{
			{bb("`code span`"), Code{bb("code span")}},
			{bb("``code span` ends``"), Code{bb("code span` ends")}},
			{bb("`code span`` ends`"), Code{bb("code span`` ends")}},
			{bb("````code span`` ``ends````"), Code{bb("code span`` ``ends")}},
			{bb("`code span\\`"), Code{bb(`code span\`)}},
		}),
		blocks("code/multiline.md", spans{
			{bb("`code span\ncan span multiple\nlines`"), Code{bb("code span\ncan span multiple\nlines")}},
		}),
		lines("code/vs_emph.md", spans{
			{bb("`code containing *em* text`"), Code{bb("code containing *em* text")}},
			{bb("`code containing **strong** text`"), Code{bb("code containing **strong** text")}},
			{bb("`code containing _em_ text`"), Code{bb("code containing _em_ text")}},
			{bb("`code containing __strong__ text`"), Code{bb("code containing __strong__ text")}},

			{bb("*"), EmphasisBegin{Level: 1}},
			{bb("`code`"), Code{bb("code")}},
			{bb("*"), EmphasisEnd{Level: 1}},
			{bb("**"), EmphasisBegin{Level: 2}},
			{bb("`code`"), Code{bb("code")}},
			{bb("**"), EmphasisEnd{Level: 2}},
			{bb("_"), EmphasisBegin{Level: 1}},
			{bb("`code`"), Code{bb("code")}},
			{bb("_"), EmphasisEnd{Level: 1}},
			{bb("__"), EmphasisBegin{Level: 2}},
			{bb("`code`"), Code{bb("code")}},
			{bb("__"), EmphasisEnd{Level: 2}},

			{bb("`code *intertwined`"), Code{bb("code *intertwined")}},
			{bb("`with em* text`"), Code{bb("with em* text")}},
			{bb("`code **intertwined`"), Code{bb("code **intertwined")}},
			{bb("`with strong** text`"), Code{bb("with strong** text")}},
			{bb("`code _intertwined`"), Code{bb("code _intertwined")}},
			{bb("`with em_ text`"), Code{bb("with em_ text")}},
			{bb("`code __intertwined`"), Code{bb("code __intertwined")}},
			{bb("`with strong__ text`"), Code{bb("with strong__ text")}},
		}),
		lines("code/vs_image.md", spans{
			{bb("`code containing ![image](url)`"), Code{bb("code containing ![image](url)")}},
			{bb("`code containing ![image][ref]`"), Code{bb("code containing ![image][ref]")}},
			{bb("`code containing ![ref]`"), Code{bb("code containing ![ref]")}},

			{bb("`containing code`"), Code{bb("containing code")}},
			{bb("`containing code`"), Code{bb("containing code")}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "ref"}},
			{bb("`containing code`"), Code{bb("containing code")}},

			{bb("`code ![intertwined`"), Code{bb("code ![intertwined")}},
			{bb("`intertwined](with) image`"), Code{bb("intertwined](with) image")}},
			{bb("`code ![intertwined`"), Code{bb("code ![intertwined")}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "ref"}},
			{bb("`intertwined with][ref] image`"), Code{bb("intertwined with][ref] image")}},
			{bb("`code ![intertwined`"), Code{bb("code ![intertwined")}},
			{bb("`with] image`"), Code{bb("with] image")}},
		}, head(30)),
		lines("code/vs_link.md", spans{
			{bb("`code containing [link](url)`"), Code{bb("code containing [link](url)")}},
			{bb("`code containing [link][ref]`"), Code{bb("code containing [link][ref]")}},
			{bb("`code containing [ref]`"), Code{bb("code containing [ref]")}},

			{bb("["), LinkBegin{}},
			{bb("`containing code`"), Code{bb("containing code")}},
			{bb("](url)"), LinkEnd{URL: "url"}},
			{bb("["), LinkBegin{}},
			{bb("`containing code`"), Code{bb("containing code")}},
			{bb("][ref]"), LinkEnd{ReferenceID: "ref"}},
			{bb("["), LinkBegin{}},
			{bb("`containing code`"), Code{bb("containing code")}},
			{bb("]"), LinkEnd{ReferenceID: "link `containing code`"}},

			{bb("`code [intertwined`"), Code{bb("code [intertwined")}},
			{bb("`intertwined](with) link`"), Code{bb("intertwined](with) link")}},
			{bb("`code [intertwined`"), Code{bb("code [intertwined")}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "ref"}},
			{bb("`intertwined with][ref] link`"), Code{bb("intertwined with][ref] link")}},
			{bb("`code [intertwined`"), Code{bb("code [intertwined")}},
			{bb("`with] link`"), Code{bb("with] link")}},
		}, head(30)),
		lines("code/well_formed.md", spans{
			{bb("`code span`"), Code{bb("code span")}},
			{bb("``code ` span``"), Code{bb("code ` span")}},
			{bb("`` ` code span``"), Code{bb("` code span")}},
			{bb("``code span ` ``"), Code{bb("code span `")}},
			{bb("`` `code span` ``"), Code{bb("`code span`")}},
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
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "_"}},
			emB("_"), emE("_"), emB("_"), emE("_"),
			// NOTE(akavel): link below not expected in testdata
			// because it's not defined below; but we leave this to
			// user.
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "_"}},
		}),
		lines("image/direct_link.md", spans{
			{bb("![image](url)"), Image{AltText: bb("image"), URL: "url"}},
			{bb(`![image](url "title")`), Image{AltText: bb("image"), URL: "url", Title: "title"}},
		}),
		lines("image/direct_link_with_2separating_spaces.md", spans{
			{bb("![linking]  (/img.png)"), Image{AltText: bb("linking"), URL: "/img.png"}},
		}),
		blocks("image/direct_link_with_separating_newline.md", spans{
			{bb("![link]\n(/img.png)"), Image{AltText: bb("link"), URL: "/img.png"}},
		}),
		lines("image/direct_link_with_separating_space.md", spans{
			{bb("![link] (http://example.net/img.png)"), Image{AltText: bb("link"), URL: "http://example.net/img.png"}},
		}),
		lines("image/image_title.md", spans{
			{bb(`![link](url "title")`), Image{AltText: bb("link"), URL: "url", Title: `title`}},
			{bb(`![link](url 'title')`), Image{AltText: bb("link"), URL: "url", Title: `title`}},
			// TODO(akavel): unquote contents of Title when
			// processing? doesn't seem noted in spec, send fix for
			// spec?
			{bb(`![link](url "title 'with' \"quotes\"")`), Image{AltText: bb("link"), URL: "url", Title: `title 'with' \"quotes\"`}},
			{bb(`![link](url 'title \'with\' "quotes"')`), Image{AltText: bb("link"), URL: "url", Title: `title \'with\' "quotes"`}},
			{bb(`![link](url "title with (brackets)")`), Image{AltText: bb("link"), URL: "url", Title: `title with (brackets)`}},
			{bb(`![link](url 'title with (brackets)')`), Image{AltText: bb("link"), URL: "url", Title: `title with (brackets)`}},

			{bb("![ref id1]"), Image{ReferenceID: "ref id1", AltText: bb("ref id1")}},
			{bb("![ref id2]"), Image{ReferenceID: "ref id2", AltText: bb("ref id2")}},
			{bb("![ref id3]"), Image{ReferenceID: "ref id3", AltText: bb("ref id3")}},
			{bb("![ref id4]"), Image{ReferenceID: "ref id4", AltText: bb("ref id4")}},
			{bb("![ref id5]"), Image{ReferenceID: "ref id5", AltText: bb("ref id5")}},
			{bb("![ref id6]"), Image{ReferenceID: "ref id6", AltText: bb("ref id6")}},
			{bb("![ref id7]"), Image{ReferenceID: "ref id7", AltText: bb("ref id7")}},
			{bb("![ref id8]"), Image{ReferenceID: "ref id8", AltText: bb("ref id8")}},
			{bb("![ref id9]"), Image{ReferenceID: "ref id9", AltText: bb("ref id9")}},
			{bb("![ref id10]"), Image{ReferenceID: "ref id10", AltText: bb("ref id10")}},
			{bb("![ref id11]"), Image{ReferenceID: "ref id11", AltText: bb("ref id11")}},
			{bb("![ref id12]"), Image{ReferenceID: "ref id12", AltText: bb("ref id12")}},
		}, head(19)),
		lines("image/incomplete.md", spans{
			{bb("![ref undefined]"), Image{ReferenceID: "ref undefined", AltText: bb("ref undefined")}},
			{bb("![ref 1]"), Image{ReferenceID: "ref 1", AltText: bb("ref 1")}},
			{bb("![ref undefined]"), Image{ReferenceID: "ref undefined", AltText: bb("ref undefined")}},
			{bb("![ref 1]"), Image{ReferenceID: "ref 1", AltText: bb("ref 1")}},
		}, head(8)),
		blocks("image/link_text_with_newline.md", spans{
			{bb("![link\ntext](url1)"), Image{AltText: bb("link\ntext"), URL: "url1"}},
			{bb("![ref\nid][]"), Image{ReferenceID: "ref id", AltText: bb("ref\nid")}},
			{bb("![ref\nid]"), Image{ReferenceID: "ref id", AltText: bb("ref\nid")}},
		}, head(9)),
		lines("image/link_with_parenthesis.md", spans{
			{bb("![bad link](url)"), Image{URL: "url", AltText: bb("bad link")}},
			{bb(`![bad link](url\)`), Image{URL: `url\`, AltText: bb("bad link")}},
			{bb(`![link](<url)> "title")`), Image{URL: `url)`, AltText: bb("link"), Title: "title"}},
		}),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// lines("image/multiple_ref_id_definitions.md", spans{}),
		lines("image/nested_images.md", spans{
			{bb("![link2](url2)"), Image{AltText: bb("link2"), URL: "url2"}},
		}),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// lines("image/ref_case_sensitivity.md", spans{}),
		blocks("image/ref_id_matching.md", spans{
			{bb("![link][ref id]"), Image{ReferenceID: "ref id", AltText: bb("link")}},
			{bb("![link][ref   id]"), Image{ReferenceID: "ref id", AltText: bb("link")}},
			{bb("![link][  ref id  ]"), Image{ReferenceID: "ref id", AltText: bb("link")}},
			{bb("![link][ref\n   id]"), Image{ReferenceID: "ref id", AltText: bb("link")}},
			{bb("![ref id][]"), Image{ReferenceID: "ref id", AltText: bb("ref id")}},
			{bb("![ref   id][]"), Image{ReferenceID: "ref id", AltText: bb("ref   id")}},
			{bb("![  ref id  ][]"), Image{ReferenceID: "ref id", AltText: bb("  ref id  ")}},
			{bb("![ref\n   id][]"), Image{ReferenceID: "ref id", AltText: bb("ref\n   id")}},
			{bb("![ref id]"), Image{ReferenceID: "ref id", AltText: bb("ref id")}},
			{bb("![ref   id]"), Image{ReferenceID: "ref id", AltText: bb("ref   id")}},
			{bb("![  ref id  ]"), Image{ReferenceID: "ref id", AltText: bb("  ref id  ")}},
			{bb("![ref\n   id]"), Image{ReferenceID: "ref id", AltText: bb("ref\n   id")}},
		}, head(18)),
		// NOTE(akavel): below tests are not really interesting for us
		// here now.
		// lines("image/ref_link.md", spans{}),
		// lines("image/ref_link_empty.md", spans{}),
		// lines("image/ref_link_self.md", spans{}),
		lines("image/ref_link_with_2separating_spaces.md", spans{
			{bb("![link]  [ref]"), Image{AltText: bb("link"), ReferenceID: "ref"}},
		}, head(2)),
		blocks("image/ref_link_with_separating_newline.md", spans{
			{bb("![link]\n[ref]"), Image{AltText: bb("link"), ReferenceID: "ref"}},
		}, head(3)),
		lines("image/ref_link_with_separating_space.md", spans{
			{bb("![link] [ref]"), Image{AltText: bb("link"), ReferenceID: "ref"}},
		}, head(2)),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// lines("image/ref_resolution_within_other_blocks.md", spans{}),
		lines("image/square_brackets_in_link_or_ref.md", spans{
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "1"}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "2"}},
			{bb("![2]"), Image{ReferenceID: "2", AltText: bb("2")}},
			// TODO(akavel): make sure we handled escaping properly in cases below
			{bb(`![link\[1\]](url)`), Image{URL: "url", AltText: bb(`link\[1\]`)}},
			{bb(`![link\[2\]](url)`), Image{URL: "url", AltText: bb(`link\[2\]`)}},
			{bb(`![link!\[2\]](url)`), Image{URL: "url", AltText: bb(`link!\[2\]`)}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "2"}},
			{bb("![link]"), Image{ReferenceID: "link", AltText: bb("link")}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "3"}},
			{bb("![link]"), Image{ReferenceID: "link", AltText: bb("link")}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "4"}},
			// TODO(akavel): make sure we handled escaping properly in cases below
			{bb(`![link][ref\[3\]]`), Image{ReferenceID: `ref\[3\]`, AltText: bb(`link`)}},
			{bb(`![link][ref\[4\]]`), Image{ReferenceID: `ref\[4\]`, AltText: bb(`link`)}},
			{bb("![link]"), Image{ReferenceID: "link", AltText: bb("link")}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "5"}},
			{bb("![link][ref]"), Image{ReferenceID: "ref", AltText: bb("link")}},
			{bb(`![link][ref\[5]`), Image{ReferenceID: `ref\[5`, AltText: bb(`link`)}},
			{bb(`![link][ref\]6]`), Image{ReferenceID: `ref\]6`, AltText: bb(`link`)}},
		}, head(16)),
		lines("image/two_consecutive_refs.md", spans{
			{bb("![one][two]"), Image{ReferenceID: "two", AltText: bb("one")}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "three"}},
			{bb("![one][four]"), Image{ReferenceID: "four", AltText: bb("one")}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "three"}},
		}, head(4)),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// image/unused_ref.md
		lines("image/url_escapes.md", spans{
			{bb(`![link](url\_\:\$\?)`), Image{AltText: bb("link"), URL: `url\_\:\$\?`}},
			{bb(`![link](http://g&ouml;&ouml;gle.com)`), Image{AltText: bb("link"), URL: `http://g&ouml;&ouml;gle.com`}},
		}, head(2)),
		lines("image/url_in_angle_brackets.md", spans{
			{bb(`![link](<url>)`), Image{AltText: bb("link"), URL: "url"}},
			{bb(`![link](<url(>)`), Image{AltText: bb("link"), URL: "url("}},
			{bb(`![link](<url)>)`), Image{AltText: bb("link"), URL: "url)"}},
			{bb(`![link](<url)> "title")`), Image{AltText: bb("link"), URL: "url)", Title: "title"}},
		}, head(4)),
		lines("image/url_special_chars.md", spans{
			{bb(`![link](url*#$%^&\~)`), Image{AltText: bb("link"), URL: `url*#$%^&\~`}},
			{bb("![link][ref id1]"), Image{ReferenceID: "ref id1", AltText: bb("link")}},
			{bb("![ref id1]"), Image{ReferenceID: "ref id1", AltText: bb("ref id1")}},
			{bb("![link]"), Image{ReferenceID: "link", AltText: bb("link")}},
		}, head(8)),
		blocks("image/url_whitespace.md", spans{
			{bb("![link]"), Image{AltText: bb("link"), ReferenceID: "link"}},
			{bb("![link]"), Image{AltText: bb("link"), ReferenceID: "link"}},
			{bb("![link](<url 1>)"), Image{AltText: bb("link"), URL: "url1"}},
			{bb("![link](<url \n   1>)"), Image{AltText: bb("link"), URL: "url1"}},
		}, head(6)),
		lines("image/vs_code.md", spans{
			{bb("`code`"), Code{bb("code")}},
			{bb("`containing ![image](url)`"), Code{bb("containing ![image](url)")}},
			{bb("`containing ![image][ref]`"), Code{bb("containing ![image][ref]")}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "ref"}},
			{bb("`intertwined](url) with code`"), Code{bb("intertwined](url) with code")}},
			{bb("`intertwined ![with code`"), Code{bb("intertwined ![with code")}},
		}),
		lines("image/vs_emph.md", spans{
			{bb("![image containing *em* text](url)"), Image{URL: "url", AltText: bb("image containing *em* text")}},
			{bb("![image containing **strong** text](url)"), Image{URL: "url", AltText: bb("image containing **strong** text")}},
			{bb("![image containing _em_ text](url)"), Image{URL: "url", AltText: bb("image containing _em_ text")}},
			{bb("![image containing __strong__ text](url)"), Image{URL: "url", AltText: bb("image containing __strong__ text")}},

			emB("*"),
			{bb("![image](url)"), Image{AltText: bb("image"), URL: "url"}},
			emE("*"),
			emB("**"),
			{bb("![image](url)"), Image{AltText: bb("image"), URL: "url"}},
			emE("**"),
			emB("_"),
			{bb("![image](url)"), Image{AltText: bb("image"), URL: "url"}},
			emE("_"),
			emB("__"),
			{bb("![image](url)"), Image{AltText: bb("image"), URL: "url"}},
			emE("__"),

			{bb("![image *intertwined](url)"), Image{URL: "url", AltText: bb("image *intertwined")}},
			{bb("![with em* text](url)"), Image{URL: "url", AltText: bb("with em* text")}},
			{bb("![image **intertwined](url)"), Image{URL: "url", AltText: bb("image **intertwined")}},
			{bb("![with strong** text](url)"), Image{URL: "url", AltText: bb("with strong** text")}},
			{bb("![image _intertwined](url)"), Image{URL: "url", AltText: bb("image _intertwined")}},
			{bb("![with em_ text](url)"), Image{URL: "url", AltText: bb("with em_ text")}},
			{bb("![image __intertwined](url)"), Image{URL: "url", AltText: bb("image __intertwined")}},
			{bb("![with strong__ text](url)"), Image{URL: "url", AltText: bb("with strong__ text")}},
		}),
		lines("image/within_link.md", spans{
			{bb("["), LinkBegin{}},
			{bb("![kitten 1]"), Image{ReferenceID: "kitten 1", AltText: bb("kitten 1")}},
			{bb("]"), LinkEnd{ReferenceID: "![kitten 1]"}},
			{bb("["), LinkBegin{}},
			{bb("![kitten 2]"), Image{ReferenceID: "kitten 2", AltText: bb("kitten 2")}},
			{bb("]"), LinkEnd{ReferenceID: "![kitten 2]"}},
		}, head(5)),
		lines("link/direct_link.md", spans{
			{bb("["), LinkBegin{}},
			{bb("](url)"), LinkEnd{URL: "url"}},
			{bb("["), LinkBegin{}},
			{bb(`](url "title")`), LinkEnd{URL: "url", Title: "title"}},
		}),
		lines("link/direct_link_with_2separating_spaces.md", spans{
			{bb("["), LinkBegin{}},
			{bb("]  (/example.html)"), LinkEnd{URL: "/example.html"}},
		}),
		blocks("link/direct_link_with_separating_newline.md", spans{
			{bb("["), LinkBegin{}},
			{bb("]\n(/example.html)"), LinkEnd{URL: "/example.html"}},
		}),
		lines("link/direct_link_with_separating_space.md", spans{
			{bb("["), LinkBegin{}},
			{bb("] (http://example.net)"), LinkEnd{URL: "http://example.net"}},
		}),
		lines("link/incomplete.md", spans{
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "ref undefined"}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "ref 1"}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "ref undefined"}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "ref 1"}},
		}, head(8)),
		blocks("link/link_text_with_newline.md", spans{
			{bb("["), LinkBegin{}},
			{bb("](url1)"), LinkEnd{URL: "url1"}},
			{bb("["), LinkBegin{}},
			{bb("][]"), LinkEnd{ReferenceID: "ref id"}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "ref id"}},
		}, head(9)),
		lines("link/link_title.md", spans{
			{bb("["), LinkBegin{}},
			{bb(`](url "title")`), LinkEnd{URL: "url", Title: "title"}},
			{bb("["), LinkBegin{}},
			{bb(`](url 'title')`), LinkEnd{URL: "url", Title: "title"}},
			{bb("["), LinkBegin{}},
			{bb(`](url "title 'with' \"quotes\"")`), LinkEnd{URL: "url", Title: `title 'with' \"quotes\"`}},
			{bb("["), LinkBegin{}},
			{bb(`](url 'title \'with\' "quotes"')`), LinkEnd{URL: "url", Title: `title \'with\' "quotes"`}},
			{bb("["), LinkBegin{}},
			{bb(`](url "title with (brackets)")`), LinkEnd{URL: "url", Title: "title with (brackets)"}},
			{bb("["), LinkBegin{}},
			{bb(`](url 'title with (brackets)')`), LinkEnd{URL: "url", Title: "title with (brackets)"}},
		}, head(6)),
		lines("link/link_with_parenthesis.md", spans{
			{bb("["), LinkBegin{}},
			{bb("](url)"), LinkEnd{URL: "url"}},
			{bb("["), LinkBegin{}},
			{bb(`](url\)`), LinkEnd{URL: `url\`}},
			{bb("["), LinkBegin{}},
			{bb(`](<url)> "title")`), LinkEnd{URL: `url)`, Title: "title"}},
		}),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// link/multiple_ref_id_definitions.md
		lines("link/nested_links.md", spans{
			{bb("["), LinkBegin{}},
			{bb("](url2)"), LinkEnd{URL: "url2"}},
		}),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// link/ref_case_sensitivity.md
		blocks("link/ref_id_matching.md", spans{
			{bb("["), LinkBegin{}}, {bb("][ref id]"), LinkEnd{ReferenceID: "ref id"}},
			{bb("["), LinkBegin{}}, {bb("][ref   id]"), LinkEnd{ReferenceID: "ref id"}},
			{bb("["), LinkBegin{}}, {bb("][  ref id  ]"), LinkEnd{ReferenceID: "ref id"}},
			{bb("["), LinkBegin{}}, {bb("][ref\n   id]"), LinkEnd{ReferenceID: "ref id"}},

			{bb("["), LinkBegin{}}, {bb("][]"), LinkEnd{ReferenceID: "ref id"}},
			{bb("["), LinkBegin{}}, {bb("][]"), LinkEnd{ReferenceID: "ref id"}},
			{bb("["), LinkBegin{}}, {bb("][]"), LinkEnd{ReferenceID: "ref id"}},
			{bb("["), LinkBegin{}}, {bb("][]"), LinkEnd{ReferenceID: "ref id"}},

			{bb("["), LinkBegin{}}, {bb("]"), LinkEnd{ReferenceID: "ref id"}},
			{bb("["), LinkBegin{}}, {bb("]"), LinkEnd{ReferenceID: "ref id"}},
			{bb("["), LinkBegin{}}, {bb("]"), LinkEnd{ReferenceID: "ref id"}},
			{bb("["), LinkBegin{}}, {bb("]"), LinkEnd{ReferenceID: "ref id"}},
		}, head(18)),
		// NOTE(akavel): below tests are not really interesting for us
		// here now.
		// link/ref_link.md
		// link/ref_link_empty.md
		// link/ref_link_self.md
		lines("link/ref_link_with_2separating_spaces.md", spans{
			{bb("["), LinkBegin{}},
			{bb("]  [ref]"), LinkEnd{ReferenceID: "ref"}},
		}, head(2)),
		blocks("link/ref_link_with_separating_newline.md", spans{
			{bb("["), LinkBegin{}},
			{bb("]\n[ref]"), LinkEnd{ReferenceID: "ref"}},
		}, head(3)),
		lines("link/ref_link_with_separating_space.md", spans{
			{bb("["), LinkBegin{}},
			{bb("] [ref]"), LinkEnd{ReferenceID: "ref"}},
		}, head(2)),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// link/ref_resolution_within_other_blocks.md
		lines("link/square_brackets_in_link_or_ref.md", spans{
			{bb("["), LinkBegin{}}, {bb("]"), LinkEnd{ReferenceID: "1"}},
			{bb("["), LinkBegin{}}, {bb("]"), LinkEnd{ReferenceID: "2"}},
			{bb("["), LinkBegin{}}, {bb("](url)"), LinkEnd{URL: "url"}},
			{bb("["), LinkBegin{}}, {bb("](url)"), LinkEnd{URL: "url"}},
			{bb("["), LinkBegin{}}, {bb("]"), LinkEnd{ReferenceID: "link"}},
			{bb("["), LinkBegin{}}, {bb("]"), LinkEnd{ReferenceID: "3"}},
			{bb("["), LinkBegin{}}, {bb("]"), LinkEnd{ReferenceID: "link"}},
			{bb("["), LinkBegin{}}, {bb("]"), LinkEnd{ReferenceID: "4"}},
			{bb("["), LinkBegin{}}, {bb(`][ref\[3\]]`), LinkEnd{ReferenceID: `ref\[3\]`}},
			{bb("["), LinkBegin{}}, {bb(`][ref\[4\]]`), LinkEnd{ReferenceID: `ref\[4\]`}},
			{bb("["), LinkBegin{}}, {bb("]"), LinkEnd{ReferenceID: "link"}},
			{bb("["), LinkBegin{}}, {bb("]"), LinkEnd{ReferenceID: "5"}},
			{bb("["), LinkBegin{}}, {bb("][ref]"), LinkEnd{ReferenceID: "ref"}},
			{bb("["), LinkBegin{}}, {bb(`][ref\[5]`), LinkEnd{ReferenceID: `ref\[5`}},
			{bb("["), LinkBegin{}}, {bb(`][ref\]6]`), LinkEnd{ReferenceID: `ref\]6`}},
		}, head(13)),
		lines("link/two_consecutive_refs.md", spans{
			{bb("["), LinkBegin{}}, {bb(`][two]`), LinkEnd{ReferenceID: `two`}},
			{bb("["), LinkBegin{}}, {bb(`]`), LinkEnd{ReferenceID: `three`}},
			{bb("["), LinkBegin{}}, {bb(`][four]`), LinkEnd{ReferenceID: `four`}},
			{bb("["), LinkBegin{}}, {bb(`]`), LinkEnd{ReferenceID: `three`}},
		}, head(4)),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// link/unused_ref.md
		lines("link/url_escapes.md", spans{
			// TODO(akavel): make sure we handled escaping properly in cases below
			{bb("["), LinkBegin{}}, {bb(`](url\_\:\$\?)`), LinkEnd{URL: `url\_\:\$\?`}},
			{bb("["), LinkBegin{}}, {bb(`](http://g&ouml;&ouml;gle.com)`), LinkEnd{URL: `http://g&ouml;&ouml;gle.com`}},
		}, head(2)),
		lines("link/url_in_angle_brackets.md", spans{
			{bb("["), LinkBegin{}}, {bb(`](<url>)`), LinkEnd{URL: `url`}},
			{bb("["), LinkBegin{}}, {bb(`](<url(>)`), LinkEnd{URL: `url(`}},
			{bb("["), LinkBegin{}}, {bb(`](<url)>)`), LinkEnd{URL: `url)`}},
			{bb("["), LinkBegin{}}, {bb(`](<url)> "title")`), LinkEnd{URL: `url)`, Title: "title"}},
		}, head(4)),
		lines("link/url_special_chars.md", spans{
			{bb("["), LinkBegin{}}, {bb(`](url*#$%^&\~)`), LinkEnd{URL: `url*#$%^&\~`}},
			{bb("["), LinkBegin{}}, {bb(`][ref id1]`), LinkEnd{ReferenceID: "ref id1"}},
			{bb("["), LinkBegin{}}, {bb(`]`), LinkEnd{ReferenceID: "ref id1"}},
			{bb("["), LinkBegin{}}, {bb(`]`), LinkEnd{ReferenceID: "link"}},
		}, head(8)),
		blocks("link/url_whitespace.md", spans{
			{bb("["), LinkBegin{}}, {bb(`]`), LinkEnd{ReferenceID: "link"}},
			{bb("["), LinkBegin{}}, {bb(`]`), LinkEnd{ReferenceID: "link"}},
			{bb("["), LinkBegin{}}, {bb("](<url 1>)"), LinkEnd{URL: "url1"}},
			{bb("["), LinkBegin{}}, {bb("](<url \n   1>)"), LinkEnd{URL: "url1"}},
		}, head(6)),
		lines("link/vs_code.md", spans{
			{bb("["), LinkBegin{}},
			{bb("`code`"), Code{bb("code")}},
			{bb("](url)"), LinkEnd{URL: "url"}},
			{bb("`containing [link](url)`"), Code{bb("containing [link](url)")}},
			{bb("`containing [link][ref]`"), Code{bb("containing [link][ref]")}},
			{bb("["), LinkBegin{}},
			{bb("]"), LinkEnd{ReferenceID: "ref"}},
			{bb("`intertwined](url) with code`"), Code{bb("intertwined](url) with code")}},
			{bb("`intertwined [with code`"), Code{bb("intertwined [with code")}},
		}),
		lines("link/vs_emph.md", spans{
			{bb("["), LinkBegin{}},
			emB("*"), emE("*"),
			{bb("](url)"), LinkEnd{URL: "url"}},
			{bb("["), LinkBegin{}},
			emB("**"), emE("**"),
			{bb("](url)"), LinkEnd{URL: "url"}},
			{bb("["), LinkBegin{}},
			emB("_"), emE("_"),
			{bb("](url)"), LinkEnd{URL: "url"}},
			{bb("["), LinkBegin{}},
			emB("__"), emE("__"),
			{bb("](url)"), LinkEnd{URL: "url"}},

			emB("*"),
			{bb("["), LinkBegin{}}, {bb("](url)"), LinkEnd{URL: "url"}},
			emE("*"),
			emB("**"),
			{bb("["), LinkBegin{}}, {bb("](url)"), LinkEnd{URL: "url"}},
			emE("**"),
			emB("_"),
			{bb("["), LinkBegin{}}, {bb("](url)"), LinkEnd{URL: "url"}},
			emE("_"),
			emB("__"),
			{bb("["), LinkBegin{}}, {bb("](url)"), LinkEnd{URL: "url"}},
			emE("__"),

			{bb("["), LinkBegin{}}, {bb("](url)"), LinkEnd{URL: "url"}},
			emB("*"), emE("*"),
			{bb("["), LinkBegin{}}, {bb("](url)"), LinkEnd{URL: "url"}},
			emB("**"), emE("**"),
			{bb("["), LinkBegin{}}, {bb("](url)"), LinkEnd{URL: "url"}},
			emB("_"), emE("_"),
			{bb("["), LinkBegin{}}, {bb("](url)"), LinkEnd{URL: "url"}},
			emB("__"), emE("__"),
		}),
		lines("link/vs_image.md", spans{
			{bb("["), LinkBegin{}},
			{bb("![image](/image.jpg)"), Image{AltText: bb("image"), URL: "/image.jpg"}},
			{bb("](url)"), LinkEnd{URL: "url"}},
		}),
	}
	for _, c := range cases {
		fmt.Printf("\ncase %s\n", c.fname)
		spans := []Span{}
		for _, b := range c.blocks {
			spans = append(spans, Split(b, nil)...)
		}
		if !reflect.DeepEqual(c.spans, spans) {
			test.Errorf("case %s expected:\n%s",
				c.fname, spew.Sdump(c.spans))
			test.Errorf("got:")
			for i, span := range spans {
				off, err := span.OffsetIn(c.buf)
				test.Errorf("[%d] @ %d [%v]: %s",
					i, off, err, spew.Sdump(span))
			}
			test.Errorf("blocks:\n%s", spew.Sdump(c.blocks))
			test.Errorf("QUICK DIFF: %s\n", diff(c.spans, spans))
		}
	}
}

/*
TODO(akavel): tests for HTML after HTML is implemented:

in ROOT/testdata/tests/span_level:

code/vs_html.md
emphasis/vs_html.md
image/vs_html.md
link/vs_html.md
*/
