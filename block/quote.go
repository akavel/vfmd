package block

import (
	"bytes"

	"gopkg.in/akavel/vfmd.v0/md"
)

func trimQuote(line []byte) []byte {
	trimmed := bytes.TrimLeft(line, " ")
	if bytes.HasPrefix(trimmed, []byte{'>'}) {
		return bytes.TrimPrefix(trimmed[1:], []byte{' '})
	} else {
		return line
	}
}

func DetectQuote(first, second Line, detectors Detectors) Handler {
	ltrim := bytes.TrimLeft(first.Bytes, " ")
	if len(ltrim) == 0 || ltrim[0] != '>' {
		return nil
	}
	var buf *defaultContext
	block := md.QuoteBlock{}
	var carry *Line
	var parser *Parser
	return HandlerFunc(func(next Line, ctx Context) (bool, error) {
		// TODO(akavel): verify it's coded ok, it was converted from a different approach
		if next.EOF() {
			ctx.Emit(block)
			return quoteEnd(parser, buf, ctx)
		}
		prev := carry
		carry = &next
		// First line?
		if prev == nil {
			buf = &defaultContext{
				mode:          ctx.GetMode(),
				detectors:     ctx.GetDetectors(),
				spanDetectors: ctx.GetSpanDetectors(),
			}
			block.Raw = append(block.Raw, md.Run(next))
			if ctx.GetMode() != TopBlocks {
				parser = &Parser{
					Context: buf,
				}
			}
			return pass(parser, next, trimQuote(next.Bytes))
		}
		if prev.isBlank() {
			if next.isBlank() ||
				next.hasFourSpacePrefix() ||
				bytes.TrimLeft(next.Bytes, " ")[0] != '>' {
				ctx.Emit(block)
				return quoteEnd(parser, buf, ctx)
			}
		} else if !next.hasFourSpacePrefix() &&
			reHorizontalRule.Match(bytes.TrimRight(next.Bytes, "\n")) {
			ctx.Emit(block)
			return quoteEnd(parser, buf, ctx)
		}
		block.Raw = append(block.Raw, md.Run(next))
		return pass(parser, next, trimQuote(next.Bytes))
	})
}

func quoteEnd(parser *Parser, buf *defaultContext, ctx Context) (bool, error) {
	b, err := end(parser, buf)
	for _, t := range buf.tags {
		ctx.Emit(t)
	}
	return b, err
}
