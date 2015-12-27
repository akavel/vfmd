package vfmd // import "gopkg.in/akavel/vfmd.v0"

import (
	"io"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/mdblock"
	"gopkg.in/akavel/vfmd.v0/mdspan"
)

func QuickParse(r io.Reader, mode mdblock.Mode, detectors mdblock.Detectors, spanDetectors []mdspan.Detector) ([]md.Tag, error) {
	return mdblock.QuickParse(r, mode, detectors, spanDetectors)
}
