package block

import "bytes"

type Quote struct {
}

func trimQuote(line []byte) []byte {
	line = bytes.TrimLeft(line, " ")
	line = bytes.TrimPrefix(line, []byte{'>'})
	line = bytes.TrimPrefix(line, []byte{' '})
	return line
}

func DetectQuote(first, second Line, detectors Detectors) Handler {
	ltrim := bytes.TrimLeft(first.Bytes, " ")
	if len(ltrim) == 0 || ltrim[0] != '>' {
		return nil
	}
	var carry *Line
	var parser *Parser
	return HandlerFunc(func(next Line, ctx Context) (bool, error) {
		// TODO(akavel): verify it's coded ok, it was converted from a different approach
		if next.EOF() {
			// bool result will be ignored anyway.
			err := parser.WriteLine(next)
			ctx.Emit(End{})
			return false, err
		}
		prev := carry
		carry = &next
		if prev == nil {
			// First line of block.
			ctx.Emit(Quote{})
			parser = &Parser{
				Context:   ctx,
				Detectors: detectors,
			}
			return true, parser.WriteLine(Line{next.Line, trimQuote(next.Bytes)})
		}
		if prev.isBlank() {
			if next.isBlank() ||
				next.hasFourSpacePrefix() ||
				bytes.TrimLeft(next.Bytes, " ")[0] != '>' {
				return false, nil
			}
		} else if !next.hasFourSpacePrefix() &&
			reHorizontalRule.Match(next.Bytes) {
			return false, nil
		}
		return true, parser.WriteLine(Line{next.Line, trimQuote(next.Bytes)})
	})
}
