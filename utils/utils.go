package utils // import "gopkg.in/akavel/vfmd.v0/utils"

import "strings"

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
