package block

import (
	"bytes"
	"regexp"
)

type UnorderedList struct {
	// Starter []byte
}
type Item struct{}

var reUnorderedList = regexp.MustCompile(`^( *[\*\-\+] +)[^ ]`)

func DetectUnorderedList(start, second Line, detectors Detectors) Handler {
	m := reUnorderedList.FindSubmatch(start.Bytes)
	if m == nil {
		return nil
	}
	starter := m[1]
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
			ctx.Emit(UnorderedList{})
			ctx.Emit(Item{})
			parser = &Parser{
				Context:   ctx,
				Detectors: detectors,
			}
			return pass(parser, next, next.Bytes[len(starter):])
		}

		if prev.isBlank() {
			if next.isBlank() {
				return end(parser, ctx)
			}
			if !bytes.HasPrefix(next.Bytes, starter) &&
				// FIXME(akavel): spec refers to runes ("characters"), not bytes; fix this everywhere
				next.hasNonSpaceInPrefix(len(starter)) {
				return end2(parser, ctx)
			}
		} else {
			if !bytes.HasPrefix(next.Bytes, starter) &&
				next.hasNonSpaceInPrefix(len(starter)) &&
				!next.hasFourSpacePrefix() &&
				(reUnorderedList.Match(next.Bytes) ||
					reOrderedList.Match(next.Bytes) ||
					reHorizontalRule.Match(next.Bytes)) {
				return end2(parser, ctx)
			}
		}
		if bytes.HasPrefix(next.Bytes, starter) {
			_, err := end(parser, ctx)
			if err != nil {
				return false, err
			}
			ctx.Emit(Item{})
			parser = &Parser{
				Context:   ctx,
				Detectors: detectors,
			}
			return pass(parser, next, next.Bytes[len(starter):])
		}
		return pass(parser, next, trimLeftN(next.Bytes, " ", len(starter)))
	})
}
