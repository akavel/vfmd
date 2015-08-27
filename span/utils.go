package span

import "strings"

// FIXME(akavel): test if this works as expected
var whitespaceDeleter = strings.NewReplacer("\u0009", "",
	"\u000a", "",
	"\u000c", "",
	"\u000d", "",
	"\u0020", "")

func DelWhitespace(s string) string {
	return whitespaceDeleter.Replace(s)
}

func Simplify(s string) string {
	panic("NIY")
}
