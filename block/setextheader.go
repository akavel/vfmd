package block

import (
	"bytes"
	"regexp"

	"gopkg.in/akavel/vfmd.v0/md"
)

var reSetextHeader = regexp.MustCompile(`^(-+|=+) *$`)

func DetectSetextHeader(first, second Line, detectors Detectors) Handler {
	if second.EOF() {
		return nil
	}
	if !reSetextHeader.Match(bytes.TrimRight(second.Bytes, "\n")) {
		return nil
	}
	block := md.SetextHeaderBlock{}
	switch second.Bytes[0] {
	case '=':
		block.Level = 1
	case '-':
		block.Level = 2
	}
	done := 0
	return HandlerFunc(func(next Line, ctx Context) (bool, error) {
		if done == 2 {
			ctx.Emit(block)
			parseSpans(trim(md.Raw{block.Raw[0]}), ctx)
			ctx.Emit(md.End{})
			return false, nil
		}
		done++
		block.Raw = append(block.Raw, md.Run(next))
		return true, nil
	})
}
