package mdspan

import (
	"reflect"
	"testing"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/mdutils"

	"github.com/davecgh/go-spew/spew"
	"github.com/kylelemons/godebug/diff"
)

func reg(lines_strings ...interface{}) md.Region {
	result := md.Region{}
	run := md.Run{}
	for i, x := range lines_strings {
		if i&1 == 0 {
			run.Line = x.(int)
		} else {
			run.Bytes = []byte(x.(string))
			result = append(result, run)
		}
	}
	return result
}

func TestDetectCode(test *testing.T) {
	cases := []struct {
		expected Context
		consumed int
	}{
		{
			Context{
				Suffix: reg(0, "`<>`:"),
				Spans: []Span{
					{
						Pos:       reg(0, "`<>`"),
						Tag:       md.Code{Code: []byte("<>")},
						SelfClose: true,
					},
				},
			},
			4,
		},
		{
			Context{
				Suffix: reg(0, "`code span`` ends`"),
				Spans: []Span{
					{
						Pos:       reg(0, "`code span`` ends`"),
						Tag:       md.Code{Code: []byte("code span`` ends")},
						SelfClose: true,
					},
				},
			},
			18,
		},
		{
			Context{
				Suffix: reg(
					0, "`code span\n",
					1, "can span multiple\n",
					2, "lines`\n"),
				Spans: []Span{
					{
						Pos: reg(
							0, "`code span\n",
							1, "can span multiple\n",
							2, "lines`"),
						Tag: md.Code{Code: []byte("code span\n" +
							"can span multiple\n" +
							"lines")},
						SelfClose: true,
					},
				},
			},
			35,
		},
	}
	for _, c := range cases {
		var (
			context = Context{Suffix: mdutils.Copy(c.expected.Suffix)}
			comment = mdutils.String(c.expected.Suffix)
		)
		consumed := DetectCode(&context)
		if consumed != c.consumed {
			test.Errorf("case %q: consumed want %d have %d", comment, c.consumed, consumed)
		}
		if !reflect.DeepEqual(c.expected, context) {
			test.Errorf("case %q Context want vs have DIFF:\n%s",
				comment, diff.Diff(spew.Sdump(c.expected), spew.Sdump(context)))
		}
	}
}
