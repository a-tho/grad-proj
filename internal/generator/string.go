package generator

import (
	"unicode"
	"unicode/utf8"
)

func toLowerFirst(s string) string {
	if s == "" {
		return ""
	}
	firstRune, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(firstRune)) + s[size:]
}

func toUpperFirst(s string) string {
	if s == "" {
		return ""
	}
	firstRune, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(firstRune)) + s[size:]
}
