package mdblock

import (
	"bytes"

	"gopkg.in/akavel/vfmd.v1/md"
	"gopkg.in/akavel/vfmd.v1/mdutils"
)

type ParagraphDetector struct {
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
	parseSpans(trim(block.Raw), ctx)
	ctx.Emit(md.End{})
	return false, nil
}

func trim(region md.Raw) md.Raw {
	// rtrim
	for len(region) > 0 {
		n := len(region)
		l := bytes.TrimRight(region[n-1].Bytes, mdutils.Whites)
		if len(l) == 0 {
			region = region[:n-1]
			continue
		}
		if len(l) < len(region[n-1].Bytes) {
			region = append(append(md.Raw{}, region[:n-1]...), md.Run{
				Line:  region[n-1].Line,
				Bytes: l,
			})
		}
		break
	}
	// ltrim
	for len(region) > 0 {
		l := bytes.TrimLeft(region[0].Bytes, mdutils.Whites)
		if len(l) == 0 {
			region = region[1:]
			continue
		}
		if len(l) < len(region[0].Bytes) {
			region = append(md.Raw{{
				Line:  region[0].Line,
				Bytes: l,
			}}, region[1:]...)
		}
		break
	}
	return region
}
