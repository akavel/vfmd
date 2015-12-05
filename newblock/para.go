package block

import (
	"bytes"

	"gopkg.in/akavel/vfmd.v0/md"
)

type ParagraphDetector struct {
	// FIXME(akavel): below fields must be set appropriately when creating a ParagraphDetector
	InQuote bool
	InList  bool
}

func (p ParagraphDetector) Detect(first, second Line, detectors Detectors) Handler {
	var carry *Line
	return HandlerFunc(func(next Line, ctx Context) (bool, error) {
		if next.EOF() {
			ctx.Emit(md.EndBlock{})
			return false, nil
		}
		prev := carry
		carry = &next
		if prev == nil {
			ctx.Emit(md.ParagraphBlock{})
			return true, nil
		}
		// TODO(akavel): support HTML parser & related interactions [#paragraph-line-sequence]
		if prev.isBlank() {
			ctx.Emit(md.EndBlock{})
			return false, nil
		}
		if !next.hasFourSpacePrefix() {
			if reHorizontalRule.Match(next.Bytes) ||
				(p.InQuote && bytes.HasPrefix(bytes.TrimLeft(next.Bytes, " "), []byte(">"))) ||
				(p.InList && reOrderedList.Match(next.Bytes)) ||
				(p.InList && reUnorderedList.Match(next.Bytes)) {
				ctx.Emit(md.EndBlock{})
				return false, nil
			}
		}
		return true, nil
	})
}

// func (b *Paragraph) PostProcess(line Line) {
// 	if line == nil {
// 		if n := len(b.Spans); n > 0 {
// 			b.Spans[n-1] = bytes.TrimRight(b.Spans[n-1], utils.Whites)
// 		}
// 		return
// 	}

// 	if len(b.Spans) == 0 {
// 		line = bytes.TrimLeft(line, utils.Whites)
// 	}
// 	b.Spans = append(b.Spans, line)
// }
