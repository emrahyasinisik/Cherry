package colabbridge

import (
	"regexp"
	"strings"
)

var trycloudflareURL = regexp.MustCompile(`https://[a-zA-Z0-9-]+\.trycloudflare\.com`)

// ParsePublicURL extracts the first trycloudflare quick-tunnel URL from cloudflared logs.
func ParsePublicURL(logChunk string) string {
	matches := trycloudflareURL.FindAllString(logChunk, -1)
	for _, match := range matches {
		host := strings.TrimPrefix(match, "https://")
		if host == "trycloudflare.com" || strings.HasPrefix(host, "www.") {
			continue
		}
		return match
	}
	return ""
}
