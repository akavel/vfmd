package mdutils // import "gopkg.in/akavel/vfmd.v1/mdutils"

import (
	"bytes"
	"strings"
	"unicode"
	"unicode/utf8"

	"gopkg.in/akavel/vfmd.v1/md"
)

// Whites contains all whitespace characters as defined by VFMD specification.
const Whites = "\x09\x0a\x0c\x0d\x20"

func IsWhite(b byte) bool {
	return b == 0x09 || b == 0x0a || b == 0x0c || b == 0x0d || b == 0x20
}

// FIXME(akavel): test if this works as expected
var whitespaceDeleter = strings.NewReplacer("\u0009", "",
	"\u000a", "",
	"\u000c", "",
	"\u000d", "",
	"\u0020", "")

func DelWhites(s string) string {
	return whitespaceDeleter.Replace(s)
}

func Simplify(buf []byte) string {
	// FIXME(akavel): write tests for Simplify
	out := []byte{}
	// trim left + shorten multiple whitespace
	drop := true
	for _, b := range buf {
		switch {
		case !IsWhite(b):
			out = append(out, b)
			drop = false
		case !drop:
			out = append(out, ' ')
			drop = true
		default:
		}
	}
	// trim right
	if len(out) > 0 && out[len(out)-1] == ' ' {
		out = out[:len(out)-1]
	}
	return string(out)
}

func OffsetIn(s, span []byte) (int, bool) {
	// one weird trick to check if one of two slices is subslice of the other
	bigS := s[:cap(s)]
	bigSpan := span[:cap(span)]
	if &bigSpan[cap(bigSpan)-1] != &bigS[cap(bigS)-1] {
		return -1, false
	}
	if len(bigS) < len(bigSpan) {
		return -1, false
	}
	return len(bigS) - len(bigSpan), true
}

func DeEscape(s string) string {
	buf := bytes.NewBuffer(nil)
	esc := false
	for _, c := range []rune(s) {
		if esc {
			if !unicode.IsPunct(c) && !unicode.IsSymbol(c) {
				buf.WriteByte('\\')
			}
			buf.WriteRune(c)
			esc = false
			continue
		}
		if c != '\\' {
			buf.WriteRune(c)
			continue
		}
		esc = true
	}
	if esc {
		buf.WriteByte('\\')
	}
	return buf.String()
}

func DeEscapeProse(p md.Prose) md.Prose {
	result := make(md.Prose, 0, len(p))
	var buf []byte
runs:
	for i := 0; i < len(p); i++ {
		if buf == nil {
			buf = p[i].Bytes
		}
		for j := 0; ; {
			k := bytes.IndexByte(buf[j:], '\\')
			if k == -1 {
				result = append(result, md.Run{
					Line:  p[i].Line,
					Bytes: buf,
				})
				buf = nil
				continue runs
			}
			j += k
			r, _ := utf8.DecodeRune(buf[j+1:])
			if unicode.IsPunct(r) || unicode.IsSymbol(r) {
				result = append(result, md.Run{
					Line:  p[i].Line,
					Bytes: buf[:j],
				})
				buf = buf[j+1:]
				i--
				continue runs
			}
			j++
		}
	}
	return result
}
