package mdspan

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kylelemons/godebug/diff"

	"gopkg.in/akavel/vfmd.v0/md"
)

func bb(s string) []byte { return []byte(s) }

type spanCase struct {
	fname  string
	buf    []byte
	blocks [][]byte
	spans  []md.Tag
}

type spans []md.Tag

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

func emB(tag string) md.Tag { return md.Emphasis{len(tag)} }
func emE(tag string) md.Tag { return md.End{} }

func TestSpan(test *testing.T) {
	cases := []spanCase{
		lines(`automatic_links/angle_brackets_in_link.md`, spans{
			md.AutomaticLink{URL: `http://exampl`, Text: `http://exampl`},
			// TODO(akavel): below is expected by testdata/, but
			// invalid according to spec, because preceding "<" is
			// not a 'word-separator' character (it has unicode
			// general class Sm - "Symbol, math"); try to resolve
			// this with the spec author.
			// {bb("http://exampl"), md.AutomaticLink{URL: `http://exampl`, Text: `http://exampl`}},
		}),
		lines("automatic_links/ending_with_punctuation.md", spans{
			md.AutomaticLink{URL: "http://example.net", Text: "http://example.net"},
			md.AutomaticLink{URL: "http://example.net/", Text: "http://example.net/"},
			md.AutomaticLink{URL: "http://example.net", Text: "http://example.net"},
			md.AutomaticLink{URL: "http://example.net/", Text: "http://example.net/"},

			md.AutomaticLink{URL: "http://example.net,", Text: "http://example.net,"},
			md.AutomaticLink{URL: "http://example.net/,", Text: "http://example.net/,"},
			md.AutomaticLink{URL: "http://example.net)", Text: "http://example.net)"},
			md.AutomaticLink{URL: "http://example.net/)", Text: "http://example.net/)"},
		}),
		lines("automatic_links/mail_url_in_angle_brackets.md", spans{
			md.AutomaticLink{URL: "mailto:someone@example.net", Text: "mailto:someone@example.net"},
			md.AutomaticLink{URL: "mailto:someone@example.net", Text: "someone@example.net"},
		}),
		lines("automatic_links/mail_url_without_angle_brackets.md", spans{
			// NOTE(akavel): below line is unexpected according to
			// testdata/, but from spec this seems totally expected,
			// so I added it
			md.AutomaticLink{URL: "mailto:someone@example.net", Text: "mailto:someone@example.net"},
		}),
		lines("automatic_links/url_schemes.md", spans{
			md.AutomaticLink{URL: "http://example.net", Text: "http://example.net"},
			md.AutomaticLink{URL: "http://example.net", Text: "http://example.net"},
			md.AutomaticLink{URL: "file:///tmp/tmp.html", Text: "file:///tmp/tmp.html"},
			md.AutomaticLink{URL: "file:///tmp/tmp.html", Text: "file:///tmp/tmp.html"},
			md.AutomaticLink{URL: "feed://example.net/rss.xml", Text: "feed://example.net/rss.xml"},
			md.AutomaticLink{URL: "feed://example.net/rss.xml", Text: "feed://example.net/rss.xml"},
			md.AutomaticLink{URL: "googlechrome://example.net/", Text: "googlechrome://example.net/"},
			md.AutomaticLink{URL: "googlechrome://example.net/", Text: "googlechrome://example.net/"},
			md.Code{bb("<>")},
			// NOTE(akavel): below line is unexpected according to
			// testdata/, but from spec this seems totally expected,
			// so I added it
			md.AutomaticLink{URL: "mailto:me@example.net", Text: "mailto:me@example.net"},
			md.AutomaticLink{URL: "mailto:me@example.net", Text: "mailto:me@example.net"},
		}),
		lines("automatic_links/url_special_chars.md", spans{
			md.AutomaticLink{URL: `http://example.net/*#$%^&\~/blah`, Text: `http://example.net/*#$%^&\~/blah`},
			md.AutomaticLink{URL: `http://example.net/*#$%^&\~)/blah`, Text: `http://example.net/*#$%^&\~)/blah`},
			// NOTE(akavel): testdata expects below commented entry,
			// but this seems wrong compared to spec; I've added
			// fixed entry
			// {bb(`http://example.net/blah/`), md.AutomaticLink{URL: `http://example.net/blah/`, Text: `http://example.net/blah/`}},
			md.AutomaticLink{URL: `http://example.net/blah/*#$%^&\~`, Text: `http://example.net/blah/*#$%^&\~`},
			md.AutomaticLink{URL: `http://example.net/blah/*#$%^&\~)`, Text: `http://example.net/blah/*#$%^&\~)`},
		}),
		lines("automatic_links/web_url_in_angle_brackets.md", spans{
			md.AutomaticLink{URL: "http://example.net/path/", Text: "http://example.net/path/"},
			md.AutomaticLink{URL: "https://example.net/path/", Text: "https://example.net/path/"},
			md.AutomaticLink{URL: "ftp://example.net/path/", Text: "ftp://example.net/path/"},
		}),
		lines("automatic_links/web_url_without_angle_brackets.md", spans{
			md.AutomaticLink{URL: "http://example.net/path/", Text: "http://example.net/path/"},
			md.AutomaticLink{URL: "https://example.net/path/", Text: "https://example.net/path/"},
			md.AutomaticLink{URL: "ftp://example.net/path/", Text: "ftp://example.net/path/"},
		}),
		lines("code/end_of_codespan.md", spans{
			md.Code{bb("code span")},
			md.Code{bb("code span` ends")},
			md.Code{bb("code span`` ends")},
			md.Code{bb("code span`` ``ends")},
			md.Code{bb(`code span\`)},
		}),
		blocks("code/multiline.md", spans{
			md.Code{bb("code span\ncan span multiple\nlines")},
		}),
		lines("code/vs_emph.md", spans{
			md.Code{bb("code containing *em* text")},
			md.Code{bb("code containing **strong** text")},
			md.Code{bb("code containing _em_ text")},
			md.Code{bb("code containing __strong__ text")},

			md.Emphasis{Level: 1},
			md.Code{bb("code")},
			md.End{},
			md.Emphasis{Level: 2},
			md.Code{bb("code")},
			md.End{},
			md.Emphasis{Level: 1},
			md.Code{bb("code")},
			md.End{},
			md.Emphasis{Level: 2},
			md.Code{bb("code")},
			md.End{},

			md.Code{bb("code *intertwined")},
			md.Code{bb("with em* text")},
			md.Code{bb("code **intertwined")},
			md.Code{bb("with strong** text")},
			md.Code{bb("code _intertwined")},
			md.Code{bb("with em_ text")},
			md.Code{bb("code __intertwined")},
			md.Code{bb("with strong__ text")},
		}),
		lines("code/vs_image.md", spans{
			md.Code{bb("code containing ![image](url)")},
			md.Code{bb("code containing ![image][ref]")},
			md.Code{bb("code containing ![ref]")},

			md.Code{bb("containing code")},
			md.Code{bb("containing code")},
			md.Link{ReferenceID: "ref", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Code{bb("containing code")},

			md.Code{bb("code ![intertwined")},
			md.Code{bb("intertwined](with) image")},
			md.Code{bb("code ![intertwined")},
			md.Link{ReferenceID: "ref", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Code{bb("intertwined with][ref] image")},
			md.Code{bb("code ![intertwined")},
			md.Code{bb("with] image")},
		}, head(30)),
		lines("code/vs_link.md", spans{
			md.Code{bb("code containing [link](url)")},
			md.Code{bb("code containing [link][ref]")},
			md.Code{bb("code containing [ref]")},

			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Code{bb("containing code")},

			md.End{},
			md.Link{ReferenceID: "ref", RawEnd: md.Raw{{-1, bb("][ref]")}}},
			md.Code{bb("containing code")},
			md.End{},
			md.Link{ReferenceID: "link `containing code`", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Code{bb("containing code")},
			md.End{},

			md.Code{bb("code [intertwined")},
			md.Code{bb("intertwined](with) link")},
			md.Code{bb("code [intertwined")},
			md.Link{ReferenceID: "ref", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Code{bb("intertwined with][ref] link")},
			md.Code{bb("code [intertwined")},
			md.Code{bb("with] link")},
		}, head(30)),
		lines("code/well_formed.md", spans{
			md.Code{bb("code span")},
			md.Code{bb("code ` span")},
			md.Code{bb("` code span")},
			md.Code{bb("code span `")},
			md.Code{bb("`code span`")},
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
			md.Link{ReferenceID: "_", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			emB("_"), emE("_"), emB("_"), emE("_"),
			// NOTE(akavel): link below not expected in testdata
			// because it's not defined below; but we leave this to
			// user.
			md.Link{ReferenceID: "_", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
		}),
		lines("image/direct_link.md", spans{
			md.Image{AltText: "image", URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Image{AltText: "image", URL: "url", Title: "title", RawEnd: md.Raw{{-1, bb(`](url "title")`)}}},
		}),
		lines("image/direct_link_with_2separating_spaces.md", spans{
			md.Image{AltText: "linking", URL: "/img.png", RawEnd: md.Raw{{-1, bb("]  (/img.png)")}}},
		}),
		blocks("image/direct_link_with_separating_newline.md", spans{
			md.Image{AltText: "link", URL: "/img.png", RawEnd: md.Raw{{-1, bb("]\n(/img.png)")}}},
		}),
		lines("image/direct_link_with_separating_space.md", spans{
			md.Image{AltText: "link", URL: "http://example.net/img.png", RawEnd: md.Raw{{-1, bb("] (http://example.net/img.png)")}}},
		}),
		lines("image/image_title.md", spans{
			md.Image{AltText: "link", URL: "url", Title: `title`, RawEnd: md.Raw{{-1, bb(`](url "title")`)}}},
			md.Image{AltText: "link", URL: "url", Title: `title`, RawEnd: md.Raw{{-1, bb(`](url 'title')`)}}},
			// TODO(akavel): unquote contents of Title when
			// processing? doesn't seem noted in spec, send fix for
			// spec?
			md.Image{AltText: "link", URL: "url", Title: `title 'with' "quotes"`, RawEnd: md.Raw{{-1, bb(`](url "title 'with' \"quotes\"")`)}}},
			md.Image{AltText: "link", URL: "url", Title: `title 'with' "quotes"`, RawEnd: md.Raw{{-1, bb(`](url 'title \'with\' "quotes"')`)}}},
			md.Image{AltText: "link", URL: "url", Title: `title with (brackets)`, RawEnd: md.Raw{{-1, bb(`](url "title with (brackets)")`)}}},
			md.Image{AltText: "link", URL: "url", Title: `title with (brackets)`, RawEnd: md.Raw{{-1, bb(`](url 'title with (brackets)')`)}}},

			md.Image{ReferenceID: "ref id1", AltText: "ref id1", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref id2", AltText: "ref id2", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref id3", AltText: "ref id3", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref id4", AltText: "ref id4", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref id5", AltText: "ref id5", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref id6", AltText: "ref id6", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref id7", AltText: "ref id7", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref id8", AltText: "ref id8", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref id9", AltText: "ref id9", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref id10", AltText: "ref id10", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref id11", AltText: "ref id11", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref id12", AltText: "ref id12", RawEnd: md.Raw{{-1, bb("]")}}},
		}, head(19)),
		lines("image/incomplete.md", spans{
			md.Image{ReferenceID: "ref undefined", AltText: "ref undefined", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref 1", AltText: "ref 1", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref undefined", AltText: "ref undefined", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref 1", AltText: "ref 1", RawEnd: md.Raw{{-1, bb("]")}}},
		}, head(8)),
		blocks("image/link_text_with_newline.md", spans{
			md.Image{AltText: "link\ntext", URL: "url1", RawEnd: md.Raw{{-1, bb("](url1)")}}},
			md.Image{ReferenceID: "ref id", AltText: "ref\nid", RawEnd: md.Raw{{-1, bb("][]")}}},
			md.Image{ReferenceID: "ref id", AltText: "ref\nid", RawEnd: md.Raw{{-1, bb("]")}}},
		}, head(9)),
		lines("image/link_with_parenthesis.md", spans{
			md.Image{URL: "url", AltText: "bad link", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Image{URL: `url\`, AltText: "bad link", RawEnd: md.Raw{{-1, bb(`](url\)`)}}},
			md.Image{URL: `url)`, AltText: "link", Title: "title", RawEnd: md.Raw{{-1, bb(`](<url)> "title")`)}}},
		}),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// lines("image/multiple_ref_id_definitions.md", spans{}),
		lines("image/nested_images.md", spans{
			md.Image{AltText: "link2", URL: "url2", RawEnd: md.Raw{{-1, bb("](url2)")}}},
		}),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// lines("image/ref_case_sensitivity.md", spans{}),
		blocks("image/ref_id_matching.md", spans{
			md.Image{ReferenceID: "ref id", AltText: "link", RawEnd: md.Raw{{-1, bb("][ref id]")}}},
			md.Image{ReferenceID: "ref id", AltText: "link", RawEnd: md.Raw{{-1, bb("][ref   id]")}}},
			md.Image{ReferenceID: "ref id", AltText: "link", RawEnd: md.Raw{{-1, bb("][  ref id  ]")}}},
			md.Image{ReferenceID: "ref id", AltText: "link", RawEnd: md.Raw{{-1, bb("][ref\n   id]")}}},
			md.Image{ReferenceID: "ref id", AltText: "ref id", RawEnd: md.Raw{{-1, bb("][]")}}},
			md.Image{ReferenceID: "ref id", AltText: "ref   id", RawEnd: md.Raw{{-1, bb("][]")}}},
			md.Image{ReferenceID: "ref id", AltText: "  ref id  ", RawEnd: md.Raw{{-1, bb("][]")}}},
			md.Image{ReferenceID: "ref id", AltText: "ref\n   id", RawEnd: md.Raw{{-1, bb("][]")}}},
			md.Image{ReferenceID: "ref id", AltText: "ref id", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref id", AltText: "ref   id", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref id", AltText: "  ref id  ", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "ref id", AltText: "ref\n   id", RawEnd: md.Raw{{-1, bb("]")}}},
		}, head(18)),
		// NOTE(akavel): below tests are not really interesting for us
		// here now.
		// lines("image/ref_link.md", spans{}),
		// lines("image/ref_link_empty.md", spans{}),
		// lines("image/ref_link_self.md", spans{}),
		lines("image/ref_link_with_2separating_spaces.md", spans{
			md.Image{AltText: "link", ReferenceID: "ref", RawEnd: md.Raw{{-1, bb("]  [ref]")}}},
		}, head(2)),
		blocks("image/ref_link_with_separating_newline.md", spans{
			md.Image{AltText: "link", ReferenceID: "ref", RawEnd: md.Raw{{-1, bb("]\n[ref]")}}},
		}, head(3)),
		lines("image/ref_link_with_separating_space.md", spans{
			md.Image{AltText: "link", ReferenceID: "ref", RawEnd: md.Raw{{-1, bb("] [ref]")}}},
		}, head(2)),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// lines("image/ref_resolution_within_other_blocks.md", spans{}),
		lines("image/square_brackets_in_link_or_ref.md", spans{
			md.Link{ReferenceID: "1", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Link{ReferenceID: "2", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Image{ReferenceID: "2", AltText: "2", RawEnd: md.Raw{{-1, bb("]")}}},
			// TODO(akavel): make sure we handled escaping properly in cases below
			md.Image{URL: "url", AltText: `link[1]`, RawEnd: md.Raw{{-1, bb(`](url)`)}}},
			md.Image{URL: "url", AltText: `link[2]`, RawEnd: md.Raw{{-1, bb(`](url)`)}}},
			md.Image{URL: "url", AltText: `link![2]`, RawEnd: md.Raw{{-1, bb(`](url)`)}}},
			md.Link{ReferenceID: "2", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Image{ReferenceID: "link", AltText: "link", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Link{ReferenceID: "3", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Image{ReferenceID: "link", AltText: "link", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Link{ReferenceID: "4", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			// TODO(akavel): make sure we handled escaping properly in cases below
			md.Image{ReferenceID: `ref\[3\]`, AltText: `link`, RawEnd: md.Raw{{-1, bb(`][ref\[3\]]`)}}},
			md.Image{ReferenceID: `ref\[4\]`, AltText: `link`, RawEnd: md.Raw{{-1, bb(`][ref\[4\]]`)}}},
			md.Image{ReferenceID: "link", AltText: "link", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Link{ReferenceID: "5", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Image{ReferenceID: "ref", AltText: "link", RawEnd: md.Raw{{-1, bb("][ref]")}}},
			md.Image{ReferenceID: `ref\[5`, AltText: `link`, RawEnd: md.Raw{{-1, bb(`][ref\[5]`)}}},
			md.Image{ReferenceID: `ref\]6`, AltText: `link`, RawEnd: md.Raw{{-1, bb(`][ref\]6]`)}}},
		}, head(16)),
		lines("image/two_consecutive_refs.md", spans{
			md.Image{ReferenceID: "two", AltText: "one", RawEnd: md.Raw{{-1, bb("][two]")}}},
			md.Link{ReferenceID: "three", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Image{ReferenceID: "four", AltText: "one", RawEnd: md.Raw{{-1, bb("][four]")}}},
			md.Link{ReferenceID: "three", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
		}, head(4)),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// image/unused_ref.md
		lines("image/url_escapes.md", spans{
			md.Image{AltText: "link", URL: `url\_\:\$\?`, RawEnd: md.Raw{{-1, bb(`](url\_\:\$\?)`)}}},
			md.Image{AltText: "link", URL: `http://g&ouml;&ouml;gle.com`, RawEnd: md.Raw{{-1, bb(`](http://g&ouml;&ouml;gle.com)`)}}},
		}, head(2)),
		lines("image/url_in_angle_brackets.md", spans{
			md.Image{AltText: "link", URL: "url", RawEnd: md.Raw{{-1, bb(`](<url>)`)}}},
			md.Image{AltText: "link", URL: "url(", RawEnd: md.Raw{{-1, bb(`](<url(>)`)}}},
			md.Image{AltText: "link", URL: "url)", RawEnd: md.Raw{{-1, bb(`](<url)>)`)}}},
			md.Image{AltText: "link", URL: "url)", Title: "title", RawEnd: md.Raw{{-1, bb(`](<url)> "title")`)}}},
		}, head(4)),
		lines("image/url_special_chars.md", spans{
			md.Image{AltText: "link", URL: `url*#$%^&\~`, RawEnd: md.Raw{{-1, bb(`](url*#$%^&\~)`)}}},
			md.Image{ReferenceID: "ref id1", AltText: "link", RawEnd: md.Raw{{-1, bb("][ref id1]")}}},
			md.Image{ReferenceID: "ref id1", AltText: "ref id1", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "link", AltText: "link", RawEnd: md.Raw{{-1, bb("]")}}},
		}, head(8)),
		blocks("image/url_whitespace.md", spans{
			md.Image{AltText: "link", ReferenceID: "link", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{AltText: "link", ReferenceID: "link", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{AltText: "link", URL: "url1", RawEnd: md.Raw{{-1, bb("](<url 1>)")}}},
			md.Image{AltText: "link", URL: "url1", RawEnd: md.Raw{{-1, bb("](<url \n   1>)")}}},
		}, head(6)),
		lines("image/vs_code.md", spans{
			md.Code{bb("code")},
			md.Code{bb("containing ![image](url)")},
			md.Code{bb("containing ![image][ref]")},
			md.Link{ReferenceID: "ref", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Code{bb("intertwined](url) with code")},
			md.Code{bb("intertwined ![with code")},
		}),
		lines("image/vs_emph.md", spans{
			md.Image{URL: "url", AltText: "image containing *em* text", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Image{URL: "url", AltText: "image containing **strong** text", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Image{URL: "url", AltText: "image containing _em_ text", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Image{URL: "url", AltText: "image containing __strong__ text", RawEnd: md.Raw{{-1, bb("](url)")}}},

			emB("*"),
			md.Image{AltText: "image", URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}},
			emE("*"),
			emB("**"),
			md.Image{AltText: "image", URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}},
			emE("**"),
			emB("_"),
			md.Image{AltText: "image", URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}},
			emE("_"),
			emB("__"),
			md.Image{AltText: "image", URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}},
			emE("__"),

			md.Image{URL: "url", AltText: "image *intertwined", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Image{URL: "url", AltText: "with em* text", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Image{URL: "url", AltText: "image **intertwined", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Image{URL: "url", AltText: "with strong** text", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Image{URL: "url", AltText: "image _intertwined", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Image{URL: "url", AltText: "with em_ text", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Image{URL: "url", AltText: "image __intertwined", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Image{URL: "url", AltText: "with strong__ text", RawEnd: md.Raw{{-1, bb("](url)")}}},
		}),
		lines("image/within_link.md", spans{
			md.Link{ReferenceID: "![kitten 1]", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "kitten 1", AltText: "kitten 1", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Link{ReferenceID: "![kitten 2]", RawEnd: md.Raw{{-1, bb("]")}}},
			md.Image{ReferenceID: "kitten 2", AltText: "kitten 2", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
		}, head(5)),
		lines("link/direct_link.md", spans{
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.End{},
			md.Link{URL: "url", Title: "title", RawEnd: md.Raw{{-1, bb(`](url "title")`)}}},
			md.End{},
		}),
		lines("link/direct_link_with_2separating_spaces.md", spans{
			md.Link{URL: "/example.html", RawEnd: md.Raw{{-1, bb("]  (/example.html)")}}},
			md.End{},
		}),
		blocks("link/direct_link_with_separating_newline.md", spans{
			md.Link{URL: "/example.html", RawEnd: md.Raw{{-1, bb("]\n(/example.html)")}}},
			md.End{},
		}),
		lines("link/direct_link_with_separating_space.md", spans{
			md.Link{URL: "http://example.net", RawEnd: md.Raw{{-1, bb("] (http://example.net)")}}},
			md.End{},
		}),
		lines("link/incomplete.md", spans{
			md.Link{ReferenceID: "ref undefined", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Link{ReferenceID: "ref 1", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Link{ReferenceID: "ref undefined", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Link{ReferenceID: "ref 1", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
		}, head(8)),
		blocks("link/link_text_with_newline.md", spans{
			md.Link{URL: "url1", RawEnd: md.Raw{{-1, bb("](url1)")}}},
			md.End{},
			md.Link{ReferenceID: "ref id", RawEnd: md.Raw{{-1, bb("][]")}}},
			md.End{},
			md.Link{ReferenceID: "ref id", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
		}, head(9)),
		lines("link/link_title.md", spans{
			md.Link{URL: "url", Title: "title", RawEnd: md.Raw{{-1, bb(`](url "title")`)}}},
			md.End{},
			md.Link{URL: "url", Title: "title", RawEnd: md.Raw{{-1, bb(`](url 'title')`)}}},
			md.End{},
			md.Link{URL: "url", Title: `title 'with' "quotes"`, RawEnd: md.Raw{{-1, bb(`](url "title 'with' \"quotes\"")`)}}},
			md.End{},
			md.Link{URL: "url", Title: `title 'with' "quotes"`, RawEnd: md.Raw{{-1, bb(`](url 'title \'with\' "quotes"')`)}}},
			md.End{},
			md.Link{URL: "url", Title: "title with (brackets)", RawEnd: md.Raw{{-1, bb(`](url "title with (brackets)")`)}}},
			md.End{},
			md.Link{URL: "url", Title: "title with (brackets)", RawEnd: md.Raw{{-1, bb(`](url 'title with (brackets)')`)}}},
			md.End{},
		}, head(6)),
		lines("link/link_with_parenthesis.md", spans{
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.End{},
			md.Link{URL: `url\`, RawEnd: md.Raw{{-1, bb(`](url\)`)}}},
			md.End{},
			md.Link{URL: `url)`, Title: "title", RawEnd: md.Raw{{-1, bb(`](<url)> "title")`)}}},
			md.End{},
		}),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// link/multiple_ref_id_definitions.md
		lines("link/nested_links.md", spans{
			md.Link{URL: "url2", RawEnd: md.Raw{{-1, bb("](url2)")}}},
			md.End{},
		}),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// link/ref_case_sensitivity.md
		blocks("link/ref_id_matching.md", spans{
			md.Link{ReferenceID: "ref id", RawEnd: md.Raw{{-1, bb("][ref id]")}}}, md.End{},
			md.Link{ReferenceID: "ref id", RawEnd: md.Raw{{-1, bb("][ref   id]")}}}, md.End{},
			md.Link{ReferenceID: "ref id", RawEnd: md.Raw{{-1, bb("][  ref id  ]")}}}, md.End{},
			md.Link{ReferenceID: "ref id", RawEnd: md.Raw{{-1, bb("][ref\n   id]")}}}, md.End{},

			md.Link{ReferenceID: "ref id", RawEnd: md.Raw{{-1, bb("][]")}}}, md.End{},
			md.Link{ReferenceID: "ref id", RawEnd: md.Raw{{-1, bb("][]")}}}, md.End{},
			md.Link{ReferenceID: "ref id", RawEnd: md.Raw{{-1, bb("][]")}}}, md.End{},
			md.Link{ReferenceID: "ref id", RawEnd: md.Raw{{-1, bb("][]")}}}, md.End{},

			md.Link{ReferenceID: "ref id", RawEnd: md.Raw{{-1, bb("]")}}}, md.End{},
			md.Link{ReferenceID: "ref id", RawEnd: md.Raw{{-1, bb("]")}}}, md.End{},
			md.Link{ReferenceID: "ref id", RawEnd: md.Raw{{-1, bb("]")}}}, md.End{},
			md.Link{ReferenceID: "ref id", RawEnd: md.Raw{{-1, bb("]")}}}, md.End{},
		}, head(18)),
		// NOTE(akavel): below tests are not really interesting for us
		// here now.
		// link/ref_link.md
		// link/ref_link_empty.md
		// link/ref_link_self.md
		lines("link/ref_link_with_2separating_spaces.md", spans{
			md.Link{ReferenceID: "ref", RawEnd: md.Raw{{-1, bb("]  [ref]")}}},
			md.End{},
		}, head(2)),
		blocks("link/ref_link_with_separating_newline.md", spans{
			md.Link{ReferenceID: "ref", RawEnd: md.Raw{{-1, bb("]\n[ref]")}}},
			md.End{},
		}, head(3)),
		lines("link/ref_link_with_separating_space.md", spans{
			md.Link{ReferenceID: "ref", RawEnd: md.Raw{{-1, bb("] [ref]")}}},
			md.End{},
		}, head(2)),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// link/ref_resolution_within_other_blocks.md
		lines("link/square_brackets_in_link_or_ref.md", spans{
			md.Link{ReferenceID: "1", RawEnd: md.Raw{{-1, bb("]")}}}, md.End{},
			md.Link{ReferenceID: "2", RawEnd: md.Raw{{-1, bb("]")}}}, md.End{},
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}}, md.End{},
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}}, md.End{},
			md.Link{ReferenceID: "link", RawEnd: md.Raw{{-1, bb("]")}}}, md.End{},
			md.Link{ReferenceID: "3", RawEnd: md.Raw{{-1, bb("]")}}}, md.End{},
			md.Link{ReferenceID: "link", RawEnd: md.Raw{{-1, bb("]")}}}, md.End{},
			md.Link{ReferenceID: "4", RawEnd: md.Raw{{-1, bb("]")}}}, md.End{},
			md.Link{ReferenceID: `ref\[3\]`, RawEnd: md.Raw{{-1, bb(`][ref\[3\]]`)}}}, md.End{},
			md.Link{ReferenceID: `ref\[4\]`, RawEnd: md.Raw{{-1, bb(`][ref\[4\]]`)}}}, md.End{},
			md.Link{ReferenceID: "link", RawEnd: md.Raw{{-1, bb("]")}}}, md.End{},
			md.Link{ReferenceID: "5", RawEnd: md.Raw{{-1, bb("]")}}}, md.End{},
			md.Link{ReferenceID: "ref", RawEnd: md.Raw{{-1, bb("][ref]")}}}, md.End{},
			md.Link{ReferenceID: `ref\[5`, RawEnd: md.Raw{{-1, bb(`][ref\[5]`)}}}, md.End{},
			md.Link{ReferenceID: `ref\]6`, RawEnd: md.Raw{{-1, bb(`][ref\]6]`)}}}, md.End{},
		}, head(13)),
		lines("link/two_consecutive_refs.md", spans{
			md.Link{ReferenceID: `two`, RawEnd: md.Raw{{-1, bb(`][two]`)}}}, md.End{},
			md.Link{ReferenceID: `three`, RawEnd: md.Raw{{-1, bb(`]`)}}}, md.End{},
			md.Link{ReferenceID: `four`, RawEnd: md.Raw{{-1, bb(`][four]`)}}}, md.End{},
			md.Link{ReferenceID: `three`, RawEnd: md.Raw{{-1, bb(`]`)}}}, md.End{},
		}, head(4)),
		// NOTE(akavel): below test is not really interesting for us
		// here now.
		// link/unused_ref.md
		lines("link/url_escapes.md", spans{
			// TODO(akavel): make sure we handled escaping properly in cases below
			md.Link{URL: `url\_\:\$\?`, RawEnd: md.Raw{{-1, bb(`](url\_\:\$\?)`)}}}, md.End{},
			md.Link{URL: `http://g&ouml;&ouml;gle.com`, RawEnd: md.Raw{{-1, bb(`](http://g&ouml;&ouml;gle.com)`)}}}, md.End{},
		}, head(2)),
		lines("link/url_in_angle_brackets.md", spans{
			md.Link{URL: `url`, RawEnd: md.Raw{{-1, bb(`](<url>)`)}}}, md.End{},
			md.Link{URL: `url(`, RawEnd: md.Raw{{-1, bb(`](<url(>)`)}}}, md.End{},
			md.Link{URL: `url)`, RawEnd: md.Raw{{-1, bb(`](<url)>)`)}}}, md.End{},
			md.Link{URL: `url)`, Title: "title", RawEnd: md.Raw{{-1, bb(`](<url)> "title")`)}}}, md.End{},
		}, head(4)),
		lines("link/url_special_chars.md", spans{
			md.Link{URL: `url*#$%^&\~`, RawEnd: md.Raw{{-1, bb(`](url*#$%^&\~)`)}}}, md.End{},
			md.Link{ReferenceID: "ref id1", RawEnd: md.Raw{{-1, bb(`][ref id1]`)}}}, md.End{},
			md.Link{ReferenceID: "ref id1", RawEnd: md.Raw{{-1, bb(`]`)}}}, md.End{},
			md.Link{ReferenceID: "link", RawEnd: md.Raw{{-1, bb(`]`)}}}, md.End{},
		}, head(8)),
		blocks("link/url_whitespace.md", spans{
			md.Link{ReferenceID: "link", RawEnd: md.Raw{{-1, bb(`]`)}}}, md.End{},
			md.Link{ReferenceID: "link", RawEnd: md.Raw{{-1, bb(`]`)}}}, md.End{},
			md.Link{URL: "url1", RawEnd: md.Raw{{-1, bb("](<url 1>)")}}}, md.End{},
			md.Link{URL: "url1", RawEnd: md.Raw{{-1, bb("](<url \n   1>)")}}}, md.End{},
		}, head(6)),
		lines("link/vs_code.md", spans{
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Code{bb("code")},
			md.End{},
			md.Code{bb("containing [link](url)")},
			md.Code{bb("containing [link][ref]")},
			md.Link{ReferenceID: "ref", RawEnd: md.Raw{{-1, bb("]")}}},
			md.End{},
			md.Code{bb("intertwined](url) with code")},
			md.Code{bb("intertwined [with code")},
		}),
		lines("link/vs_emph.md", spans{
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}},
			emB("*"), emE("*"),
			md.End{},
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}},
			emB("**"), emE("**"),
			md.End{},
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}},
			emB("_"), emE("_"),
			md.End{},
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}},
			emB("__"), emE("__"),
			md.End{},

			emB("*"),
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}}, md.End{},
			emE("*"),
			emB("**"),
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}}, md.End{},
			emE("**"),
			emB("_"),
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}}, md.End{},
			emE("_"),
			emB("__"),
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}}, md.End{},
			emE("__"),

			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}}, md.End{},
			emB("*"), emE("*"),
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}}, md.End{},
			emB("**"), emE("**"),
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}}, md.End{},
			emB("_"), emE("_"),
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}}, md.End{},
			emB("__"), emE("__"),
		}),
		lines("link/vs_image.md", spans{
			md.Link{URL: "url", RawEnd: md.Raw{{-1, bb("](url)")}}},
			md.Image{AltText: "image", URL: "/image.jpg", RawEnd: md.Raw{{-1, bb("](/image.jpg)")}}},
			md.End{},
		}),
	}
	for _, c := range cases {
		fmt.Printf("\ncase %s\n", c.fname)
		tags := []md.Tag{}
		for _, b := range c.blocks {
			tags = append(tags, Parse(b, nil)...)
		}
		spans := []md.Tag{}
		for _, t := range tags {
			if _, ok := t.(md.Prose); !ok {
				spans = append(spans, t)
			}
		}
		if !reflect.DeepEqual(c.spans, spans) {
			test.Errorf("case %s expected vs. got DIFF:\n%s",
				c.fname,
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
