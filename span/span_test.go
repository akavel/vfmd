package span

import (
	"bytes"
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
	}
	for _, c := range cases {
		spans := []Span{}
		for _, b := range c.blocks {
			spans = append(spans, Process(b, nil)...)
		}
		if !reflect.DeepEqual(c.spans, spans) {
			test.Errorf("case %s expected:\n%s\ngot:",
				c.fname, spew.Sdump(c.spans))
			for i, span := range spans {
				off, err := span.OffsetIn(c.buf)
				test.Errorf("[%d] @ %d [%v]: %s",
					i, off, err, spew.Sdump(span))
			}
		}
	}
}

/*
in ROOT/testdata/tests/span_level:

automatic_links/mail_url_in_angle_brackets.md
automatic_links/mail_url_without_angle_brackets.md
automatic_links/url_schemes.md
automatic_links/url_special_chars.md
automatic_links/web_url_in_angle_brackets.md
automatic_links/web_url_without_angle_brackets.md
code/end_of_codespan.md
code/multiline.md
code/vs_emph.md
code/vs_html.md
code/vs_image.md
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
