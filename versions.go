package unsafehttp

import "strings"

// SupportedVersions
type SupportedVersion string

const (
	// HTTP1
	HTTP1 SupportedVersion = "HTTP/1.0"
	// HTTP1_1
	HTTP1_1 SupportedVersion = "HTTP/1.1"
)

func httpVersionFromString(s string) (SupportedVersion, bool) {
	switch strings.ToLower(s) {
	case "http/1.0":
		return HTTP1, true
	case "http/1.1":
		return HTTP1_1, true
	default:
		return "", false
	}
}
