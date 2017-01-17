package mdspan

import (
	"reflect"
	"testing"

	"gopkg.in/akavel/vfmd.v0/md"

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
		comment  string
		context  Context
		expected Context
		consumed int
	}{
		{
			"`<>`:",
			Context{
				Suffix: reg(0, "`<>`:"),
			},
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
	}
	for _, c := range cases {
		consumed := DetectCode(&c.context)
		if consumed != c.consumed {
			test.Errorf("case %q: consumed want %d have %d", c.comment, c.consumed, consumed)
		}
		if !reflect.DeepEqual(c.expected, c.context) {
			test.Errorf("case %q Context want vs have DIFF:\n%s",
				c.comment, diff.Diff(spew.Sdump(c.expected), spew.Sdump(c.context)))
		}
	}
}
