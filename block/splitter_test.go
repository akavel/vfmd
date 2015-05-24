package block_test

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/akavel/vfmd-go"
	. "github.com/akavel/vfmd-go/block"
)

const dir = "../testdata/tests/block_level"

/*
+atx_header/blank_text.md
+atx_header/enclosed_blank_text.md
+atx_header/hash_in_text.md
+atx_header/left_only.md
+atx_header/left_right.md
+atx_header/more_than_six_hashes.md
+atx_header/space_in_text.md
+atx_header/span_in_text.md
+blockquote/containing_atx_header.md
+blockquote/containing_blockquote.md
+blockquote/containing_codeblock.md
+blockquote/containing_hr.md
+blockquote/containing_list.md
+blockquote/containing_setext_header.md
+blockquote/followed_by_atx_header.md
+blockquote/followed_by_codeblock.md
+blockquote/followed_by_hr.md
+blockquote/followed_by_list.md
+blockquote/followed_by_para.md
+blockquote/followed_by_setext_header.md
+blockquote/indented_differently1.md
+blockquote/indented_differently2.md
+blockquote/many_level_nesting.md
+blockquote/many_lines.md
+blockquote/many_lines_lazy.md
+blockquote/many_paras.md
+blockquote/many_paras_2blank.md
+blockquote/many_paras_2blank_lazy.md
+blockquote/many_paras_2blank_lazy2.md
+blockquote/many_paras_lazy.md
+blockquote/many_paras_lazy2.md
+blockquote/no_space_after_gt.md
+blockquote/one_line.md
+blockquote/space_before_gt.md
+codeblock/followed_by_para.md
codeblock/html_escaping.md
codeblock/many_lines.md
codeblock/more_than_four_leading_space.md
codeblock/one_blank_line_bw_codeblocks.md
codeblock/one_line.md
codeblock/two_blank_lines_bw_codeblocks.md
codeblock/vs_atx_header.md
codeblock/vs_blockquote.md
codeblock/vs_hr.md
codeblock/vs_list.md
horizontal_rule/end_with_space.md
horizontal_rule/followed_by_block.md
horizontal_rule/loose.md
horizontal_rule/sparse.md
horizontal_rule/start_with_space.md
horizontal_rule/tight.md
ordered_list/all_items_loose.md
ordered_list/all_items_tight.md
ordered_list/all_items_tight_even_with_blanks.md
ordered_list/at_end_of_parent_without_blank_line.md
ordered_list/bw_unordered_lists.md
ordered_list/followed_by_hr.md
ordered_list/followed_by_list.md
ordered_list/indent_and_sub_blocks.md
ordered_list/list_ends_with_2blanks.md
ordered_list/many_level_nesting.md
ordered_list/many_lines.md
ordered_list/many_lines_lazy.md
ordered_list/many_paras.md
ordered_list/many_paras_2blank.md
ordered_list/many_paras_2blank_lazy.md
ordered_list/many_paras_lazy.md
ordered_list/no_space_after_number.md
ordered_list/no_space_before_number.md
ordered_list/numbering_from_two.md
ordered_list/numbering_not_in_order.md
ordered_list/numbers_left_aligned.md
ordered_list/numbers_right_aligned.md
ordered_list/numbers_wiggly.md
ordered_list/one_line.md
ordered_list/some_items_loose.md
ordered_list/space_before_number.md
ordered_list/three_paras_loose.md
ordered_list/three_paras_tight.md
ordered_list/two_paras_loose.md
ordered_list/with_atx_header.md
ordered_list/with_blockquote.md
ordered_list/with_codeblock.md
ordered_list/with_para.md
ordered_list/with_setext_header.md
paragraph/blanks_within_html_comment.md
paragraph/blanks_within_html_tag.md
paragraph/blanks_within_verbatim_html.md
paragraph/followed_by_atx_header.md
paragraph/followed_by_blockquote.md
paragraph/followed_by_codeblock.md
paragraph/followed_by_horizontal_rule.md
paragraph/followed_by_list.md
paragraph/followed_by_setext_header.md
paragraph/html_block.md
paragraph/html_comment.md
paragraph/md_within_html.md
paragraph/misnested_html.md
paragraph/non_phrasing_html_tag.md
paragraph/phrasing_html_tag.md
paragraph/simple_para.md
paragraph/two_paras_1blank.md
paragraph/two_paras_2blank.md
setext_header/blank_text.md
setext_header/enclosed_space_in_underline.md
setext_header/leading_space_in_text.md
setext_header/leading_space_in_underline.md
setext_header/simple.md
setext_header/span_in_text.md
setext_header/trailing_space_in_underline.md
setext_header/vs_atx_header.md
setext_header/vs_blockquote.md
setext_header/vs_codeblock.md
setext_header/vs_list.md
unordered_list/all_items_loose.md
unordered_list/all_items_tight.md
unordered_list/all_items_tight_even_with_blanks.md
unordered_list/at_end_of_parent_without_blank_line.md
unordered_list/bw_ordered_lists.md
unordered_list/changing_bullet.md
unordered_list/changing_starter_string.md
unordered_list/different_bullet_chars.md
unordered_list/followed_by_hr.md
unordered_list/followed_by_list.md
unordered_list/indent_and_sub_blocks.md
unordered_list/list_ends_with_2blanks.md
unordered_list/many_level_nesting.md
unordered_list/many_lines.md
unordered_list/many_lines_lazy.md
unordered_list/many_paras.md
unordered_list/many_paras_2blank.md
unordered_list/many_paras_2blank_lazy.md
unordered_list/many_paras_lazy.md
unordered_list/no_space_after_bullet.md
unordered_list/no_space_before_bullet.md
unordered_list/one_line.md
unordered_list/some_items_loose.md
unordered_list/space_before_bullet.md
unordered_list/three_paras_loose.md
unordered_list/three_paras_tight.md
unordered_list/two_paras_loose.md
unordered_list/with_atx_header.md
unordered_list/with_blockquote.md
unordered_list/with_codeblock.md
unordered_list/with_para.md
unordered_list/with_setext_header.md
*/

