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
			return end(parser, ctx)
		}
		prev := carry
		carry = &next
		if prev == nil {
			// First line of block.
			ctx.Emit(Quote{})
			parser = &Parser{
				Context: ctx,
			}
			return pass(parser, next, trimQuote(next.Bytes))
		}
		if prev.isBlank() {
			if next.isBlank() ||
				next.hasFourSpacePrefix() ||
				bytes.TrimLeft(next.Bytes, " ")[0] != '>' {
				return end(parser, ctx)
			}
		} else if !next.hasFourSpacePrefix() &&
			reHorizontalRule.Match(next.Bytes) {
			return end(parser, ctx)
		}
		return pass(parser, next, trimQuote(next.Bytes))
	})
}
