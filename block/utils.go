package block

func IsWhite(b byte) bool {
	return b == 0x09 || b == 0x0a || b == 0x0c || b == 0x0d || b == 0x20
}

func Simplify(buf []byte) []byte {
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
	return out
}

func HasQuotedStringPrefix(s string) string {
	// FIXME(akavel): write tests for HasQuotedStringPrefix
	switch {
	case len(s) < 2:
		return ""
	case s[0] != '"' && s[0] != '\'':
		return ""
	}
	q := s[0]
	for i := 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			// skip next char, it's escaped
			i++
		case q:
			return s[0 : i+1]
		}
	}
	return ""
}
