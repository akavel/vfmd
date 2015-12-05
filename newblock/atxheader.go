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
		a := md.AtxHeaderBlock{}
		text := bytes.Trim(line.Bytes, "#")
		if len(text) > 0 {
			a.Level, _ = utils.OffsetIn(line.Bytes, text)
		}
		if a.Level > 6 {
			a.Level = 6
		}
		ctx.Emit(a)
		// TODO(akavel): ctx.Emit(spans & text contents)
		ctx.Emit(md.EndBlock{})
		return true, nil
	})
}
