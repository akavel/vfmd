package block

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
