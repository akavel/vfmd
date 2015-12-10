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
paragraph/blanks_within_html_comment.md
paragraph/blanks_within_html_tag.md
paragraph/blanks_within_verbatim_html.md
paragraph/html_block.md
paragraph/html_comment.md
paragraph/md_within_html.md
paragraph/misnested_html.md
paragraph/non_phrasing_html_tag.md
paragraph/phrasing_html_tag.md
*/

func TestHTMLFiles(test *testing.T) {
	cases := []struct {
		path string
	}{
		{"atx_header/blank_text.md"},
		{"atx_header/enclosed_blank_text.md"},
		{"atx_header/hash_in_text.md"},
		{"atx_header/left_only.md"},
		{"atx_header/left_right.md"},
		{"atx_header/more_than_six_hashes.md"},
		{"atx_header/space_in_text.md"},
		{"atx_header/span_in_text.md"},
		{"blockquote/containing_atx_header.md"},
		{"blockquote/containing_blockquote.md"},
		{"blockquote/containing_codeblock.md"},
		{"blockquote/containing_hr.md"},
		{"blockquote/containing_list.md"},
		{"blockquote/containing_setext_header.md"},
		{"blockquote/followed_by_atx_header.md"},
		{"blockquote/followed_by_codeblock.md"},
		{"blockquote/followed_by_hr.md"},
		{"blockquote/followed_by_list.md"},
		{"blockquote/followed_by_para.md"},
		{"blockquote/followed_by_setext_header.md"},
		{"blockquote/indented_differently1.md"},
		{"blockquote/indented_differently2.md"},
		{"blockquote/many_level_nesting.md"},
		{"blockquote/many_lines.md"},
		{"blockquote/many_lines_lazy.md"},
		{"blockquote/many_paras.md"},
		{"blockquote/many_paras_2blank.md"},
		{"blockquote/many_paras_2blank_lazy.md"},
		{"blockquote/many_paras_2blank_lazy2.md"},
		{"blockquote/many_paras_lazy.md"},
		{"blockquote/many_paras_lazy2.md"},
		{"blockquote/no_space_after_gt.md"},
		{"blockquote/one_line.md"},
		{"blockquote/space_before_gt.md"},
		{"codeblock/followed_by_para.md"},
		{"codeblock/html_escaping.md"},
		{"codeblock/many_lines.md"},
		{"codeblock/more_than_four_leading_space.md"},
		{"codeblock/one_blank_line_bw_codeblocks.md"},
		{"codeblock/one_line.md"},
		{"codeblock/two_blank_lines_bw_codeblocks.md"},
		{"codeblock/vs_atx_header.md"},
		{"codeblock/vs_blockquote.md"},
		{"codeblock/vs_hr.md"},
		{"codeblock/vs_list.md"},
		{"horizontal_rule/end_with_space.md"},
		{"horizontal_rule/followed_by_block.md"},
		{"horizontal_rule/loose.md"},
		{"horizontal_rule/sparse.md"},
		{"horizontal_rule/start_with_space.md"},
		{"horizontal_rule/tight.md"},
		{"ordered_list/all_items_loose.md"},
		{"ordered_list/all_items_tight.md"},
		{"ordered_list/all_items_tight_even_with_blanks.md"},
		{"ordered_list/at_end_of_parent_without_blank_line.md"},
		{"ordered_list/bw_unordered_lists.md"},
		{"ordered_list/followed_by_hr.md"},
		{"ordered_list/followed_by_list.md"},
		{"ordered_list/indent_and_sub_blocks.md"},
		{"ordered_list/list_ends_with_2blanks.md"},
		{"ordered_list/many_level_nesting.md"},
		{"ordered_list/many_lines.md"},
		{"ordered_list/many_lines_lazy.md"},
		{"ordered_list/many_paras.md"},
		{"ordered_list/many_paras_2blank.md"},
		{"ordered_list/many_paras_2blank_lazy.md"},
		{"ordered_list/many_paras_lazy.md"},
		{"ordered_list/no_space_after_number.md"},
		{"ordered_list/no_space_before_number.md"},
		{"ordered_list/numbering_from_two.md"},
		{"ordered_list/numbering_not_in_order.md"},
		{"ordered_list/numbers_left_aligned.md"},
		{"ordered_list/numbers_right_aligned.md"},
		{"ordered_list/numbers_wiggly.md"},
		{"ordered_list/one_line.md"},
		{"ordered_list/some_items_loose.md"},
		{"ordered_list/space_before_number.md"},
		{"ordered_list/three_paras_loose.md"},
		{"ordered_list/three_paras_tight.md"},
		{"ordered_list/two_paras_loose.md"},
		{"ordered_list/with_atx_header.md"},
		{"ordered_list/with_blockquote.md"},
		{"ordered_list/with_codeblock.md"},
		{"ordered_list/with_para.md"},
		{"ordered_list/with_setext_header.md"},
		{"paragraph/followed_by_atx_header.md"},
		{"paragraph/followed_by_blockquote.md"},
		{"paragraph/followed_by_codeblock.md"},
		{"paragraph/followed_by_horizontal_rule.md"},
		{"paragraph/followed_by_list.md"},
		{"paragraph/followed_by_setext_header.md"},
		{"paragraph/simple_para.md"},
		{"paragraph/two_paras_1blank.md"},
		{"paragraph/two_paras_2blank.md"},
		{"setext_header/blank_text.md"},
		{"setext_header/enclosed_space_in_underline.md"},
		{"setext_header/leading_space_in_text.md"},
		{"setext_header/leading_space_in_underline.md"},
		{"setext_header/simple.md"},
		{"setext_header/span_in_text.md"},
		{"setext_header/trailing_space_in_underline.md"},
		{"setext_header/vs_atx_header.md"},
		{"setext_header/vs_blockquote.md"},
		{"setext_header/vs_codeblock.md"},
		{"setext_header/vs_list.md"},
		{"unordered_list/all_items_loose.md"},
		{"unordered_list/all_items_tight.md"},
		{"unordered_list/all_items_tight_even_with_blanks.md"},
		{"unordered_list/at_end_of_parent_without_blank_line.md"},
		{"unordered_list/bw_ordered_lists.md"},
		{"unordered_list/changing_bullet.md"},
		{"unordered_list/changing_starter_string.md"},
		{"unordered_list/different_bullet_chars.md"},
		{"unordered_list/followed_by_hr.md"},
		{"unordered_list/followed_by_list.md"},
		{"unordered_list/indent_and_sub_blocks.md"},
		{"unordered_list/list_ends_with_2blanks.md"},
		{"unordered_list/many_level_nesting.md"},
		{"unordered_list/many_lines.md"},
		{"unordered_list/many_lines_lazy.md"},
		{"unordered_list/many_paras.md"},
		{"unordered_list/many_paras_2blank.md"},
		{"unordered_list/many_paras_2blank_lazy.md"},
		{"unordered_list/many_paras_lazy.md"},
		{"unordered_list/no_space_after_bullet.md"},
		{"unordered_list/no_space_before_bullet.md"},
		{"unordered_list/one_line.md"},
		{"unordered_list/some_items_loose.md"},
		{"unordered_list/space_before_bullet.md"},
		{"unordered_list/three_paras_loose.md"},
		{"unordered_list/three_paras_tight.md"},
		{"unordered_list/two_paras_loose.md"},
		{"unordered_list/with_atx_header.md"},
		{"unordered_list/with_blockquote.md"},
		{"unordered_list/with_codeblock.md"},
		{"unordered_list/with_para.md"},
		{"unordered_list/with_setext_header.md"},
	}

	for i, c := range cases {
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
		expectedOutput = simplifyHtml(expectedOutput)
		if !bytes.Equal(html, expectedOutput) {
			test.Errorf("case %s blocks:\n%s",
				c.path, spew.Sdump(blocks))
			test.Errorf("case %s expected vs. got DIFF:\n%s",
				c.path, diff.Diff(string(expectedOutput), string(html)))
		}

		if i >= 70 {
			test.Fatal("NIY, TODO finish the test")
		}
	}
}

