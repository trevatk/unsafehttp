package unsafehttp

// SupportedVersions
type SupportedVersion []byte

var (
	// HTTP1
	HTTP1 SupportedVersion = []byte("HTTP/1.0")
	// HTTP1_1
	HTTP1_1 SupportedVersion = []byte("HTTP/1.1")
)
