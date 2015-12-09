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
	block := md.ParagraphBlock{}
	return HandlerFunc(func(next Line, ctx Context) (bool, error) {
		if next.EOF() {
			return p.close(block, ctx)
		}
		if len(block.Raw) == 0 {
			block.Raw = append(block.Raw, md.Run(next))
			return true, nil
		}
		prev := Line(block.Raw[len(block.Raw)-1])
		// TODO(akavel): support HTML parser & related interactions [#paragraph-line-sequence]
		if prev.isBlank() {
			return p.close(block, ctx)
		}
		nextBytes := bytes.TrimRight(next.Bytes, "\n")
		if !next.hasFourSpacePrefix() {
			if reHorizontalRule.Match(nextBytes) ||
				(p.InQuote && bytes.HasPrefix(bytes.TrimLeft(next.Bytes, " "), []byte(">"))) ||
				(p.InList && reOrderedList.Match(nextBytes)) ||
				(p.InList && reUnorderedList.Match(nextBytes)) {
				return p.close(block, ctx)
			}
		}
		block.Raw = append(block.Raw, md.Run(next))
		return true, nil
	})
}

func (ParagraphDetector) close(block md.ParagraphBlock, ctx Context) (bool, error) {
	ctx.Emit(block)
	parseSpans(block.Raw, ctx)
	ctx.Emit(md.End{})
	return false, nil
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
