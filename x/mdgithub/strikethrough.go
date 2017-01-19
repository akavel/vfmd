// Package mdgithub provides Github-flavored Markdown extensions to vfmd.
// Not all are implemented yet. On the other hand, some are already implemented
// by base vfmd.
//
// Reference: https://help.github.com/articles/github-flavored-markdown/
package mdgithub

import (
	"unicode"
	"unicode/utf8"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/mdspan"
	"gopkg.in/akavel/vfmd.v0/mdutils"
	"gopkg.in/akavel/vfmd.v0/x/mdhtml"
)

type StrikeThrough struct{}

func (StrikeThrough) Detect(ctx *mdspan.Context) (consumed int) {
	if !mdutils.HasPrefix(ctx.Suffix, []byte("~~")) {
		return 0
	}

	// TODO(akavel): make this work on word fringes, as md.Emphasis?
	leftEdge, rightEdge := false, false
	if len(ctx.Prefix) == 0 {
		leftEdge = true
	} else {
		prev, _ := mdutils.DecodeLastRune(ctx.Prefix)
		leftEdge = unicode.IsSpace(prev)
		if prev == '~' {
			return 0
		}
	}
	more := mdutils.CopyN(ctx.Suffix, 2+utf8.UTFMax)
	mdutils.Skip(&more, 2)
	if mdutils.Empty(more) {
		rightEdge = true
	} else {
		// TODO(akavel): decode full rune
		next, _ := mdutils.DecodeRune(more)
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
			Pos: ctx.Suffix,
		})
		return 2
	}
	// rightEdge; find matching leftEdge
	open, ok := ctx.Openings.PopTo(func(o *mdspan.MaybeOpening) bool {
		return o.Tag == "~~"
	})
	if !ok {
		return 0
	}
	ctx.Emit(mdutils.CopyN(open.Pos, 2), StrikeThrough{}, false)
	ctx.Emit(mdutils.CopyN(ctx.Suffix, 2), md.End{}, false)
	return 2
}

func (s StrikeThrough) HTMLSpan(ctx mdhtml.Context, opt mdhtml.Opt) ([]md.Tag, error) {
	ctx.Printf("<del>")
	ctx.Spans(ctx.Tags[1:], opt)
	ctx.Printf("</del>")
	return ctx.Tags, ctx.Err
}