func TestFiles(test *testing.T) {
	type blocks []struct {
		n int
		Detector
	}
	cases := []struct {
		path string
		blocks
	}{
		{
			"atx_header/blank_text.md",
			blocks{
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
			},
		},
		{
			"atx_header/enclosed_blank_text.md",
			blocks{
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
			},
		},
		{
			"atx_header/hash_in_text.md",
			blocks{
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
			},
		},
		{
			"atx_header/left_only.md",
			blocks{
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
			},
		},
		{
			"atx_header/left_right.md",
			blocks{
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
			},
		},
		{
			"atx_header/more_than_six_hashes.md",
			blocks{
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
			},
		},
		{
			"atx_header/space_in_text.md",
			blocks{
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
			},
		},
		{
			"atx_header/span_in_text.md",
			blocks{
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, AtxHeader{}},
			},
		},
		{
			"blockquote/containing_atx_header.md",
			blocks{{6, Quote{}}},
		},
		{
			"blockquote/containing_blockquote.md",
			blocks{{7, Quote{}}},
		},
		{
			"blockquote/containing_codeblock.md",
			blocks{{12, Quote{}}},
		},
		{
			"blockquote/containing_hr.md",
			blocks{{6, Quote{}}},
		},
		{
			"blockquote/containing_list.md",
			blocks{{20, Quote{}}},
		},
		{
			"blockquote/containing_setext_header.md",
			blocks{{9, Quote{}}},
		},
		{
			"blockquote/followed_by_atx_header.md",
			blocks{
				{3, Quote{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{3, Quote{}},
			},
		},
		{
			"blockquote/followed_by_codeblock.md",
			blocks{
				{3, Quote{}},
				{1, Code{}},
				{1, Null{}},
				{3, Quote{}},
			},
		},
		{
			"blockquote/followed_by_hr.md",
			blocks{
				{3, Quote{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{2, Quote{}},
				{1, HorizontalRule{}},
			},
		},
		{
			"blockquote/followed_by_list.md",
			blocks{
				{3, Quote{}},
				{2, &OrderedList{}},
				{7, Quote{}},
				{2, &UnorderedList{}},
				{7, Quote{}},
				{2, &UnorderedList{}},
				{7, Quote{}},
				{2, &UnorderedList{}},
				{3, Quote{}},
			},
		},
		{
			"blockquote/followed_by_para.md",
			blocks{
				{3, Quote{}},
				{2, Paragraph{}},
				{3, Quote{}},
			},
		},
		{
			"blockquote/followed_by_setext_header.md",
			blocks{
				{3, Quote{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{4, Quote{}},
			},
		},
		{
			"blockquote/indented_differently1.md",
			blocks{{6, Quote{}}},
		},
		{
			"blockquote/indented_differently2.md",
			blocks{{11, Quote{}}},
		},
		{
			"blockquote/many_level_nesting.md",
			blocks{
				{2, Paragraph{}},
				{26, Quote{}},
				{1, Paragraph{}},
			},
		},
		{
			"blockquote/many_lines.md",
			blocks{{3, Quote{}}},
		},
		{
			"blockquote/many_lines_lazy.md",
			blocks{{3, Quote{}}},
		},
		{
			"blockquote/many_paras.md",
			blocks{{11, Quote{}}},
		},
		{
			"blockquote/many_paras_2blank.md",
			blocks{{13, Quote{}}},
		},
		{
			"blockquote/many_paras_2blank_lazy.md",
			blocks{
				{4, Quote{}},
				{1, Null{}},
				{4, Quote{}},
				{1, Null{}},
				{3, Quote{}},
			},
		},
		{
			"blockquote/many_paras_2blank_lazy2.md",
			blocks{
				{4, Quote{}},
				{1, Null{}},
				{4, Quote{}},
				{1, Null{}},
				{3, Quote{}},
			},
		},
		{
			"blockquote/many_paras_lazy.md",
			blocks{{11, Quote{}}},
		},
		{
			"blockquote/many_paras_lazy2.md",
			blocks{{11, Quote{}}},
		},
		{
			"blockquote/no_space_after_gt.md",
			blocks{
				{2, Paragraph{}},
				{2, Quote{}},
				{2, Paragraph{}},
				{2, Quote{}},
			},
		},
		{
			"blockquote/one_line.md",
			blocks{{1, Quote{}}},
		},
		{
			"blockquote/space_before_gt.md",
			blocks{
				{2, Paragraph{}},
				{2, Quote{}},
				{2, Paragraph{}},
				{2, Quote{}},
				{2, Paragraph{}},
				{2, Quote{}},
				{2, Paragraph{}},
				{1, Code{}},
			},
		},
		{
			"codeblock/followed_by_para.md",
			blocks{
				{2, Paragraph{}},
				{2, Code{}},
				{2, Paragraph{}},
				{2, Code{}},
				{1, Null{}},
				{1, Paragraph{}},
			},
		},
	}

Cases:
	for _, c := range cases {
		test.Logf("case %s", c.path)
		s := Splitter{}
		data, err := ioutil.ReadFile(filepath.Join(dir, c.path))
		if err != nil {
			test.Error(err)
			continue
		}
		scan := bufio.NewScanner(bytes.NewReader(data))
		prep := vfmd.Preprocessor{}
		for i := 0; scan.Scan(); i++ {
			test.Logf("line %d", i)
			_, err := prep.Write(scan.Bytes())
			if err != nil {
				test.Fatal(err)
			}
			if len(prep.Pending) > 0 {
				test.Fatal(err)
			}
			line := []byte{}
			for _, chunk := range prep.Chunks {
				line = append(line, chunk.Bytes...)
			}
			prep.Chunks = nil
			test.Logf("prepped=%q", string(line))

			err = s.WriteLine(line)
			if err != nil {
				test.Error(err)
				continue Cases
			}
		}
		if scan.Err() != nil {
			test.Error(err)
			continue
		}
		err = s.Close()
		if err != nil {
			test.Error(err)
		}

		if len(c.blocks) != len(s.Blocks) {
			test.Errorf("case %s length mismatch, expected:\n%d (%#v)\ngot:\n%d (%#v)",
				c.path, len(c.blocks), c.blocks, len(s.Blocks), s.Blocks)
			continue
		}
		for i := range c.blocks {
			if c.blocks[i].n != s.Blocks[i].Last-s.Blocks[i].First+1 {
				test.Errorf("case %s block %d length expected %d got %#v", c.path, i, c.blocks[i].n, s.Blocks[i])
			}
			if reflect.TypeOf(c.blocks[i].Detector).String() != reflect.TypeOf(s.Blocks[i].Detector).String() {
				test.Errorf("case %s block %d type expected %T got %#v", c.path, i, c.blocks[i].Detector, s.Blocks[i])
			}
		}
	}
}
