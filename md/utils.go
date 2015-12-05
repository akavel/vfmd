package md

import "io"

type Region []Run

func (r Region) Reader(pos int) io.Reader {
	for i := range r {
		if pos < len(r[i].Bytes) {
			return &regionReader{r, i, pos}
		}
		pos -= len(r[i].Bytes)
	}
	panic("bad pos in Region.Reader")
}

// TODO(akavel): rename Run to something prettier?
type Run struct {
	Line  int
	Bytes []byte // with Line, allows to find out position in line
}

// func (r Run) String() string { return string(r.Bytes) }

type Prose Region

func (p Prose) Prose() Region { return Region(p) }

type Proser interface {
	Prose() Region
}

type Tag interface{}

type regionReader struct {
	r      Region
	i, off int
}

func (r *regionReader) Read(b []byte) (int, error) {
	if r.i >= len(r.r) {
		return 0, io.EOF
	}
	if r.off >= len(r.r[r.i].Bytes) {
		r.off = 0
		r.i++
		return r.Read(b)
	}
	n := copy(b, r.r[r.i].Bytes)
	r.off += n
	return n, nil
}
