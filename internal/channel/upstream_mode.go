package channel

import "strings"

// normalizeURLMode canonicalizes an upstream URL mode value.
// Accepted values (case-insensitive): "host", "prefix", "full".
// Empty or unrecognized values default to "host" (the historical behavior).
func normalizeURLMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "prefix", "full":
		return strings.ToLower(strings.TrimSpace(mode))
	default:
		return "host"
	}
}
