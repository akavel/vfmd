package block_test

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kylelemons/godebug/diff"

	"gopkg.in/akavel/vfmd.v0"
	. "gopkg.in/akavel/vfmd.v0/block"
	"gopkg.in/akavel/vfmd.v0/md"
)

/*
TODO(akavel): missing tests:
{"block_level/atx_header/span_in_text.md"},
{"block_level/paragraph/blanks_within_html_comment.md"},
{"block_level/paragraph/blanks_within_html_tag.md"},
{"block_level/paragraph/blanks_within_verbatim_html.md"},
{"block_level/paragraph/html_block.md"},
{"block_level/paragraph/html_comment.md"},
{"block_level/paragraph/md_within_html.md"},
{"block_level/paragraph/misnested_html.md"},
{"block_level/paragraph/non_phrasing_html_tag.md"},
{"block_level/paragraph/phrasing_html_tag.md"},
{"block_level/setext_header/span_in_text.md"},

// {"span_level/automatic_links/angle_brackets_in_link.md"},
// {"span_level/automatic_links/mail_url_without_angle_brackets.md"},
// {"span_level/automatic_links/url_schemes.md"},
// {"span_level/automatic_links/url_special_chars.md"},
// {"span_level/code/vs_html.md"},
{"span_level/code/well_formed.md"},
{"span_level/emphasis/emphasis_tag_combinations.md"},
{"span_level/emphasis/intertwined.md"},
{"span_level/emphasis/intraword.md"},
{"span_level/emphasis/nested_homogenous.md"},
{"span_level/emphasis/opening_and_closing_tags.md"},
{"span_level/emphasis/simple.md"},
{"span_level/emphasis/vs_html.md"},
{"span_level/emphasis/within_whitespace.md"},
// {"span_level/emphasis/with_punctuation.md"},
{"span_level/image/direct_link.md"},
{"span_level/image/direct_link_with_2separating_spaces.md"},
{"span_level/image/direct_link_with_separating_newline.md"},
{"span_level/image/direct_link_with_separating_space.md"},
// {"span_level/image/image_title.md"},
{"span_level/image/incomplete.md"},
{"span_level/image/link_text_with_newline.md"},
{"span_level/image/link_with_parenthesis.md"},
{"span_level/image/multiple_ref_id_definitions.md"},
{"span_level/image/nested_images.md"},
{"span_level/image/ref_case_sensitivity.md"},
{"span_level/image/ref_id_matching.md"},
{"span_level/image/ref_link.md"},
{"span_level/image/ref_link_empty.md"},
{"span_level/image/ref_link_self.md"},
{"span_level/image/ref_link_with_2separating_spaces.md"},
{"span_level/image/ref_link_with_separating_newline.md"},
{"span_level/image/ref_link_with_separating_space.md"},
{"span_level/image/ref_resolution_within_other_blocks.md"},
{"span_level/image/square_brackets_in_link_or_ref.md"},
{"span_level/image/two_consecutive_refs.md"},
{"span_level/image/unused_ref.md"},
{"span_level/image/url_escapes.md"},
{"span_level/image/url_in_angle_brackets.md"},
{"span_level/image/url_special_chars.md"},
{"span_level/image/url_whitespace.md"},
{"span_level/image/vs_code.md"},
{"span_level/image/vs_emph.md"},
{"span_level/image/vs_html.md"},
{"span_level/image/within_link.md"},
{"span_level/link/direct_link.md"},
{"span_level/link/direct_link_with_2separating_spaces.md"},
{"span_level/link/direct_link_with_separating_newline.md"},
{"span_level/link/direct_link_with_separating_space.md"},
{"span_level/link/incomplete.md"},
{"span_level/link/link_text_with_newline.md"},
{"span_level/link/link_title.md"},
{"span_level/link/link_with_parenthesis.md"},
{"span_level/link/multiple_ref_id_definitions.md"},
{"span_level/link/nested_links.md"},
{"span_level/link/ref_case_sensitivity.md"},
{"span_level/link/ref_id_matching.md"},
{"span_level/link/ref_link.md"},
{"span_level/link/ref_link_empty.md"},
{"span_level/link/ref_link_self.md"},
{"span_level/link/ref_link_with_2separating_spaces.md"},
{"span_level/link/ref_link_with_separating_newline.md"},
{"span_level/link/ref_link_with_separating_space.md"},
{"span_level/link/ref_resolution_within_other_blocks.md"},
{"span_level/link/square_brackets_in_link_or_ref.md"},
{"span_level/link/two_consecutive_refs.md"},
{"span_level/link/unused_ref.md"},
{"span_level/link/url_escapes.md"},
{"span_level/link/url_in_angle_brackets.md"},
{"span_level/link/url_special_chars.md"},
{"span_level/link/url_whitespace.md"},
{"span_level/link/vs_code.md"},
{"span_level/link/vs_emph.md"},
{"span_level/link/vs_html.md"},
{"span_level/link/vs_image.md"},
*/