func quickHtml(blocks []md.Tag) []byte {
	buf := bytes.NewBuffer(nil)
	var err error
	tags := blocks
	for len(tags) > 0 {
		tags, err = htmlBlock(tags, buf, 0)
		if err != nil {
			i := len(blocks) - len(tags)
			fmt.Fprintf(os.Stderr, "%s\n%s\n%s\n",
				spew.Sdump(blocks[:i]), err, spew.Sdump(blocks[i:]))
			panic(err)
		}
	}
	return buf.Bytes()
}

func htmlBlock(tags []md.Tag, w io.Writer, opt int) ([]md.Tag, error) {
	var err error
	switch t := tags[0].(type) {
	case md.AtxHeaderBlock:
		fmt.Fprintf(w, "<h%d>", t.Level)
		tags, err = htmlSpans(tags[1:], w)
		fmt.Fprintf(w, "</h%d>\n", t.Level)
		return tags, err
	case md.SetextHeaderBlock:
		fmt.Fprintf(w, "<h%d>", t.Level)
		tags, err = htmlSpans(tags[1:], w)
		fmt.Fprintf(w, "</h%d>\n", t.Level)
		return tags, err
	case md.NullBlock:
		fmt.Fprintln(w)
		return tags[2:], nil
	case md.QuoteBlock:
		fmt.Fprintf(w, "<blockquote>\n  ")
		tags, err = htmlBlocks(tags[1:], w, 0)
		fmt.Fprintf(w, "</blockquote>\n")
		return tags, err
	case md.ParagraphBlock:
		n := len(t.Raw)
		no_p := (opt&1 != 0) ||
			(opt&2 != 0 && t.Raw[n-1].Line == opt>>2)
		if !no_p {
			fmt.Fprintf(w, "<p>")
		}
		tags, err = htmlSpans(tags[1:], w)
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
		fmt.Fprintf(w, "<ol>\n")
		tags, err = htmlItems(tags[1:], w, t.Raw)
		fmt.Fprintf(w, "</ol>\n")
		return tags, err
	case md.UnorderedListBlock:
		fmt.Fprintf(w, "<ul>\n")
		tags, err = htmlItems(tags[1:], w, t.Raw)
		fmt.Fprintf(w, "</ul>\n")
		return tags, err
	default:
		return tags, fmt.Errorf("block type %T not supported yet", t)
	}
}

