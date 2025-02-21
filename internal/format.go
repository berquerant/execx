package internal

import (
	"regexp"
	"strings"
)

func EscapeQuote(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

func EscapeQuoteAndJoin(ss []string) string {
	r := make([]string, len(ss))
	for i, s := range ss {
		r[i] = EscapeQuote(s)
	}
	return strings.Join(r, " ")
}

var (
	sprintfRegex = regexp.MustCompile(`%\[[^]]+\]`)
)

func unwrapKey(x string) string {
	return strings.TrimSuffix(strings.TrimPrefix(x, "%["), "]")
}

// Sprintf replaces %[KEY] with map[KEY] if exists.
func Sprintf(src string, a map[string]string) string {
	if len(a) == 0 {
		return src
	}

	return sprintfRegex.ReplaceAllStringFunc(src, func(x string) string {
		key := unwrapKey(x)
		if v, ok := a[key]; ok {
			return v
		}
		return x
	})
}

func isNewline(s string) bool {
	return s == "\r" || s == "\r\n" || s == "\n"
}

func IndentN(s string, n int) string {
	var (
		prefix = strings.Repeat(" ", n)
		b      strings.Builder
	)
	for x := range strings.Lines(s) {
		if !isNewline(x) {
			x = prefix + x
		}
		b.WriteString(x)
	}
	return b.String()
}
