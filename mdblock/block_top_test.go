package mdblock

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/kylelemons/godebug/diff"

	"gopkg.in/akavel/vfmd.v1"
	"gopkg.in/akavel/vfmd.v1/md"
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
	type block struct {
		n   int
		tag string
	}
	type blocks []block
	cases := []struct {
		path string
		blocks
	}{
		{
			"atx_header/blank_text.md",
			blocks{
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
			},
		},
		{
			"atx_header/enclosed_blank_text.md",
			blocks{
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
			},
		},
		{
			"atx_header/hash_in_text.md",
			blocks{
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
			},
		},
		{
			"atx_header/left_only.md",
			blocks{
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
			},
		},
		{
			"atx_header/left_right.md",
			blocks{
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
			},
		},
		{
			"atx_header/more_than_six_hashes.md",
			blocks{
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
			},
		},
		{
			"atx_header/space_in_text.md",
			blocks{
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
			},
		},
		{
			"atx_header/span_in_text.md",
			blocks{
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
			},
		},
		{
			"blockquote/containing_atx_header.md",
			blocks{{6, "md.QuoteBlock"}},
		},
		{
			"blockquote/containing_blockquote.md",
			blocks{{7, "md.QuoteBlock"}},
		},
		{
			"blockquote/containing_codeblock.md",
			blocks{{12, "md.QuoteBlock"}},
		},
		{
			"blockquote/containing_hr.md",
			blocks{{6, "md.QuoteBlock"}},
		},
		{
			"blockquote/containing_list.md",
			blocks{{20, "md.QuoteBlock"}},
		},
		{
			"blockquote/containing_setext_header.md",
			blocks{{9, "md.QuoteBlock"}},
		},
		{
			"blockquote/followed_by_atx_header.md",
			blocks{
				{3, "md.QuoteBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{3, "md.QuoteBlock"},
			},
		},
		{
			"blockquote/followed_by_codeblock.md",
			blocks{
				{3, "md.QuoteBlock"},
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},
				{3, "md.QuoteBlock"},
			},
		},
		{
			"blockquote/followed_by_hr.md",
			blocks{
				{3, "md.QuoteBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{2, "md.QuoteBlock"},
				{1, "md.HorizontalRuleBlock"},
			},
		},
		{
			"blockquote/followed_by_list.md",
			blocks{
				{3, "md.QuoteBlock"},
				{2, "md.OrderedListBlock"},
				{7, "md.QuoteBlock"},
				{2, "md.UnorderedListBlock"},
				{7, "md.QuoteBlock"},
				{2, "md.UnorderedListBlock"},
				{7, "md.QuoteBlock"},
				{2, "md.UnorderedListBlock"},
				{3, "md.QuoteBlock"},
			},
		},
		{
			"blockquote/followed_by_para.md",
			blocks{
				{3, "md.QuoteBlock"},
				{2, "md.ParagraphBlock"},
				{3, "md.QuoteBlock"},
			},
		},
		{
			"blockquote/followed_by_setext_header.md",
			blocks{
				{3, "md.QuoteBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{4, "md.QuoteBlock"},
			},
		},
		{
			"blockquote/indented_differently1.md",
			blocks{{6, "md.QuoteBlock"}},
		},
		{
			"blockquote/indented_differently2.md",
			blocks{{11, "md.QuoteBlock"}},
		},
		{
			"blockquote/many_level_nesting.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{26, "md.QuoteBlock"},
				{1, "md.ParagraphBlock"},
			},
		},
		{
			"blockquote/many_lines.md",
			blocks{{3, "md.QuoteBlock"}},
		},
		{
			"blockquote/many_lines_lazy.md",
			blocks{{3, "md.QuoteBlock"}},
		},
		{
			"blockquote/many_paras.md",
			blocks{{11, "md.QuoteBlock"}},
		},
		{
			"blockquote/many_paras_2blank.md",
			blocks{{13, "md.QuoteBlock"}},
		},
		{
			"blockquote/many_paras_2blank_lazy.md",
			blocks{
				{4, "md.QuoteBlock"},
				{1, "md.NullBlock"},
				{4, "md.QuoteBlock"},
				{1, "md.NullBlock"},
				{3, "md.QuoteBlock"},
			},
		},
		{
			"blockquote/many_paras_2blank_lazy2.md",
			blocks{
				{4, "md.QuoteBlock"},
				{1, "md.NullBlock"},
				{4, "md.QuoteBlock"},
				{1, "md.NullBlock"},
				{3, "md.QuoteBlock"},
			},
		},
		{
			"blockquote/many_paras_lazy.md",
			blocks{{11, "md.QuoteBlock"}},
		},
		{
			"blockquote/many_paras_lazy2.md",
			blocks{{11, "md.QuoteBlock"}},
		},
		{
			"blockquote/no_space_after_gt.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{2, "md.QuoteBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.QuoteBlock"},
			},
		},
		{
			"blockquote/one_line.md",
			blocks{{1, "md.QuoteBlock"}},
		},
		{
			"blockquote/space_before_gt.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{2, "md.QuoteBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.QuoteBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.QuoteBlock"},
				{2, "md.ParagraphBlock"},
				{1, "md.CodeBlock"},
			},
		},
		{
			"codeblock/followed_by_para.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{2, "md.CodeBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.CodeBlock"},
				{1, "md.NullBlock"},
				{1, "md.ParagraphBlock"},
			},
		},
		{
			"codeblock/html_escaping.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},
			},
		},
		{
			"codeblock/many_lines.md",
			blocks{{3, "md.CodeBlock"}},
		},
		{
			"codeblock/more_than_four_leading_space.md",
			blocks{
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},
				{2, "md.ParagraphBlock"},
				{3, "md.CodeBlock"},
			},
		},
		{
			"codeblock/one_blank_line_bw_codeblocks.md",
			blocks{{7, "md.CodeBlock"}},
		},
		{
			"codeblock/one_line.md",
			blocks{{1, "md.CodeBlock"}},
		},
		{
			"codeblock/two_blank_lines_bw_codeblocks.md",
			blocks{
				{3, "md.CodeBlock"},
				{1, "md.NullBlock"},
				{1, "md.NullBlock"},
				{3, "md.CodeBlock"},
			},
		},
		{
			"codeblock/vs_atx_header.md",
			blocks{
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},
				{2, "md.ParagraphBlock"},
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.CodeBlock"},
			},
		},
		{
			"codeblock/vs_blockquote.md",
			blocks{
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},
				{2, "md.ParagraphBlock"},
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.CodeBlock"},
			},
		},
		{
			"codeblock/vs_hr.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},
			},
		},
		{
			"codeblock/vs_list.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{2, "md.CodeBlock"},
				{1, "md.NullBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.CodeBlock"},
				{1, "md.NullBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.CodeBlock"},
				{1, "md.NullBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.CodeBlock"},
				{1, "md.NullBlock"},
			},
		},
		{
			"horizontal_rule/end_with_space.md",
			blocks{
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
			},
		},
		{
			"horizontal_rule/followed_by_block.md",
			blocks{
				{1, "md.HorizontalRuleBlock"},
				{2, "md.ParagraphBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.QuoteBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.OrderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.UnorderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.UnorderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},

				{1, "md.HorizontalRuleBlock"},
				{2, "md.ParagraphBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.QuoteBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.OrderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.UnorderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.UnorderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},

				{1, "md.HorizontalRuleBlock"},
				{2, "md.ParagraphBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.QuoteBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.OrderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.UnorderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.UnorderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},
			},
		},
		{
			"horizontal_rule/loose.md",
			blocks{
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
			},
		},
		{
			"horizontal_rule/sparse.md",
			blocks{
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
			},
		},
		{
			"horizontal_rule/start_with_space.md",
			blocks{
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
			},
		},
		{
			"horizontal_rule/tight.md",
			blocks{
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.HorizontalRuleBlock"},
			},
		},
		{
			"ordered_list/all_items_loose.md",
			blocks{{7, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/all_items_tight.md",
			blocks{{4, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/all_items_tight_even_with_blanks.md",
			blocks{{7, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/at_end_of_parent_without_blank_line.md",
			blocks{
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{8, "md.UnorderedListBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{5, "md.OrderedListBlock"},
			},
		},
		{
			"ordered_list/bw_unordered_lists.md",
			blocks{
				{3, "md.UnorderedListBlock"},
				{3, "md.OrderedListBlock"},
				{3, "md.UnorderedListBlock"},
			},
		},
		{
			"ordered_list/followed_by_hr.md",
			blocks{
				{4, "md.OrderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{5, "md.OrderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{6, "md.OrderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
			},
		},
		{
			"ordered_list/followed_by_list.md",
			blocks{
				{4, "md.OrderedListBlock"},
				{2, "md.UnorderedListBlock"},
				{5, "md.OrderedListBlock"},
				{2, "md.UnorderedListBlock"},
				{6, "md.OrderedListBlock"},
				{1, "md.UnorderedListBlock"},
			},
		},
		{
			"ordered_list/indent_and_sub_blocks.md",
			blocks{
				{6, "md.OrderedListBlock"},
				{2, "md.ParagraphBlock"},
				{6, "md.OrderedListBlock"},
				{2, "md.QuoteBlock"},
				{6, "md.OrderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{6, "md.OrderedListBlock"},
				{2, "md.UnorderedListBlock"},
				{12, "md.OrderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{2, "md.OrderedListBlock"},
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},
			},
		},
		{
			"ordered_list/list_ends_with_2blanks.md",
			blocks{
				{4, "md.OrderedListBlock"},
				{1, "md.NullBlock"},
				{10, "md.OrderedListBlock"},
				{1, "md.NullBlock"},
				{1, "md.CodeBlock"},
			},
		},
		{
			"ordered_list/many_level_nesting.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{22, "md.OrderedListBlock"},
				{1, "md.ParagraphBlock"},
			},
		},
		{
			"ordered_list/many_lines.md",
			blocks{{5, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/many_lines_lazy.md",
			blocks{{5, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/many_paras.md",
			blocks{{9, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/many_paras_2blank.md",
			blocks{
				{5, "md.OrderedListBlock"},
				{1, "md.NullBlock"},
				{4, "md.ParagraphBlock"},
			},
		},
		{
			"ordered_list/many_paras_2blank_lazy.md",
			blocks{
				{5, "md.OrderedListBlock"},
				{1, "md.NullBlock"},
				{4, "md.ParagraphBlock"},
			},
		},
		{
			"ordered_list/many_paras_lazy.md",
			blocks{{9, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/no_space_after_number.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.ParagraphBlock"},
				{1, "md.ParagraphBlock"},
			},
		},
		{
			"ordered_list/no_space_before_number.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{2, "md.OrderedListBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.OrderedListBlock"},
				{2, "md.ParagraphBlock"},
				{1, "md.OrderedListBlock"},
			},
		},
		{
			"ordered_list/numbering_from_two.md",
			blocks{{11, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/numbering_not_in_order.md",
			blocks{{12, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/numbers_left_aligned.md",
			blocks{{12, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/numbers_right_aligned.md",
			blocks{{12, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/numbers_wiggly.md",
			blocks{{12, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/one_line.md",
			blocks{{3, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/some_items_loose.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{6, "md.OrderedListBlock"},
				{2, "md.ParagraphBlock"},
				{8, "md.OrderedListBlock"},
				{2, "md.ParagraphBlock"},
				{10, "md.OrderedListBlock"},
				{2, "md.ParagraphBlock"},
				{6, "md.OrderedListBlock"},
				{1, "md.ParagraphBlock"},
			},
		},
		{
			"ordered_list/space_before_number.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{2, "md.OrderedListBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.OrderedListBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.OrderedListBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.OrderedListBlock"},
				{2, "md.ParagraphBlock"},
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},
			},
		},
		{
			"ordered_list/three_paras_loose.md",
			blocks{{15, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/three_paras_tight.md",
			blocks{{13, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/two_paras_loose.md",
			blocks{{11, "md.OrderedListBlock"}},
		},
		{
			"ordered_list/with_atx_header.md",
			blocks{
				{16, "md.OrderedListBlock"},
				{1, "md.AtxHeaderBlock"},
			},
		},
		{
			"ordered_list/with_blockquote.md",
			blocks{
				{8, "md.OrderedListBlock"},
				{1, "md.QuoteBlock"},
			},
		},
		{
			"ordered_list/with_codeblock.md",
			blocks{
				{8, "md.OrderedListBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.OrderedListBlock"},
				{1, "md.CodeBlock"},
			},
		},
		{
			"ordered_list/with_para.md",
			blocks{
				{10, "md.OrderedListBlock"},
				{1, "md.ParagraphBlock"},
			},
		},
		{
			"ordered_list/with_setext_header.md",
			blocks{
				{25, "md.OrderedListBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{10, "md.OrderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{3, "md.OrderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{3, "md.OrderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{3, "md.OrderedListBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
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
				{3, "md.ParagraphBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{3, "md.ParagraphBlock"},
			},
		},
		{
			"paragraph/followed_by_blockquote.md",
			blocks{
				{3, "md.ParagraphBlock"},
				{3, "md.QuoteBlock"},
				{4, "md.ParagraphBlock"},
			},
		},
		{
			"paragraph/followed_by_codeblock.md",
			blocks{
				{3, "md.ParagraphBlock"},
				{2, "md.CodeBlock"},
				{1, "md.NullBlock"},
				{4, "md.ParagraphBlock"},
			},
		},
		{
			"paragraph/followed_by_horizontal_rule.md",
			blocks{
				{3, "md.ParagraphBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{1, "md.NullBlock"},
				{2, "md.ParagraphBlock"},
				{1, "md.HorizontalRuleBlock"},
			},
		},
		{
			"paragraph/followed_by_list.md",
			blocks{
				{3, "md.ParagraphBlock"},
				{2, "md.UnorderedListBlock"},
				{4, "md.ParagraphBlock"},
				{3, "md.ParagraphBlock"},
				{2, "md.OrderedListBlock"},
				{4, "md.ParagraphBlock"},
			},
		},
		{
			"paragraph/followed_by_setext_header.md",
			blocks{
				{3, "md.ParagraphBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{4, "md.ParagraphBlock"},
			},
		},
		{
			"paragraph/simple_para.md",
			blocks{{3, "md.ParagraphBlock"}},
		},
		{
			"paragraph/two_paras_1blank.md",
			blocks{
				{3, "md.ParagraphBlock"},
				{2, "md.ParagraphBlock"},
			},
		},
		{
			"paragraph/two_paras_2blank.md",
			blocks{
				{3, "md.ParagraphBlock"},
				{1, "md.NullBlock"},
				{2, "md.ParagraphBlock"},
			},
		},
		{
			"setext_header/blank_text.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.ParagraphBlock"},
				{1, "md.HorizontalRuleBlock"},
			},
		},
		{
			"setext_header/enclosed_space_in_underline.md",
			blocks{
				{3, "md.ParagraphBlock"},
				{1, "md.ParagraphBlock"},
				{1, "md.HorizontalRuleBlock"},
			},
		},
		{
			"setext_header/leading_space_in_text.md",
			blocks{
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
			},
		},
		{
			"setext_header/leading_space_in_underline.md",
			blocks{
				{3, "md.ParagraphBlock"},
				{1, "md.ParagraphBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
			},
		},
		{
			"setext_header/simple.md",
			blocks{
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
			},
		},
		{
			"setext_header/span_in_text.md",
			blocks{
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
			},
		},
		{
			"setext_header/trailing_space_in_underline.md",
			blocks{
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
			},
		},
		{
			"setext_header/vs_atx_header.md",
			blocks{
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
			},
		},
		{
			"setext_header/vs_blockquote.md",
			blocks{
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
			},
		},
		{
			"setext_header/vs_codeblock.md",
			blocks{
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
			},
		},
		{
			"setext_header/vs_list.md",
			blocks{
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{2, "md.SetextHeaderBlock"},
			},
		},
		{
			"unordered_list/all_items_loose.md",
			blocks{{7, "md.UnorderedListBlock"}},
		},
		{
			"unordered_list/all_items_tight.md",
			blocks{{4, "md.UnorderedListBlock"}},
		},
		{
			"unordered_list/all_items_tight_even_with_blanks.md",
			blocks{{7, "md.UnorderedListBlock"}},
		},
		{
			"unordered_list/at_end_of_parent_without_blank_line.md",
			blocks{
				{1, "md.NullBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{8, "md.UnorderedListBlock"},
				{1, "md.AtxHeaderBlock"},
				{1, "md.NullBlock"},
				{5, "md.UnorderedListBlock"},
			},
		},
		{
			"unordered_list/bw_ordered_lists.md",
			blocks{
				{3, "md.OrderedListBlock"},
				{3, "md.UnorderedListBlock"},
				{3, "md.OrderedListBlock"},
			},
		},
		{
			"unordered_list/changing_bullet.md",
			blocks{
				{3, "md.UnorderedListBlock"},
				{3, "md.UnorderedListBlock"},
				{3, "md.UnorderedListBlock"},
			},
		},
		{
			"unordered_list/changing_starter_string.md",
			blocks{
				{4, "md.UnorderedListBlock"},
				{2, "md.UnorderedListBlock"},
				{4, "md.UnorderedListBlock"},
				{2, "md.UnorderedListBlock"},
				{1, "md.UnorderedListBlock"},
				{2, "md.OrderedListBlock"},
				{1, "md.UnorderedListBlock"},
			},
		},
		{
			"unordered_list/different_bullet_chars.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{2, "md.UnorderedListBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.UnorderedListBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.UnorderedListBlock"},
			},
		},
		{
			"unordered_list/followed_by_hr.md",
			blocks{
				{4, "md.UnorderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{5, "md.UnorderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{6, "md.UnorderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
			},
		},
		{
			"unordered_list/followed_by_list.md",
			blocks{
				{4, "md.UnorderedListBlock"},
				{2, "md.UnorderedListBlock"},
				{5, "md.UnorderedListBlock"},
				{2, "md.UnorderedListBlock"},
				{6, "md.UnorderedListBlock"},
				{1, "md.UnorderedListBlock"},

				{4, "md.UnorderedListBlock"},
				{2, "md.OrderedListBlock"},
				{5, "md.UnorderedListBlock"},
				{2, "md.OrderedListBlock"},
				{6, "md.UnorderedListBlock"},
				{1, "md.OrderedListBlock"},
			},
		},
		{
			"unordered_list/indent_and_sub_blocks.md",
			blocks{
				{6, "md.UnorderedListBlock"},
				{2, "md.ParagraphBlock"},
				{6, "md.UnorderedListBlock"},
				{2, "md.QuoteBlock"},
				{6, "md.UnorderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{6, "md.UnorderedListBlock"},
				{2, "md.UnorderedListBlock"},
				{6, "md.UnorderedListBlock"},
				{2, "md.OrderedListBlock"},
				{6, "md.UnorderedListBlock"},
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},
			},
		},
		{
			"unordered_list/list_ends_with_2blanks.md",
			blocks{
				{4, "md.UnorderedListBlock"},
				{1, "md.NullBlock"},
				{10, "md.UnorderedListBlock"},
				{1, "md.NullBlock"},
				{1, "md.CodeBlock"},
			},
		},
		{
			"unordered_list/many_level_nesting.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{22, "md.UnorderedListBlock"},
				{1, "md.ParagraphBlock"},
			},
		},
		{
			"unordered_list/many_lines.md",
			blocks{{5, "md.UnorderedListBlock"}},
		},
		{
			"unordered_list/many_lines_lazy.md",
			blocks{{5, "md.UnorderedListBlock"}},
		},
		{
			"unordered_list/many_paras.md",
			blocks{{9, "md.UnorderedListBlock"}},
		},
		{
			"unordered_list/many_paras_2blank.md",
			blocks{
				{5, "md.UnorderedListBlock"},
				{1, "md.NullBlock"},
				{4, "md.ParagraphBlock"},
			},
		},
		{
			"unordered_list/many_paras_2blank_lazy.md",
			blocks{
				{5, "md.UnorderedListBlock"},
				{1, "md.NullBlock"},
				{4, "md.ParagraphBlock"},
			},
		},
		{
			"unordered_list/many_paras_lazy.md",
			blocks{{9, "md.UnorderedListBlock"}},
		},
		{
			"unordered_list/no_space_after_bullet.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.ParagraphBlock"},
				{1, "md.ParagraphBlock"},
			},
		},
		{
			"unordered_list/no_space_before_bullet.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{2, "md.UnorderedListBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.UnorderedListBlock"},
				{2, "md.ParagraphBlock"},
				{1, "md.UnorderedListBlock"},
			},
		},
		{
			"unordered_list/one_line.md",
			blocks{{3, "md.UnorderedListBlock"}},
		},
		{
			"unordered_list/some_items_loose.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{6, "md.UnorderedListBlock"},
				{2, "md.ParagraphBlock"},
				{8, "md.UnorderedListBlock"},
				{2, "md.ParagraphBlock"},
				{10, "md.UnorderedListBlock"},
				{2, "md.ParagraphBlock"},
				{6, "md.UnorderedListBlock"},
				{1, "md.ParagraphBlock"},
			},
		},
		{
			"unordered_list/space_before_bullet.md",
			blocks{
				{2, "md.ParagraphBlock"},
				{2, "md.UnorderedListBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.UnorderedListBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.UnorderedListBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.UnorderedListBlock"},
				{2, "md.ParagraphBlock"},
				{1, "md.CodeBlock"},
				{1, "md.NullBlock"},
			},
		},
		{
			"unordered_list/three_paras_loose.md",
			blocks{{15, "md.UnorderedListBlock"}},
		},
		{
			"unordered_list/three_paras_tight.md",
			blocks{{13, "md.UnorderedListBlock"}},
		},
		{
			"unordered_list/two_paras_loose.md",
			blocks{{11, "md.UnorderedListBlock"}},
		},
		{
			"unordered_list/with_atx_header.md",
			blocks{
				{16, "md.UnorderedListBlock"},
				{1, "md.AtxHeaderBlock"},
			},
		},
		{
			"unordered_list/with_blockquote.md",
			blocks{
				{8, "md.UnorderedListBlock"},
				{1, "md.QuoteBlock"},
			},
		},
		{
			"unordered_list/with_codeblock.md",
			blocks{
				{8, "md.UnorderedListBlock"},
				{2, "md.ParagraphBlock"},
				{2, "md.UnorderedListBlock"},
				{1, "md.CodeBlock"},
			},
		},
		{
			"unordered_list/with_para.md",
			blocks{
				{10, "md.UnorderedListBlock"},
				{1, "md.ParagraphBlock"},
			},
		},
		{
			"unordered_list/with_setext_header.md",
			blocks{
				{25, "md.UnorderedListBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
				{10, "md.UnorderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{3, "md.UnorderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{3, "md.UnorderedListBlock"},
				{1, "md.HorizontalRuleBlock"},
				{1, "md.NullBlock"},
				{3, "md.UnorderedListBlock"},
				{2, "md.SetextHeaderBlock"},
				{1, "md.NullBlock"},
			},
		},
	}

Cases:
	for _, c := range cases {
		// test.Logf("case %s", c.path)
		data, err := ioutil.ReadFile(filepath.Join(dir, c.path))
		if err != nil {
			test.Error(err)
			continue
		}
		prep, _ := vfmd.QuickPrep(bytes.NewReader(data))
		result, err := QuickParse(bytes.NewReader(prep), TopBlocks, nil, nil)

		if err != nil {
			test.Error(err)
			continue Cases
		}

		summary := blocks{}
		for _, b := range result {
			if _, ok := b.(md.End); ok {
				continue
			}
			type getRawer interface {
				GetRaw() md.Region
			}
			if b, ok := b.(getRawer); ok {
				summary = append(summary, block{
					n:   len(b.GetRaw()),
					tag: fmt.Sprintf("%T", b),
				})
			}
		}

		if !reflect.DeepEqual(summary, c.blocks) {
			test.Errorf("case %s expected vs. got DIFF:\n%s",
				c.path, diff.Diff(spew.Sdump(c.blocks), spew.Sdump(summary)))
		}
	}
}
