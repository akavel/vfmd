package mdutils

import (
	"bufio"
	"io"
	"regexp"

	"gopkg.in/akavel/vfmd.v0/md"
)

func Copy(r md.Region) md.Region {
	return append(md.Region(nil), r...)
}

func FindSubmatch(r md.Region, p *regexp.Regexp) []md.Region {
	// FIXME(akavel): verify if below func returns byte offsets, or rune indexes
	rr := regionReader(r)
	idx := p.FindReaderSubmatchIndex(bufio.NewReader(&rr))
	if idx == nil {
		return nil
	}
	r = Copy(r)
	regions := make([]md.Region, 0, len(idx)/2)
	skipped := 0
	for i := 0; i < len(idx); i += 2 {
		begin, end := idx[i]-skipped, idx[i+1]-skipped
		Skip(&r, begin)
		skipped += begin
		newr := Copy(r)
		Limit(&newr, end-begin)
		regions = append(regions, newr)
	}
	return regions
}

type regionReader md.Region

func (r *regionReader) Read(buf []byte) (int, error) {
Retry:
	if len(*r) == 0 {
		return 0, io.EOF
	}
	run := &(*r)[0]
	if len(run.Bytes) == 0 {
		*r = (*r)[1:]
		goto Retry
	}
	n := copy(buf, run.Bytes)
	if n < len(run.Bytes) {
		run.Bytes = run.Bytes[n:]
	} else {
		*r = (*r)[1:]
	}
	return n, nil
}

func Skip(r *md.Region, n int) {
	if n < 0 {
		panic("mdutils.Skip negative n")
	}
	for n > 0 {
		run := &(*r)[0]
		if n < len(run.Bytes) {
			run.Bytes = run.Bytes[n:]
			return
		}
		n -= len(run.Bytes)
		*r = (*r)[1:]
	}
}

func Limit(r *md.Region, n int) {
	if n < 0 {
		panic("mdutils.Limit negative n")
	}
	if n == 0 {
		*r = (*r)[:0]
		return
	}
	for i := 0; ; i++ {
		run := &(*r)[i]
		if n <= len(run.Bytes) {
			run.Bytes = run.Bytes[:n]
			*r = (*r)[:i+1]
			return
		}
	}
}
