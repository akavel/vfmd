package mdblock

import (
	"bytes"
	"regexp"
	"strings"

	"gopkg.in/akavel/vfmd.v0/md"
	"gopkg.in/akavel/vfmd.v0/mdutils"
)

// TODO(akavel): name below regexps better
var (
	reRefResolution1 = regexp.MustCompile(`^ *\[(([^\\\[\]\!]|\\.|\![^\[])*((\!\[([^\\\[\]]|\\.)*\](\[([^\\\[\]]|\\.)*\])?)?([^\\\[\]]|\\.)*)*)\] *:(.*)$`)
	reRefResolution2 = regexp.MustCompile(`^ *([^ \<\>]+|\<[^\<\>]*\>)( .*)?$`)
	reRefResolution3 = regexp.MustCompile(`^ +("(([^"\\]|\\.)*)"|'(([^'\\]|\\.)*)'|\(([^\\\(\)]|\\.)*\)) *$`)
	reRefResolution4 = regexp.MustCompile(`^\((([^\\\(\)]|\\.)*)\)`)
)

func DetectReferenceResolution(start, second Line, detectors Detectors) Handler {
	if start.hasFourSpacePrefix() {
		return nil
	}
	m := reRefResolution1.FindSubmatch(bytes.TrimRight(start.Bytes, "\n"))
	if len(m) == 0 {
		return nil
	}
	b := md.ReferenceResolutionBlock{}
	unprocessedReferenceID := m[1]
	b.ReferenceID = mdutils.Simplify(unprocessedReferenceID)
	refValueSequence := m[9] // TODO(akavel): verify if right one
	m = reRefResolution2.FindSubmatch(refValueSequence)
	if len(m) == 0 {
		return nil
	}
	unprocessedURL := m[1]
	{
		tmp := make([]byte, 0, len(unprocessedURL))
		for _, c := range unprocessedURL {
			if c != ' ' && c != '<' && c != '>' {
				tmp = append(tmp, c)
			}
		}
		b.URL = string(tmp)
	}
	refDefinitionTrailingSequence := m[2]

	// Detected ok. Now check if 1 or 2 lines.
	var nlines int
	titleContainer := ""
	if bytes.IndexAny(refDefinitionTrailingSequence, " ") == -1 &&
		!second.EOF() &&
		reRefResolution3.Match(bytes.TrimRight(second.Bytes, "\n")) {
		nlines = 2
		titleContainer = string(bytes.TrimRight(second.Bytes, "\n"))
	} else {
		nlines = 1
		titleContainer = string(refDefinitionTrailingSequence)
	}

	// NOTE(akavel): below line seems not in the spec, but seems necessary (refDefinitionTrailingSequence always starts with space IIUC).
	titleContainer = strings.TrimLeft(titleContainer, " ")
	if m := reRefResolution4.FindStringSubmatch(titleContainer); len(m) != 0 {
		b.Title = mdutils.DeEscape(m[1])
	}
	if s := hasQuotedStringPrefix(titleContainer); s != "" {
		b.Title = mdutils.DeEscape(s[1 : len(s)-1])
	}

	return HandlerFunc(func(line Line, ctx Context) (bool, error) {
		if nlines == 0 {
			return false, nil
		}
		nlines--
		b.Raw = append(b.Raw, md.Run(line))
		if nlines == 0 {
			ctx.Emit(b)
			ctx.Emit(md.End{})
		}
		return true, nil
	})
}

func hasQuotedStringPrefix(s string) string {
	// FIXME(akavel): write tests for hasQuotedStringPrefix
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
