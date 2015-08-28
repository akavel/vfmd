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

func lines(filename string, spans spans) spanCase {
	buf, err := ioutil.ReadFile(dir + filename)
	if err != nil {
		panic(err)
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
func blocks(filename string, spans spans) spanCase {
	buf, err := ioutil.ReadFile(dir + filename)
	if err != nil {
		panic(err)
	}
	buf = bytes.Replace(buf, bb("\r"), bb(""), -1)
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

func TestSpan(test *testing.T) {
	cases := []spanCase{
		lines(`automatic_links/angle_brackets_in_link.md`, spans{
			{bb("http://exampl"), AutoLink{URL: `http://exampl`, Text: `http://exampl`}},
			{bb("http://exampl"), AutoLink{URL: `http://exampl`, Text: `http://exampl`}},
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
			{bb("["), LinkBegin{ReferenceID: "ref"}},
			{bb("]"), LinkEnd{}},
			{bb("`containing code`"), Code{bb("containing code")}},

			{bb("`code ![intertwined`"), Code{bb("code ![intertwined")}},
			{bb("`intertwined](with) image`"), Code{bb("intertwined](with) image")}},
			{bb("`code ![intertwined`"), Code{bb("code ![intertwined")}},
			{bb("["), LinkBegin{ReferenceID: "ref"}},
			{bb("]"), LinkEnd{}},
			{bb("`intertwined with][ref] image`"), Code{bb("intertwined with][ref] image")}},
			{bb("`code ![intertwined`"), Code{bb("code ![intertwined")}},
			{bb("`with] image`"), Code{bb("with] image")}},

			// NOTE(akavel): below are accidental because of testing
			// approach; they're irrelevant to algorithm correctness
			{bb("["), LinkBegin{ReferenceID: "ref"}},
			{bb("]"), LinkEnd{}},
			{bb("["), LinkBegin{ReferenceID: "image `containing code`"}},
			{bb("`containing code`"), Code{bb("containing code")}},
			{bb("]"), LinkEnd{}},
			{bb("["), LinkBegin{ReferenceID: "intertwined`with"}},
			{bb("]"), LinkEnd{}},
		}),
	}
	for _, c := range cases {
		spans := []Span{}
		for _, b := range c.blocks {
			spans = append(spans, Process(b, nil)...)
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
			test.Errorf("QUICK DIFF: %s\n", diff(c.spans, spans))
		}
	}
}

/*
in ROOT/testdata/tests/span_level:

code/vs_html.md
code/vs_link.md
code/well_formed.md
emphasis/emphasis_tag_combinations.md
emphasis/intertwined.md
emphasis/intraword.md
emphasis/nested_homogenous.md
emphasis/opening_and_closing_tags.md
emphasis/simple.md
emphasis/vs_html.md
emphasis/within_whitespace.md
emphasis/with_punctuation.md
image/direct_link.md
image/direct_link_with_2separating_spaces.md
image/direct_link_with_separating_newline.md
image/direct_link_with_separating_space.md
image/image_title.md
image/incomplete.md
image/link_text_with_newline.md
image/link_with_parenthesis.md
image/multiple_ref_id_definitions.md
image/nested_images.md
image/ref_case_sensitivity.md
image/ref_id_matching.md
image/ref_link.md
image/ref_link_empty.md
image/ref_link_self.md
image/ref_link_with_2separating_spaces.md
image/ref_link_with_separating_newline.md
image/ref_link_with_separating_space.md
image/ref_resolution_within_other_blocks.md
image/square_brackets_in_link_or_ref.md
image/two_consecutive_refs.md
image/unused_ref.md
image/url_escapes.md
image/url_in_angle_brackets.md
image/url_special_chars.md
image/url_whitespace.md
image/vs_code.md
image/vs_emph.md
image/vs_html.md
image/within_link.md
link/direct_link.md
link/direct_link_with_2separating_spaces.md
link/direct_link_with_separating_newline.md
link/direct_link_with_separating_space.md
link/incomplete.md
link/link_text_with_newline.md
link/link_title.md
link/link_with_parenthesis.md
link/multiple_ref_id_definitions.md
link/nested_links.md
link/ref_case_sensitivity.md
link/ref_id_matching.md
link/ref_link.md
link/ref_link_empty.md
link/ref_link_self.md
link/ref_link_with_2separating_spaces.md
link/ref_link_with_separating_newline.md
link/ref_link_with_separating_space.md
link/ref_resolution_within_other_blocks.md
link/square_brackets_in_link_or_ref.md
link/two_consecutive_refs.md
link/unused_ref.md
link/url_escapes.md
link/url_in_angle_brackets.md
link/url_special_chars.md
link/url_whitespace.md
link/vs_code.md
link/vs_emph.md
link/vs_html.md
link/vs_image.md
*/
