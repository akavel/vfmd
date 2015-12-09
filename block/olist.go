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
	starter := m[1]
	// firstNumber, _ := strconv.Atoi(string(m[2]))
	var carry *Line
	var parser *Parser
	return HandlerFunc(func(next Line, ctx Context) (bool, error) {
		// func (b *UnorderedList) Continue(paused []Line, next Line) (consume, pause int) {
		if next.EOF() {
			// if carry == nil {
			// 	panic("empty carry")
			// }
			return end2(parser, ctx)
		}
		prev := carry
		carry = &next
		if prev == nil {
			ctx.Emit(md.OrderedListBlock{})
			if ctx.GetMode() != TopBlocks {
				ctx.Emit(md.ItemBlock{})
				parser = &Parser{
					Context: ctx,
				}
			}
			return pass(parser, next, next.Bytes[len(starter):])
		}

		if prev.isBlank() {
			if next.isBlank() {
				return end2(parser, ctx)
			}
			if !reOrderedList.Match(next.Bytes) &&
				next.hasNonSpaceInPrefix(len(starter)) {
				return end2(parser, ctx)
			}
		} else {
			if !reOrderedList.Match(next.Bytes) &&
				next.hasNonSpaceInPrefix(len(starter)) &&
				!next.hasFourSpacePrefix() &&
				(reUnorderedList.Match(next.Bytes) ||
					reHorizontalRule.Match(next.Bytes)) {
				return end2(parser, ctx)
			}
		}

		m := reOrderedList.FindSubmatch(next.Bytes)
		if m != nil {
			text := bytes.TrimLeft(m[1], " ")
			spaces, _ := utils.OffsetIn(m[1], text)
			if spaces >= len(starter) {
				m = nil
			}
		}
		if m != nil {
			if ctx.GetMode() != TopBlocks {
				_, err := end(parser, ctx)
				if err != nil {
					return false, err
				}
				ctx.Emit(md.ItemBlock{})
				parser = &Parser{
					Context: ctx,
				}
			}
			return pass(parser, next, next.Bytes[len(m[1]):])
		}
		return pass(parser, next, trimLeftN(next.Bytes, " ", len(starter)))
	})
}
