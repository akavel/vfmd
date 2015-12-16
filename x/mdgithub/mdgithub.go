// Package mdgithub provides Github-flavored Markdown extensions to vfmd.
// Not all are implemented yet. On the other hand, some are already implemented
// by base vfmd.
//
// Reference: https://help.github.com/articles/github-flavored-markdown/
package mdgithub

import (
	"bytes"
	"unicode"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/mdspan"
	"gopkg.in/akavel/vfmd.v0/x/mdhtml"
)

type StrikeThrough struct{}

func (StrikeThrough) Detect(ctx *mdspan.Context) (consumed int) {
	rest := ctx.Buf[ctx.Pos:]
	if !bytes.HasPrefix(rest, []byte("~~")) {
		return 0
	}

	// TODO(akavel): make this work on word fringes, as md.Emphasis?
	leftEdge, rightEdge := false, false
	if ctx.Pos == 0 {
		leftEdge = true
	} else {
		// TODO(akavel): decode full rune
		prev := ctx.Buf[ctx.Pos-1]
		leftEdge = unicode.IsSpace(rune(prev))
		if prev == '~' {
			return 0
		}
	}
	if ctx.Pos == len(ctx.Buf)-1 {
		rightEdge = true
	} else {
		// TODO(akavel): decode full rune
		next := ctx.Buf[ctx.Pos+2]
		rightEdge = unicode.IsSpace(rune(next))
		if next == '~' {
			return 0
		}
	}

	if leftEdge == rightEdge { // both or none
		return 0
	}
	if leftEdge {
		ctx.Openings.Push(mdspan.MaybeOpening{
			Tag: "~~",
			Pos: ctx.Pos,
		})
		return 2
	}
	// rightEdge; find matching leftEdge
	o, ok := ctx.Openings.PopTo(func(o *mdspan.MaybeOpening) bool {
		return o.Tag == "~~"
	})
	if !ok {
		return 0
	}
	ctx.Emit(ctx.Buf[o.Pos:][:2], StrikeThrough{}, false)
	ctx.Emit(ctx.Buf[ctx.Pos:][:2], md.End{}, false)
	return 2
}

func (s StrikeThrough) Span(ctx mdhtml.Context, opt mdhtml.Opt) ([]md.Tag, error) {
	ctx.Printf("<del>")
	ctx.Spans(ctx.Tags[1:], opt)
	ctx.Printf("</del>")
	return ctx.Tags, ctx.Err
}
