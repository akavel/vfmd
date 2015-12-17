package mdblock

import (
	"bytes"
	"regexp"

	"gopkg.in/akavel/vfmd.v1/md"
)

var reHorizontalRule = regexp.MustCompile(`^ *((\* *\* *\* *[\* ]*)|(\- *\- *\- *[\- ]*)|(_ *_ *_ *[_ ]*))$`)

func DetectHorizontalRule(first, second Line, detectors Detectors) Handler {
	if !reHorizontalRule.Match(bytes.TrimRight(first.Bytes, "\n")) {
		return nil
	}
	done := false
	return HandlerFunc(func(next Line, ctx Context) (bool, error) {
		if done {
			return false, nil
		}
		done = true
		ctx.Emit(md.HorizontalRuleBlock{
			Raw: md.Raw{md.Run(next)},
		})
		ctx.Emit(md.End{})
		return true, nil
	})
}
