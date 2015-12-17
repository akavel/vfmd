package mdblock

import (
	"bytes"

	"gopkg.in/akavel/vfmd.v1/md"
	"gopkg.in/akavel/vfmd.v1/mdutils"
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
		text := bytes.TrimRight(line.Bytes, "\n")
		text = bytes.Trim(text, "#")
		if len(text) > 0 {
			block.Level, _ = mdutils.OffsetIn(line.Bytes, text)
		} else {
			block.Level = len(bytes.TrimRight(line.Bytes, "\n"))
		}
		if block.Level > 6 {
			block.Level = 6
		}

		spanRegion := md.Raw{md.Run{
			Line:  line.Line,
			Bytes: bytes.Trim(text, mdutils.Whites),
		}}
		ctx.Emit(block)
		parseSpans(spanRegion, ctx)
		ctx.Emit(md.End{})
		return true, nil
	})
}