func TestHTMLFiles(test *testing.T) {
	const dir = "../testdata/tests"
	cases := []struct {
		path string
	}{
		{"block_level/atx_header/blank_text.md"},
		{"block_level/atx_header/enclosed_blank_text.md"},
		{"block_level/atx_header/hash_in_text.md"},
		{"block_level/atx_header/left_only.md"},
		{"block_level/atx_header/left_right.md"},
		{"block_level/atx_header/more_than_six_hashes.md"},
		{"block_level/atx_header/space_in_text.md"},
		{"block_level/blockquote/containing_atx_header.md"},
		{"block_level/blockquote/containing_blockquote.md"},
		{"block_level/blockquote/containing_codeblock.md"},
		{"block_level/blockquote/containing_hr.md"},
		{"block_level/blockquote/containing_list.md"},
		{"block_level/blockquote/containing_setext_header.md"},
		{"block_level/blockquote/followed_by_atx_header.md"},
		{"block_level/blockquote/followed_by_codeblock.md"},
		{"block_level/blockquote/followed_by_hr.md"},
		{"block_level/blockquote/followed_by_list.md"},
		{"block_level/blockquote/followed_by_para.md"},
		{"block_level/blockquote/followed_by_setext_header.md"},
		{"block_level/blockquote/indented_differently1.md"},
		{"block_level/blockquote/indented_differently2.md"},
		{"block_level/blockquote/many_level_nesting.md"},
		{"block_level/blockquote/many_lines.md"},
		{"block_level/blockquote/many_lines_lazy.md"},
		{"block_level/blockquote/many_paras.md"},
		{"block_level/blockquote/many_paras_2blank.md"},
		{"block_level/blockquote/many_paras_2blank_lazy.md"},
		{"block_level/blockquote/many_paras_2blank_lazy2.md"},
		{"block_level/blockquote/many_paras_lazy.md"},
		{"block_level/blockquote/many_paras_lazy2.md"},
		{"block_level/blockquote/no_space_after_gt.md"},
		{"block_level/blockquote/one_line.md"},
		{"block_level/blockquote/space_before_gt.md"},
		{"block_level/codeblock/followed_by_para.md"},
		{"block_level/codeblock/html_escaping.md"},
		{"block_level/codeblock/many_lines.md"},
		{"block_level/codeblock/more_than_four_leading_space.md"},
		{"block_level/codeblock/one_blank_line_bw_codeblocks.md"},
		{"block_level/codeblock/one_line.md"},
		{"block_level/codeblock/two_blank_lines_bw_codeblocks.md"},
		{"block_level/codeblock/vs_atx_header.md"},
		{"block_level/codeblock/vs_blockquote.md"},
		{"block_level/codeblock/vs_hr.md"},
		{"block_level/codeblock/vs_list.md"},
		{"block_level/horizontal_rule/end_with_space.md"},
		{"block_level/horizontal_rule/followed_by_block.md"},
		{"block_level/horizontal_rule/loose.md"},
		{"block_level/horizontal_rule/sparse.md"},
		{"block_level/horizontal_rule/start_with_space.md"},
		{"block_level/horizontal_rule/tight.md"},
		{"block_level/ordered_list/all_items_loose.md"},
		{"block_level/ordered_list/all_items_tight.md"},
		{"block_level/ordered_list/all_items_tight_even_with_blanks.md"},
		{"block_level/ordered_list/at_end_of_parent_without_blank_line.md"},
		{"block_level/ordered_list/bw_unordered_lists.md"},
		{"block_level/ordered_list/followed_by_hr.md"},
		{"block_level/ordered_list/followed_by_list.md"},
		{"block_level/ordered_list/indent_and_sub_blocks.md"},
		{"block_level/ordered_list/list_ends_with_2blanks.md"},
		{"block_level/ordered_list/many_level_nesting.md"},
		{"block_level/ordered_list/many_lines.md"},
		{"block_level/ordered_list/many_lines_lazy.md"},
		{"block_level/ordered_list/many_paras.md"},
		{"block_level/ordered_list/many_paras_2blank.md"},
		{"block_level/ordered_list/many_paras_2blank_lazy.md"},
		{"block_level/ordered_list/many_paras_lazy.md"},
		{"block_level/ordered_list/no_space_after_number.md"},
		{"block_level/ordered_list/no_space_before_number.md"},
		{"block_level/ordered_list/numbering_from_two.md"},
		{"block_level/ordered_list/numbering_not_in_order.md"},
		{"block_level/ordered_list/numbers_left_aligned.md"},
		{"block_level/ordered_list/numbers_right_aligned.md"},
		{"block_level/ordered_list/numbers_wiggly.md"},
		{"block_level/ordered_list/one_line.md"},
		{"block_level/ordered_list/some_items_loose.md"},
		{"block_level/ordered_list/space_before_number.md"},
		{"block_level/ordered_list/three_paras_loose.md"},
		{"block_level/ordered_list/three_paras_tight.md"},
		{"block_level/ordered_list/two_paras_loose.md"},
		{"block_level/ordered_list/with_atx_header.md"},
		{"block_level/ordered_list/with_blockquote.md"},
		{"block_level/ordered_list/with_codeblock.md"},
		{"block_level/ordered_list/with_para.md"},
		{"block_level/ordered_list/with_setext_header.md"},
		{"block_level/paragraph/followed_by_atx_header.md"},
		{"block_level/paragraph/followed_by_blockquote.md"},
		{"block_level/paragraph/followed_by_codeblock.md"},
		{"block_level/paragraph/followed_by_horizontal_rule.md"},
		{"block_level/paragraph/followed_by_list.md"},
		{"block_level/paragraph/followed_by_setext_header.md"},
		{"block_level/paragraph/simple_para.md"},
		{"block_level/paragraph/two_paras_1blank.md"},
		{"block_level/paragraph/two_paras_2blank.md"},
		{"block_level/setext_header/blank_text.md"},
		{"block_level/setext_header/enclosed_space_in_underline.md"},
		{"block_level/setext_header/leading_space_in_text.md"},
		{"block_level/setext_header/leading_space_in_underline.md"},
		{"block_level/setext_header/simple.md"},
		{"block_level/setext_header/trailing_space_in_underline.md"},
		{"block_level/setext_header/vs_atx_header.md"},
		{"block_level/setext_header/vs_blockquote.md"},
		{"block_level/setext_header/vs_codeblock.md"},
		{"block_level/setext_header/vs_list.md"},
		{"block_level/unordered_list/all_items_loose.md"},
		{"block_level/unordered_list/all_items_tight.md"},
		{"block_level/unordered_list/all_items_tight_even_with_blanks.md"},
		{"block_level/unordered_list/at_end_of_parent_without_blank_line.md"},
		{"block_level/unordered_list/bw_ordered_lists.md"},
		{"block_level/unordered_list/changing_bullet.md"},
		{"block_level/unordered_list/changing_starter_string.md"},
		{"block_level/unordered_list/different_bullet_chars.md"},
		{"block_level/unordered_list/followed_by_hr.md"},
		{"block_level/unordered_list/followed_by_list.md"},
		{"block_level/unordered_list/indent_and_sub_blocks.md"},
		{"block_level/unordered_list/list_ends_with_2blanks.md"},
		{"block_level/unordered_list/many_level_nesting.md"},
		{"block_level/unordered_list/many_lines.md"},
		{"block_level/unordered_list/many_lines_lazy.md"},
		{"block_level/unordered_list/many_paras.md"},
		{"block_level/unordered_list/many_paras_2blank.md"},
		{"block_level/unordered_list/many_paras_2blank_lazy.md"},
		{"block_level/unordered_list/many_paras_lazy.md"},
		{"block_level/unordered_list/no_space_after_bullet.md"},
		{"block_level/unordered_list/no_space_before_bullet.md"},
		{"block_level/unordered_list/one_line.md"},
		{"block_level/unordered_list/some_items_loose.md"},
		{"block_level/unordered_list/space_before_bullet.md"},
		{"block_level/unordered_list/three_paras_loose.md"},
		{"block_level/unordered_list/three_paras_tight.md"},
		{"block_level/unordered_list/two_paras_loose.md"},
		{"block_level/unordered_list/with_atx_header.md"},
		{"block_level/unordered_list/with_blockquote.md"},
		{"block_level/unordered_list/with_codeblock.md"},
		{"block_level/unordered_list/with_para.md"},
		{"block_level/unordered_list/with_setext_header.md"},

		{"span_level/automatic_links/ending_with_punctuation.md"},
		{"span_level/automatic_links/mail_url_in_angle_brackets.md"},
		{"span_level/automatic_links/web_url_in_angle_brackets.md"},
		{"span_level/automatic_links/web_url_without_angle_brackets.md"},
		{"span_level/code/end_of_codespan.md"},
		{"span_level/code/multiline.md"},
		{"span_level/code/vs_emph.md"},
		{"span_level/code/vs_image.md"},
		{"span_level/code/vs_link.md"},
	}

	// Patches to what I believe are bugs in the original testdata, when
	// confronted with the spec.
	replacer := strings.NewReplacer(
		"'&gt;'", "&#39;&gt;&#39;",
		"\n<li>\n    Parent list\n\n    <ol>", "\n<li>Parent list<ol>",
		"<code>Code block included in list\n</code>", "<code> Code block included in list\n</code>",
		"And another\n\n<p>Another para", "And another<p>Another para",
		"<li>Level 1\n<ol>", "<li>Level 1<ol>",
		"<li>Level 2\n<p>", "<li>Level 2<p>",
		"<li>Level 3\n<p>", "<li>Level 3<p>",
		"Level 4\n<ul>", "Level 4<ul>",
		"<li>Level 2\n<ol>", "<li>Level 2<ol>",
		"And another\n<p>", "And another<p>",
		"<li>Parent list\n    <ul>", "<li>Parent list<ul>",
		"<li>Third list\n<ul>", "<li>Third list<ul>",
		"<pre><code>Code block included in list</code></pre>", "<pre><code>Code block included in list\n</code></pre>",
		"<pre><code>Code block not included in list</code></pre>", "<pre><code>Code block not included in list\n</code></pre>",
		"<li>Level 1\n<ul>", "<li>Level 1<ul>",
		"<li>Level 2\n<ul>", "<li>Level 2<ul>",
	)

	for _, c := range cases {
		test.Log(c.path)
		subdir, fname := path.Split(c.path)
		fname = strings.TrimSuffix(fname, ".md")
		data, err := ioutil.ReadFile(filepath.Join(dir, c.path))
		if err != nil {
			test.Error(err)
			continue
		}
		expectedOutput, err := ioutil.ReadFile(
			filepath.Join(dir, subdir, "expected", fname+".html"))
		if err != nil {
			test.Error(err)
			continue
		}
		prep, _ := vfmd.QuickPrep(bytes.NewReader(data))
		blocks, err := QuickParse(bytes.NewReader(prep), BlocksAndSpans, nil, nil)
		if err != nil {
			test.Errorf("case %s error: %s", c.path, err)
			continue
		}

		html := quickHtml(blocks)
		html = simplifyHtml(html)
		expectedOutput = []byte(replacer.Replace(string(simplifyHtml(expectedOutput))))
		if !bytes.Equal(html, expectedOutput) {
			test.Errorf("case %s blocks:\n%s",
				c.path, spew.Sdump(blocks))
			test.Errorf("case %s expected vs. got DIFF:\n%s",
				c.path, diff.Diff(string(expectedOutput), string(html)))
		}
	}
}

func quickHtml(blocks []md.Tag) []byte {
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
			}[t.Level])
			tags, err = htmlSpans(tags[1:], w, opt)
			fmt.Fprint(w, map[int]string{
				1: "</em>",
				2: "</strong>",
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
				ref, found = opt.refs[strings.ToLower(t.ReferenceID)]
			}
			if found {
				// FIXME(akavel): fully correct escaping
				// TODO(akavel): ref.Title
				fmt.Fprintf(w, `<a href="%s">`, ref.URL)
			} else {
				fmt.Fprintf(w, `[`)
			}
			tags, err = htmlSpans(tags[1:], w, opt)
			if found {
				fmt.Fprintf(w, `</a>`)
			} else {
				// TODO(akavel): fmt.Fprintf(w, t.RawEnd.String())
			}

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
