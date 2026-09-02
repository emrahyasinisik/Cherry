package gdpr

import (
	"fmt"
	"regexp"
	"strings"
)

type Counts map[string]int

func (c Counts) Total() int {
	n := 0
	for _, v := range c {
		n += v
	}
	return n
}

var (
	emailRe   = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`)
	phoneRe   = regexp.MustCompile(`(?:\+|00)?\d[\d\s().\-]{8,}\d`)
	tcknRe    = regexp.MustCompile(`\b[1-9]\d{10}\b`)
	codeRe    = regexp.MustCompile(`\b\d{6}\b`)
	bearerRe  = regexp.MustCompile(`(?i)bearer\s+[a-z0-9\-._~+/]+=*`)
	apiKeyRe  = regexp.MustCompile(`(?i)\b(?:sk-|re_|ghp_|xox[baprs]-)[a-z0-9\-._]{8,}`)
	otpauthRe = regexp.MustCompile(`otpauth://[^\s]+`)
	secretRe  = regexp.MustCompile(`(?i)(api[_-]?key|secret|password|token)\s*[:=]\s*\S+`)
)

func Redact(input string) (string, Counts) {
	return apply(input, true)
}

func Scan(output string) (string, Counts) {
	return apply(output, false)
}

func apply(input string, codes bool) (string, Counts) {
	counts := Counts{}
	out := input
	out = replaceAll(out, otpauthRe, "[REDACTED_TOTP]", "totp", counts)
	out = replaceAll(out, apiKeyRe, "[REDACTED_KEY]", "key", counts)
	out = replaceAll(out, bearerRe, "[REDACTED_TOKEN]", "token", counts)
	out = replaceAll(out, secretRe, "[REDACTED_SECRET]", "secret", counts)
	out = replaceAll(out, emailRe, "[REDACTED_EMAIL]", "email", counts)
	out = replaceAll(out, tcknRe, "[REDACTED_ID]", "national_id", counts)
	out = replaceAll(out, phoneRe, "[REDACTED_PHONE]", "phone", counts)
	if codes {
		out = replaceAll(out, codeRe, "[REDACTED_CODE]", "code", counts)
	}
	return out, counts
}

func replaceAll(in string, re *regexp.Regexp, repl, kind string, counts Counts) string {
	n := 0
	out := re.ReplaceAllStringFunc(in, func(string) string {
		n++
		return repl
	})
	if n > 0 {
		counts[kind] = counts[kind] + n
	}
	return out
}

func Preview(s string, n int) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func FormatCounts(c Counts) string {
	if c.Total() == 0 {
		return "0"
	}
	parts := make([]string, 0, len(c))
	for k, v := range c {
		parts = append(parts, fmt.Sprintf("%s=%d", k, v))
	}
	return strings.Join(parts, ",")
}
