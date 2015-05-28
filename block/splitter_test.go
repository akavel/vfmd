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
		{
			"codeblock/html_escaping.md",
			blocks{
				{2, Paragraph{}},
				{1, Code{}},
				{1, Null{}},
			},
		},
		{
			"codeblock/many_lines.md",
			blocks{{3, Code{}}},
		},
		{
			"codeblock/more_than_four_leading_space.md",
			blocks{
				{1, Code{}},
				{1, Null{}},
				{2, Paragraph{}},
				{3, Code{}},
			},
		},
		{
			"codeblock/one_blank_line_bw_codeblocks.md",
			blocks{{7, Code{}}},
		},
		{
			"codeblock/one_line.md",
			blocks{{1, Code{}}},
		},
		{
			"codeblock/two_blank_lines_bw_codeblocks.md",
			blocks{
				{3, Code{}},
				{1, Null{}},
				{1, Null{}},
				{3, Code{}},
			},
		},
		{
			"codeblock/vs_atx_header.md",
			blocks{
				{1, Code{}},
				{1, Null{}},
				{2, Paragraph{}},
				{1, Code{}},
				{1, Null{}},
				{2, Paragraph{}},
				{2, Code{}},
			},
		},
		{
			"codeblock/vs_blockquote.md",
			blocks{
				{1, Code{}},
				{1, Null{}},
				{2, Paragraph{}},
				{1, Code{}},
				{1, Null{}},
				{2, Paragraph{}},
				{2, Code{}},
			},
		},
		{
			"codeblock/vs_hr.md",
			blocks{
				{2, Paragraph{}},
				{1, Code{}},
				{1, Null{}},
			},
		},
		{
			"codeblock/vs_list.md",
			blocks{
				{2, Paragraph{}},
				{2, Code{}},
				{1, Null{}},
				{2, Paragraph{}},
				{2, Code{}},
				{1, Null{}},
				{2, Paragraph{}},
				{2, Code{}},
				{1, Null{}},
				{2, Paragraph{}},
				{2, Code{}},
				{1, Null{}},
			},
		},
		{
			"horizontal_rule/end_with_space.md",
			blocks{
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
			},
		},
		{
			"horizontal_rule/followed_by_block.md",
			blocks{
				{1, HorizontalRule{}},
				{2, Paragraph{}},
				{1, HorizontalRule{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{2, Quote{}},
				{1, HorizontalRule{}},
				{2, &OrderedList{}},
				{1, HorizontalRule{}},
				{2, &UnorderedList{}},
				{1, HorizontalRule{}},
				{2, &UnorderedList{}},
				{1, HorizontalRule{}},
				{1, Code{}},
				{1, Null{}},

				{1, HorizontalRule{}},
				{2, Paragraph{}},
				{1, HorizontalRule{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{2, Quote{}},
				{1, HorizontalRule{}},
				{2, &OrderedList{}},
				{1, HorizontalRule{}},
				{2, &UnorderedList{}},
				{1, HorizontalRule{}},
				{2, &UnorderedList{}},
				{1, HorizontalRule{}},
				{1, Code{}},
				{1, Null{}},

				{1, HorizontalRule{}},
				{2, Paragraph{}},
				{1, HorizontalRule{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{2, Quote{}},
				{1, HorizontalRule{}},
				{2, &OrderedList{}},
				{1, HorizontalRule{}},
				{2, &UnorderedList{}},
				{1, HorizontalRule{}},
				{2, &UnorderedList{}},
				{1, HorizontalRule{}},
				{1, Code{}},
				{1, Null{}},
			},
		},
		{
			"horizontal_rule/loose.md",
			blocks{
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
			},
		},
		{
			"horizontal_rule/sparse.md",
			blocks{
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
			},
		},
		{
			"horizontal_rule/start_with_space.md",
			blocks{
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
			},
		},
		{
			"horizontal_rule/tight.md",
			blocks{
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, HorizontalRule{}},
			},
		},
		{
			"ordered_list/all_items_loose.md",
			blocks{{7, &OrderedList{}}},
		},
		{
			"ordered_list/all_items_tight.md",
			blocks{{4, &OrderedList{}}},
		},
		{
			"ordered_list/all_items_tight_even_with_blanks.md",
			blocks{{7, &OrderedList{}}},
		},
		{
			"ordered_list/at_end_of_parent_without_blank_line.md",
			blocks{
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{8, &UnorderedList{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{5, &OrderedList{}},
			},
		},
		{
			"ordered_list/bw_unordered_lists.md",
			blocks{
				{3, &UnorderedList{}},
				{3, &OrderedList{}},
				{3, &UnorderedList{}},
			},
		},
		{
			"ordered_list/followed_by_hr.md",
			blocks{
				{4, &OrderedList{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{5, &OrderedList{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{6, &OrderedList{}},
				{1, HorizontalRule{}},
			},
		},
		{
			"ordered_list/followed_by_list.md",
			blocks{
				{4, &OrderedList{}},
				{2, &UnorderedList{}},
				{5, &OrderedList{}},
				{2, &UnorderedList{}},
				{6, &OrderedList{}},
				{1, &UnorderedList{}},
			},
		},
		{
			"ordered_list/indent_and_sub_blocks.md",
			blocks{
				{6, &OrderedList{}},
				{2, Paragraph{}},
				{6, &OrderedList{}},
				{2, Quote{}},
				{6, &OrderedList{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{6, &OrderedList{}},
				{2, &UnorderedList{}},
				{12, &OrderedList{}},
				{1, HorizontalRule{}},
				{2, &OrderedList{}},
				{1, Code{}},
				{1, Null{}},
			},
		},
		{
			"ordered_list/list_ends_with_2blanks.md",
			blocks{
				{4, &OrderedList{}},
				{1, Null{}},
				{10, &OrderedList{}},
				{1, Null{}},
				{1, Code{}},
			},
		},
		{
			"ordered_list/many_level_nesting.md",
			blocks{
				{2, Paragraph{}},
				{22, &OrderedList{}},
				{1, Paragraph{}},
			},
		},
		{
			"ordered_list/many_lines.md",
			blocks{{5, &OrderedList{}}},
		},
		{
			"ordered_list/many_lines_lazy.md",
			blocks{{5, &OrderedList{}}},
		},
		{
			"ordered_list/many_paras.md",
			blocks{{9, &OrderedList{}}},
		},
		{
			"ordered_list/many_paras_2blank.md",
			blocks{
				{5, &OrderedList{}},
				{1, Null{}},
				{4, Paragraph{}},
			},
		},
		{
			"ordered_list/many_paras_2blank_lazy.md",
			blocks{
				{5, &OrderedList{}},
				{1, Null{}},
				{4, Paragraph{}},
			},
		},
		{
			"ordered_list/many_paras_lazy.md",
			blocks{{9, &OrderedList{}}},
		},
		{
			"ordered_list/no_space_after_number.md",
			blocks{
				{2, Paragraph{}},
				{2, Paragraph{}},
				{2, Paragraph{}},
				{2, Paragraph{}},
				{2, Paragraph{}},
				{1, Paragraph{}},
			},
		},
		{
			"ordered_list/no_space_before_number.md",
			blocks{
				{2, Paragraph{}},
				{2, &OrderedList{}},
				{2, Paragraph{}},
				{2, &OrderedList{}},
				{2, Paragraph{}},
				{1, &OrderedList{}},
			},
		},
		{
			"ordered_list/numbering_from_two.md",
			blocks{{11, &OrderedList{}}},
		},
		{
			"ordered_list/numbering_not_in_order.md",
			blocks{{12, &OrderedList{}}},
		},
		{
			"ordered_list/numbers_left_aligned.md",
			blocks{{12, &OrderedList{}}},
		},
		{
			"ordered_list/numbers_right_aligned.md",
			blocks{{12, &OrderedList{}}},
		},
		{
			"ordered_list/numbers_wiggly.md",
			blocks{{12, &OrderedList{}}},
		},
		{
			"ordered_list/one_line.md",
			blocks{{3, &OrderedList{}}},
		},
		{
			"ordered_list/some_items_loose.md",
			blocks{
				{2, Paragraph{}},
				{6, &OrderedList{}},
				{2, Paragraph{}},
				{8, &OrderedList{}},
				{2, Paragraph{}},
				{10, &OrderedList{}},
				{2, Paragraph{}},
				{6, &OrderedList{}},
				{1, Paragraph{}},
			},
		},
		{
			"ordered_list/space_before_number.md",
			blocks{
				{2, Paragraph{}},
				{2, &OrderedList{}},
				{2, Paragraph{}},
				{2, &OrderedList{}},
				{2, Paragraph{}},
				{2, &OrderedList{}},
				{2, Paragraph{}},
				{2, &OrderedList{}},
				{2, Paragraph{}},
				{1, Code{}},
				{1, Null{}},
			},
		},
		{
			"ordered_list/three_paras_loose.md",
			blocks{{15, &OrderedList{}}},
		},
		{
			"ordered_list/three_paras_tight.md",
			blocks{{13, &OrderedList{}}},
		},
		{
			"ordered_list/two_paras_loose.md",
			blocks{{11, &OrderedList{}}},
		},
		{
			"ordered_list/with_atx_header.md",
			blocks{
				{16, &OrderedList{}},
				{1, AtxHeader{}},
			},
		},
		{
			"ordered_list/with_blockquote.md",
			blocks{
				{8, &OrderedList{}},
				{1, Quote{}},
			},
		},
		{
			"ordered_list/with_codeblock.md",
			blocks{
				{8, &OrderedList{}},
				{2, Paragraph{}},
				{2, &OrderedList{}},
				{1, Code{}},
			},
		},
		{
			"ordered_list/with_para.md",
			blocks{
				{10, &OrderedList{}},
				{1, Paragraph{}},
			},
		},
		{
			"ordered_list/with_setext_header.md",
			blocks{
				{25, &OrderedList{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{10, &OrderedList{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{3, &OrderedList{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{3, &OrderedList{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{3, &OrderedList{}},
				{2, SetextHeader{}},
				{1, Null{}},
			},
		},
		// {
		// "paragraph/blanks_within_html_comment.md",
		// blocks{
		// },
		// },
		{
			"paragraph/followed_by_atx_header.md",
			blocks{
				{3, Paragraph{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{3, Paragraph{}},
			},
		},
		{
			"paragraph/followed_by_blockquote.md",
			blocks{
				{3, Paragraph{}},
				{3, Quote{}},
				{4, Paragraph{}},
			},
		},
		{
			"paragraph/followed_by_codeblock.md",
			blocks{
				{3, Paragraph{}},
				{2, Code{}},
				{1, Null{}},
				{4, Paragraph{}},
			},
		},
		{
			"paragraph/followed_by_horizontal_rule.md",
			blocks{
				{3, Paragraph{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{1, Null{}},
				{2, Paragraph{}},
				{1, HorizontalRule{}},
			},
		},
		{
			"paragraph/followed_by_list.md",
			blocks{
				{3, Paragraph{}},
				{2, &UnorderedList{}},
				{4, Paragraph{}},
				{3, Paragraph{}},
				{2, &OrderedList{}},
				{4, Paragraph{}},
			},
		},
		{
			"paragraph/followed_by_setext_header.md",
			blocks{
				{3, Paragraph{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{4, Paragraph{}},
			},
		},
		{
			"paragraph/simple_para.md",
			blocks{{3, Paragraph{}}},
		},
		{
			"paragraph/two_paras_1blank.md",
			blocks{
				{3, Paragraph{}},
				{2, Paragraph{}},
			},
		},
		{
			"paragraph/two_paras_2blank.md",
			blocks{
				{3, Paragraph{}},
				{1, Null{}},
				{2, Paragraph{}},
			},
		},
		{
			"setext_header/blank_text.md",
			blocks{
				{2, Paragraph{}},
				{2, Paragraph{}},
				{2, Paragraph{}},
				{1, HorizontalRule{}},
			},
		},
		{
			"setext_header/enclosed_space_in_underline.md",
			blocks{
				{3, Paragraph{}},
				{1, Paragraph{}},
				{1, HorizontalRule{}},
			},
		},
		{
			"setext_header/leading_space_in_text.md",
			blocks{
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
			},
		},
		{
			"setext_header/leading_space_in_underline.md",
			blocks{
				{3, Paragraph{}},
				{1, Paragraph{}},
				{1, HorizontalRule{}},
				{1, Null{}},
			},
		},
		{
			"setext_header/simple.md",
			blocks{
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
			},
		},
		{
			"setext_header/span_in_text.md",
			blocks{
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
			},
		},
		{
			"setext_header/trailing_space_in_underline.md",
			blocks{
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
			},
		},
		{
			"setext_header/vs_atx_header.md",
			blocks{
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
			},
		},
		{
			"setext_header/vs_blockquote.md",
			blocks{
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
			},
		},
		{
			"setext_header/vs_codeblock.md",
			blocks{
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
			},
		},
		{
			"setext_header/vs_list.md",
			blocks{
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{2, SetextHeader{}},
			},
		},
		{
			"unordered_list/all_items_loose.md",
			blocks{{7, &UnorderedList{}}},
		},
		{
			"unordered_list/all_items_tight.md",
			blocks{{4, &UnorderedList{}}},
		},
		{
			"unordered_list/all_items_tight_even_with_blanks.md",
			blocks{{7, &UnorderedList{}}},
		},
		{
			"unordered_list/at_end_of_parent_without_blank_line.md",
			blocks{
				{1, Null{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{8, &UnorderedList{}},
				{1, AtxHeader{}},
				{1, Null{}},
				{5, &UnorderedList{}},
			},
		},
		{
			"unordered_list/bw_ordered_lists.md",
			blocks{
				{3, &OrderedList{}},
				{3, &UnorderedList{}},
				{3, &OrderedList{}},
			},
		},
		{
			"unordered_list/changing_bullet.md",
			blocks{
				{3, &UnorderedList{}},
				{3, &UnorderedList{}},
				{3, &UnorderedList{}},
			},
		},
		{
			"unordered_list/changing_starter_string.md",
			blocks{
				{4, &UnorderedList{}},
				{2, &UnorderedList{}},
				{4, &UnorderedList{}},
				{2, &UnorderedList{}},
				{1, &UnorderedList{}},
				{2, &OrderedList{}},
				{1, &UnorderedList{}},
			},
		},
		{
			"unordered_list/different_bullet_chars.md",
			blocks{
				{2, Paragraph{}},
				{2, &UnorderedList{}},
				{2, Paragraph{}},
				{2, &UnorderedList{}},
				{2, Paragraph{}},
				{2, &UnorderedList{}},
			},
		},
		{
			"unordered_list/followed_by_hr.md",
			blocks{
				{4, &UnorderedList{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{5, &UnorderedList{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{6, &UnorderedList{}},
				{1, HorizontalRule{}},
			},
		},
		{
			"unordered_list/followed_by_list.md",
			blocks{
				{4, &UnorderedList{}},
				{2, &UnorderedList{}},
				{5, &UnorderedList{}},
				{2, &UnorderedList{}},
				{6, &UnorderedList{}},
				{1, &UnorderedList{}},

				{4, &UnorderedList{}},
				{2, &OrderedList{}},
				{5, &UnorderedList{}},
				{2, &OrderedList{}},
				{6, &UnorderedList{}},
				{1, &OrderedList{}},
			},
		},
		{
			"unordered_list/indent_and_sub_blocks.md",
			blocks{
				{6, &UnorderedList{}},
				{2, Paragraph{}},
				{6, &UnorderedList{}},
				{2, Quote{}},
				{6, &UnorderedList{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{6, &UnorderedList{}},
				{2, &UnorderedList{}},
				{6, &UnorderedList{}},
				{2, &OrderedList{}},
				{6, &UnorderedList{}},
				{1, Code{}},
				{1, Null{}},
			},
		},
		{
			"unordered_list/list_ends_with_2blanks.md",
			blocks{
				{4, &UnorderedList{}},
				{1, Null{}},
				{10, &UnorderedList{}},
				{1, Null{}},
				{1, Code{}},
			},
		},
		{
			"unordered_list/many_level_nesting.md",
			blocks{
				{2, Paragraph{}},
				{22, &UnorderedList{}},
				{1, Paragraph{}},
			},
		},
		{
			"unordered_list/many_lines.md",
			blocks{{5, &UnorderedList{}}},
		},
		{
			"unordered_list/many_lines_lazy.md",
			blocks{{5, &UnorderedList{}}},
		},
		{
			"unordered_list/many_paras.md",
			blocks{{9, &UnorderedList{}}},
		},
		{
			"unordered_list/many_paras_2blank.md",
			blocks{
				{5, &UnorderedList{}},
				{1, Null{}},
				{4, Paragraph{}},
			},
		},
		{
			"unordered_list/many_paras_2blank_lazy.md",
			blocks{
				{5, &UnorderedList{}},
				{1, Null{}},
				{4, Paragraph{}},
			},
		},
		{
			"unordered_list/many_paras_lazy.md",
			blocks{{9, &UnorderedList{}}},
		},
		{
			"unordered_list/no_space_after_bullet.md",
			blocks{
				{2, Paragraph{}},
				{2, Paragraph{}},
				{2, Paragraph{}},
				{2, Paragraph{}},
				{2, Paragraph{}},
				{1, Paragraph{}},
			},
		},
		{
			"unordered_list/no_space_before_bullet.md",
			blocks{
				{2, Paragraph{}},
				{2, &UnorderedList{}},
				{2, Paragraph{}},
				{2, &UnorderedList{}},
				{2, Paragraph{}},
				{1, &UnorderedList{}},
			},
		},
		{
			"unordered_list/one_line.md",
			blocks{{3, &UnorderedList{}}},
		},
		{
			"unordered_list/some_items_loose.md",
			blocks{
				{2, Paragraph{}},
				{6, &UnorderedList{}},
				{2, Paragraph{}},
				{8, &UnorderedList{}},
				{2, Paragraph{}},
				{10, &UnorderedList{}},
				{2, Paragraph{}},
				{6, &UnorderedList{}},
				{1, Paragraph{}},
			},
		},
		{
			"unordered_list/space_before_bullet.md",
			blocks{
				{2, Paragraph{}},
				{2, &UnorderedList{}},
				{2, Paragraph{}},
				{2, &UnorderedList{}},
				{2, Paragraph{}},
				{2, &UnorderedList{}},
				{2, Paragraph{}},
				{2, &UnorderedList{}},
				{2, Paragraph{}},
				{1, Code{}},
				{1, Null{}},
			},
		},
		{
			"unordered_list/three_paras_loose.md",
			blocks{{15, &UnorderedList{}}},
		},
		{
			"unordered_list/three_paras_tight.md",
			blocks{{13, &UnorderedList{}}},
		},
		{
			"unordered_list/two_paras_loose.md",
			blocks{{11, &UnorderedList{}}},
		},
		{
			"unordered_list/with_atx_header.md",
			blocks{
				{16, &UnorderedList{}},
				{1, AtxHeader{}},
			},
		},
		{
			"unordered_list/with_blockquote.md",
			blocks{
				{8, &UnorderedList{}},
				{1, Quote{}},
			},
		},
		{
			"unordered_list/with_codeblock.md",
			blocks{
				{8, &UnorderedList{}},
				{2, Paragraph{}},
				{2, &UnorderedList{}},
				{1, Code{}},
			},
		},
		{
			"unordered_list/with_para.md",
			blocks{
				{10, &UnorderedList{}},
				{1, Paragraph{}},
			},
		},
		{
			"unordered_list/with_setext_header.md",
			blocks{
				{25, &UnorderedList{}},
				{2, SetextHeader{}},
				{1, Null{}},
				{10, &UnorderedList{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{3, &UnorderedList{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{3, &UnorderedList{}},
				{1, HorizontalRule{}},
				{1, Null{}},
				{3, &UnorderedList{}},
				{2, SetextHeader{}},
				{1, Null{}},
			},
		},
	}

Cases:
	for _, c := range cases {
		// test.Logf("case %s", c.path)
		s := Splitter{}
		data, err := ioutil.ReadFile(filepath.Join(dir, c.path))
		if err != nil {
			test.Error(err)
			continue
		}
		scan := bufio.NewScanner(bytes.NewReader(data))
		prep := vfmd.Preprocessor{}
		for i := 0; scan.Scan(); i++ {
			// test.Logf("line %d", i)
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
			// test.Logf("prepped=%q", string(line))

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
			test.Errorf("case %s length mismatch, expected %d:",
				c.path, len(c.blocks))
			for _, b := range c.blocks {
				test.Errorf("%#v", b)
			}
			test.Errorf("got %d", len(s.Blocks))
			for _, b := range s.Blocks {
				test.Errorf("%#v", b)
			}
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
