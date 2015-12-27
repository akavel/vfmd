package mdutils

import (
	"bufio"
	"io"
	"io/ioutil"
	"regexp"
	"unicode"
	"unicode/utf8"

	"gopkg.in/akavel/vfmd.v0/md"
)

func Copy(r md.Region) md.Region {
	return append(md.Region(nil), r...)
}

func FindSubmatch(r md.Region, p *regexp.Regexp) []md.Region {
	// FIXME(akavel): verify if below func returns byte offsets, or rune indexes
	rr := regionReader{r: r}
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

type regionReader struct {
	r  md.Region
	in int
}

func (r *regionReader) Read(buf []byte) (int, error) {
Retry:
	if len(r.r) == 0 {
		return 0, io.EOF
	}
	run := r.r[0].Bytes[r.in:]
	if len(run) == 0 {
		r.r = r.r[1:]
		r.in = 0
		goto Retry
	}
	n := copy(buf, run)
	if n < len(run) {
		r.in += n
	} else {
		r.r = r.r[1:]
		r.in = 0
	}
	return n, nil
}

func String(r md.Region) string {
	rr := regionReader{r: r}
	buf, err := ioutil.ReadAll(&rr)
	if err != nil {
		panic(err.Error())
	}
	return string(buf)
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
		n -= len(run.Bytes)
	}
}

func Len(r md.Region) int {
	n := 0
	for _, run := range r {
		n += len(run.Bytes)
	}
	return n
}

func DecodeRune(r md.Region) (ch rune, size int) {
	rr := bufio.NewReader(&regionReader{r: r})
	ch, size, err := rr.ReadRune()
	if err != nil {
		return utf8.RuneError, 0
	}
	if ch == unicode.ReplacementChar {
		return utf8.RuneError, 1
	}
	return ch, size
}

func DecodeLastRune(r md.Region) (ch rune, size int) {
	rr := regionReverser{r: r}
	buf := [utf8.UTFMax]byte{}
	for i := len(buf) - 1; i >= 0; i-- {
		b, err := rr.ReadByte()
		if err != nil {
			return utf8.RuneError, 1
		}
		buf[i] = b
		ch, size = utf8.DecodeRune(buf[i:])
		if ch != utf8.RuneError {
			return ch, size
		}
	}
	return utf8.RuneError, 1
}

type regionReverser struct {
	r      md.Region
	suffix int
}

func (r *regionReverser) ReadByte() (byte, error) {
Retry:
	if len(r.r) == 0 {
		return 0, io.EOF
	}
	n := len(r.r)
	run := r.r[n-1].Bytes
	if len(run) == r.suffix {
		r.r = r.r[:n-1]
		r.suffix = 0
		goto Retry
	}
	r.suffix++
	return run[len(run)-r.suffix], nil
}
