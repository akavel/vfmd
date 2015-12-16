package mdblock

import (
	"bytes"
	"regexp"

	"gopkg.in/akavel/vfmd.v0/md"
)

var reUnorderedList = regexp.MustCompile(`^( *[\*\-\+] +)[^ ]`)

func DetectUnorderedList(start, second Line, detectors Detectors) Handler {
	m := reUnorderedList.FindSubmatch(start.Bytes)
	if m == nil {
		return nil
	}
	var buf *defaultContext
	block := &md.UnorderedListBlock{
		Starter: md.Run{start.Line, m[1]},
	}
	var item *md.ItemBlock
	var carry *Line
	var parser *Parser
	return HandlerFunc(func(next Line, ctx Context) (bool, error) {
		if next.EOF() {
			return listEnd2(parser, buf, ctx)
		}
		prev := carry
		carry = &next
		// First line? Init stuff and accept unconditionally, already tested.
		if prev == nil {
			buf = &defaultContext{
				mode:          ctx.GetMode(),
				detectors:     changedParagraphDetector(ctx, false, true),
				spanDetectors: ctx.GetSpanDetectors(),
			}
			block.Raw = append(block.Raw, md.Run(next))
			buf.Emit(block)
			if ctx.GetMode() != TopBlocks {
				item = &md.ItemBlock{}
				item.Raw = append(item.Raw, md.Run(next))
				buf.Emit(item)
				parser = &Parser{
					Context: buf,
				}
			}
			return pass(parser, next, next.Bytes[len(block.Starter.Bytes):])
		}

		if prev.isBlank() {
			if next.isBlank() {
				return listEnd2(parser, buf, ctx)
			}
			if !bytes.HasPrefix(next.Bytes, block.Starter.Bytes) &&
				// FIXME(akavel): spec refers to runes ("characters"), not bytes; fix this everywhere
				next.hasNonSpaceInPrefix(len(block.Starter.Bytes)) {
				return listEnd2(parser, buf, ctx)
			}
		} else {
			nextBytes := bytes.TrimRight(next.Bytes, "\n")
			if !bytes.HasPrefix(next.Bytes, block.Starter.Bytes) &&
				next.hasNonSpaceInPrefix(len(block.Starter.Bytes)) &&
				!next.hasFourSpacePrefix() &&
				(reUnorderedList.Match(nextBytes) ||
					reOrderedList.Match(nextBytes) ||
					reHorizontalRule.Match(nextBytes)) {
				return listEnd2(parser, buf, ctx)
			}
		}

		block.Raw = append(block.Raw, md.Run(next))
		if bytes.HasPrefix(next.Bytes, block.Starter.Bytes) {
			if ctx.GetMode() != TopBlocks {
				_, err := end(parser, buf)
				if err != nil {
					return false, err
				}
				item = &md.ItemBlock{}
				item.Raw = append(item.Raw, md.Run(next))
				buf.Emit(item)
				parser = &Parser{
					Context: buf,
				}
			}
			return pass(parser, next, next.Bytes[len(block.Starter.Bytes):])
		}
		if ctx.GetMode() != TopBlocks {
			item.Raw = append(item.Raw, md.Run(next))
		}
		return pass(parser, next, trimLeftN(next.Bytes, " ", len(block.Starter.Bytes)))
	})
}
