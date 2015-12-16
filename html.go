package vfmd

import (
	"io"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/x/mdhtml"
)

func QuickHTML(w io.Writer, blocks []md.Tag) error {
	return mdhtml.QuickRender(w, blocks)
}
