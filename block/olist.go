package block

import (
	"bytes"
	"regexp"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/utils"
)

var reOrderedList = regexp.MustCompile(`^( *([0-9]+)\. +)[^ ]`)

func DetectOrderedList(start, second Line, detectors Detectors) Handler {
	m := reOrderedList.FindSubmatch(start.Bytes)
	if m == nil {
		return nil
	}
	var buf *defaultContext
	block := &md.OrderedListBlock{
		Starter: md.Run{start.Line, m[1]},
		// firstNumber, _ := strconv.Atoi(string(m[2]))
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
		if prev == nil {
			buf = &defaultContext{
				mode:          ctx.GetMode(),
				detectors:     ctx.GetDetectors(),
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

		nextBytes := bytes.TrimRight(next.Bytes, "\n")
		if prev.isBlank() {
			if next.isBlank() {
				return listEnd2(parser, buf, ctx)
			}
			if !reOrderedList.Match(nextBytes) &&
				next.hasNonSpaceInPrefix(len(block.Starter.Bytes)) {
				return listEnd2(parser, buf, ctx)
			}
		} else {
			if !reOrderedList.Match(nextBytes) &&
				next.hasNonSpaceInPrefix(len(block.Starter.Bytes)) &&
				!next.hasFourSpacePrefix() &&
				(reUnorderedList.Match(nextBytes) ||
					reHorizontalRule.Match(nextBytes)) {
				return listEnd2(parser, buf, ctx)
			}
		}

		block.Raw = append(block.Raw, md.Run(next))
		m := reOrderedList.FindSubmatch(next.Bytes)
		if m != nil {
			text := bytes.TrimLeft(m[1], " ")
			spaces, _ := utils.OffsetIn(m[1], text)
			if spaces >= len(block.Starter.Bytes) {
				m = nil
			}
		}
		if m != nil {
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
			return pass(parser, next, next.Bytes[len(m[1]):])
		}
		return pass(parser, next, trimLeftN(next.Bytes, " ", len(block.Starter.Bytes)))
	})
}

func listEnd2(parser *Parser, buf *defaultContext, ctx Context) (bool, error) {
	b, err := end2(parser, buf)
	for _, t := range buf.tags {
		switch t := t.(type) {
		case *md.OrderedListBlock:
			ctx.Emit(*t)
		case *md.UnorderedListBlock:
			ctx.Emit(*t)
		case *md.ItemBlock:
			ctx.Emit(*t)
		default:
			ctx.Emit(t)
		}
	}
	return b, err
}
