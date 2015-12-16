package mdblock

import (
	"bytes"

	"gopkg.in/akavel/vfmd.v0/md"
)

func DetectCode(first, second Line, detectors Detectors) Handler {
	if !first.hasFourSpacePrefix() {
		return nil
	}
	block := md.CodeBlock{}
	var paused *Line
	return HandlerFunc(func(next Line, ctx Context) (bool, error) {
		if next.EOF() {
			ctx.Emit(block)
			ctx.Emit(md.End{})
			return maybeNull(paused, ctx)
		}
		// TODO(akavel): verify it's coded ok, it was converted from a different approach
		fourspace := []byte("    ")
		switch {
		// previous was blank, next is not tab-indented. Reject both.
		case paused != nil && !next.hasFourSpacePrefix():
			ctx.Emit(block)
			ctx.Emit(md.End{})
			return maybeNull(paused, ctx)
		case next.isBlank():
			if paused != nil {
				block.Raw = append(block.Raw, md.Run(*paused))
				block.Prose = append(block.Prose, md.Run{
					paused.Line, bytes.TrimPrefix(paused.Bytes, fourspace)})
			}
			paused = &next // note: only case where we pause a line
			return true, nil
		case next.hasFourSpacePrefix():
			if paused != nil {
				block.Raw = append(block.Raw, md.Run(*paused))
				block.Prose = append(block.Prose, md.Run{
					paused.Line, bytes.TrimPrefix(paused.Bytes, fourspace)})
				paused = nil
			}
			block.Raw = append(block.Raw, md.Run(next))
			block.Prose = append(block.Prose, md.Run{
				next.Line, bytes.TrimPrefix(next.Bytes, fourspace)})
			return true, nil
		// next not blank & not indented. End the block.
		default:
			if paused != nil {
				block.Raw = append(block.Raw, md.Run(*paused))
				block.Prose = append(block.Prose, md.Run{
					paused.Line, bytes.TrimPrefix(paused.Bytes, fourspace)})
			}
			ctx.Emit(block)
			ctx.Emit(md.End{})
			return false, nil
		}
	})
}

func maybeNull(paused *Line, ctx Context) (bool, error) {
	if paused != nil {
		ctx.Emit(md.NullBlock{
			Raw: md.Raw{md.Run(*paused)},
		})
		ctx.Emit(md.End{})
	}
	return false, nil
}
