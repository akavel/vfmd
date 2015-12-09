package block

import (
	"bytes"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/utils"
)

func DetectAtxHeader(first, second Line, detectors Detectors) Handler {
	if !bytes.HasPrefix(first.Bytes, []byte("#")) {
		return nil
	}
	done := false
	return HandlerFunc(func(line Line, ctx Context) (bool, error) {
		if done {
			return false, nil
		}
		done = true
		block := md.AtxHeaderBlock{
			Raw: md.Raw{md.Run(line)},
		}
		text := bytes.Trim(line.Bytes, "#")
		if len(text) > 0 {
			block.Level, _ = utils.OffsetIn(line.Bytes, text)
			if block.Level > 6 {
				block.Level = 6
			}
		}

		spanRegion := md.Raw{md.Run{
			Line:  line.Line,
			Bytes: bytes.Trim(text, utils.Whites),
		}}
		ctx.Emit(block)
		parseSpans(spanRegion, ctx)
		ctx.Emit(md.End{})
		return true, nil
	})
}
