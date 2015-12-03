package block

import "bytes"

type Quote struct {
}

func DetectQuote(first, second Line, detectors Detectors) Handler {
	ltrim := bytes.TrimLeft(first.Bytes, " ")
	if len(ltrim) == 0 || ltrim[0] != '>' {
		return nil
	}
	var carry *Line
	return HandlerFunc(func(next Line, ctx *Context) bool {
		// TODO(akavel): verify it's coded ok, it was converted from a different approach
		if next.EOF() {
			// EOF; returned result will be ignored anyway.
			return false
		}
		prev := carry
		carry = &next
		if prev == nil {
			// First line of block.
			// FIXME(akavel): ctx.Emit(Quote{})
			// FIXME(akavel): start processing sub-blocks...
			return true
		}
		if prev.isBlank() {
			if next.isBlank() ||
				next.hasFourSpacePrefix() ||
				bytes.TrimLeft(next, " ")[0] != '>' {
				return len(paused), 0
			}
		} else if !next.hasFourSpacePrefix() &&
			reHorizontalRule.Match(next) {
			return len(paused), 0
		}
		return len(paused), 1
	})
}

func (q *Quote) PostProcess(line Line) {
	if line == nil {
		// FIXME(akavel): handle error
		_ = q.splitter.Close()
		q.Blocks = q.splitter.Blocks
		return
	}

	text := bytes.TrimLeft(line, " ")
	switch {
	case bytes.HasPrefix(text, []byte("> ")):
		text = text[2:]
	case bytes.HasPrefix(text, []byte(">")):
		text = text[1:]
	}

	if q.splitter.Detectors == nil {
		q.splitter.Detectors = q.Detectors
	}
	// FIXME(akavel): handle error
	// FIXME(akavel): ignore final line if "empty"
	_ = q.splitter.WriteLine(line)
	q.Blocks = q.splitter.Blocks
}
