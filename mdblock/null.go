package mdblock

import "gopkg.in/akavel/vfmd.v1/md"

func DetectNull(first, second Line, detectors Detectors) Handler {
	if !first.isBlank() {
		return nil
	}
	done := false
	return HandlerFunc(func(line Line, ctx Context) (bool, error) {
		if done {
			return false, nil
		}
		done = true
		ctx.Emit(md.NullBlock{
			Raw: md.Raw{md.Run(line)},
		})
		ctx.Emit(md.End{})
		return true, nil
	})
}
