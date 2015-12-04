package block

import (
	"bytes"

	"gopkg.in/akavel/vfmd.v0/utils"
)

type AtxHeader struct {
	Level int
}

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
		a := AtxHeader{}
		text := bytes.Trim(line.Bytes, "#")
		if len(text) > 0 {
			a.Level, _ = utils.OffsetIn(line.Bytes, text)
		}
		if a.Level > 6 {
			a.Level = 6
		}
		ctx.Emit(a)
		// TODO(akavel): ctx.Emit(spans & text contents)
		ctx.Emit(End{})
		return true, nil
	})
}
