// +build none

package main

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

func main() {
	fmt.Println("Hello, playground")
	chars := "\\" + `~/*#$()<>`
	for _, c := range chars {
		ranges := []string{}
		for k, v := range unicode.Categories {
			if unicode.Is(v, rune(c)) {
				ranges = append(ranges, k)
			}
		}
		sort.Strings(ranges)
		fmt.Printf("%c %v,%v %s\n",
			c, isWordSep(rune(c)), isSpeculativeURLEnd(rune(c)), strings.Join(ranges, " "))
	}
}

func isWordSep(r rune) bool {
	return unicode.In(r,
		unicode.Zs, unicode.Zl, unicode.Zp,
		unicode.Pc, unicode.Pd, unicode.Ps, unicode.Pe, unicode.Pi, unicode.Pf, unicode.Po,
		unicode.Cc, unicode.Cf)
}
func isSpeculativeURLEnd(r rune) bool {
	return r != '\u002f' && isWordSep(r)
}