func isBlank(line md.Run) bool {
	return len(bytes.Trim(line.Bytes, " \t\n")) == 0
}

func htmlItems(tags []md.Tag, w io.Writer, parentRegion md.Raw) ([]md.Tag, error) {
	var err error
	for {
		if (tags[0] == md.End{}) {
			return tags[1:], nil
		}

		t := tags[0].(md.ItemBlock)
		opt := 0
		// top-packed?
		n, m := len(t.Raw), len(parentRegion)
		ifirst, ilast := t.Raw[0].Line, t.Raw[n-1].Line
		lfirst, llast := parentRegion[0].Line, parentRegion[m-1].Line
		if n == m {
			opt = 1
		} else if ifirst == lfirst && !isBlank(t.Raw[n-1]) {
			opt = 1
		} else if ifirst > lfirst && !isBlank(parentRegion[ifirst-lfirst-1]) {
			opt = 1
		}
		// bottom-packed?
		if n == m {
			opt |= 2
		} else if ilast == llast && !isBlank(parentRegion[ifirst-lfirst-1]) {
			opt |= 2
		} else if ilast < llast && !isBlank(t.Raw[n-1]) {
			opt |= 2
		}

		fmt.Fprintf(w, "<li>")
		tags, err = htmlBlocks(tags[1:], w, opt|(t.Raw[n-1].Line<<2))
		fmt.Fprintf(w, "</li>\n")
		if err != nil {
			return tags, err
		}
	}
}

func htmlBlocks(tags []md.Tag, w io.Writer, opt int) ([]md.Tag, error) {
	var err error
	for i := 0; len(tags) > 0; i++ {
		if (tags[0] == md.End{}) {
			return tags[1:], nil
		}
		oopt := opt
		if i != 0 {
			// top-packedness disables <p> only if 1st element
			oopt &= ^int(1)
		}
		if i == 1 {
			// bottom-packedness doesn't disable <p> for 2nd element
			oopt &= ^int(2)
		}
		tags, err = htmlBlock(tags, w, oopt)
		if err != nil {
			return tags, err
		}
	}
	return tags, nil
}

func htmlSpans(tags []md.Tag, w io.Writer) ([]md.Tag, error) {
	var err error
	for {
		switch t := tags[0].(type) {
		case md.Prose:
			for _, r := range t {
				w.Write(r.Bytes)
			}
			tags = tags[1:]
		case md.Emphasis:
			fmt.Fprint(w, map[int]string{
				1: "<em>",
				2: "<strong>",
			}[t.Level])
			tags, err = htmlSpans(tags[1:], w)
			fmt.Fprint(w, map[int]string{
				1: "</em>",
				2: "</strong>",
			}[t.Level])
		case md.AutomaticLink:
			fmt.Fprintf(w, `<a href="%s">%s</a>`,
				// FIXME(akavel): fully correct escaping
				t.URL, html.EscapeString(t.Text))
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
	return bytes.TrimSpace(reSimplifyHtml.ReplaceAllLiteral(buf, []byte(">\n<")))
}
