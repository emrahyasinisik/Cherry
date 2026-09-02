package factory

import (
	"strconv"
	"strings"
	"unicode"
)

func dartPackage(slug string) string {
	s := strings.ReplaceAll(slug, "-", "_")
	if s == "" {
		return "uygulama"
	}
	r := []rune(s)
	if unicode.IsDigit(r[0]) {
		return "app_" + s
	}
	return s
}

func pascalIdent(slug string) string {
	parts := strings.FieldsFunc(slug, func(r rune) bool {
		return r == '-' || r == '_'
	})
	var b strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(strings.ToLower(part))
		runes[0] = unicode.ToUpper(runes[0])
		b.WriteString(string(runes))
	}
	out := b.String()
	if out == "" {
		return "App"
	}
	return out
}

func jsonStr(s string) string {
	return strconv.Quote(s)
}

func yamlStr(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	return strconv.Quote(s)
}

func dartStr(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `\'`)
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func swiftStr(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}
