package vfmd

import (
	"io"

	"gopkg.in/akavel/vfmd.v1/md"
	"gopkg.in/akavel/vfmd.v1/x/mdhtml"
)

func QuickHTML(w io.Writer, blocks []md.Tag) error {
	return mdhtml.QuickRender(w, blocks)
}
