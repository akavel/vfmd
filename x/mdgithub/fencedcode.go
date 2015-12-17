package mdgithub

import (
	"bytes"
	"html"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/mdblock"
	"gopkg.in/akavel/vfmd.v0/x/mdhtml"
)

type FencedCodeBlock struct {
	// TODO(akavel): Language string
	md.Prose
	md.Raw
}

func (FencedCodeBlock) Detect(first, second mdblock.Line, detectors mdblock.Detectors) mdblock.Handler {
	if !bytes.HasPrefix(first.Bytes, []byte("```")) {
		return nil
	}

	block := FencedCodeBlock{}
	done := false
	return mdblock.HandlerFunc(func(next mdblock.Line, ctx mdblock.Context) (bool, error) {
		if done {
			return false, nil
		}
		if next.EOF() {
			ctx.Emit(block)
			ctx.Emit(md.End{})
			return false, nil
		}
		if len(block.Raw) > 0 {
			// Three backticks, on second on later line - means end of block.
			if bytes.HasPrefix(next.Bytes, []byte("```")) {
				done = true
				ctx.Emit(block)
				ctx.Emit(md.End{})
			} else {
				// Collect all stuff between first and last fenced line into Prose.
				block.Prose = append(block.Prose, md.Run(next))
			}
		}
		block.Raw = append(block.Raw, md.Run(next))
		return true, nil
	})
}

func (b FencedCodeBlock) HTMLBlock(ctx mdhtml.Context, opt mdhtml.Opt) ([]md.Tag, error) {
	ctx.Printf("<pre><code>")
	for _, r := range b.Prose {
		ctx.Printf("%s", html.EscapeString(string(r.Bytes)))
	}
	ctx.Printf("</code></pre>\n")
	// Skip self and subsequent md.End{}
	return ctx.Tags[2:], ctx.Err
}
