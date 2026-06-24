package channel

import (
	"testing"
)

func TestNormalizeURLMode(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "host"},
		{"host", "host"},
		{"HOST", "host"},
		{"prefix", "prefix"},
		{"Prefix", "prefix"},
		{"full", "full"},
		{"FULL", "full"},
		{"unknown", "host"},
		{"  prefix  ", "prefix"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := normalizeURLMode(tt.in); got != tt.want {
				t.Fatalf("normalizeURLMode(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCombineUpstreamPath(t *testing.T) {
	tests := []struct {
		name         string
		upstreamPath string
		requestPath  string
		urlMode      string
		want         string
	}{
		// host mode (default): verbatim append
		{"host bare + request", "", "/v1/chat/completions", "host", "/v1/chat/completions"},
		{"host with trailing slash", "/", "/v1/chat/completions", "host", "/v1/chat/completions"},
		{"host with path", "/api", "/v1/chat/completions", "host", "/api/v1/chat/completions"},
		{"host empty mode falls through", "", "/v1/chat/completions", "", "/v1/chat/completions"},

		// prefix mode: dedup common leading segments
		{"prefix dedup full overlap", "/v1", "/v1/chat/completions", "prefix", "/v1/chat/completions"},
		{"prefix dedup partial", "/v1/ai", "/v1/chat/completions", "prefix", "/v1/ai/chat/completions"},
		{"prefix no overlap", "/custom", "/v1/chat/completions", "prefix", "/custom/v1/chat/completions"},
		{"prefix upstream empty", "", "/v1/chat/completions", "prefix", "/v1/chat/completions"},
		{"prefix request empty", "/v1", "", "prefix", "/v1"},
		{"prefix both empty", "", "", "prefix", ""},
		{"prefix trailing slash upstream", "/v1/", "/v1/chat/completions", "prefix", "/v1/chat/completions"},
		{"prefix trailing slash request", "/v1", "/v1/chat/completions/", "prefix", "/v1/chat/completions"},
		{"prefix api v1", "/api/v1", "/api/v1/chat/completions", "prefix", "/api/v1/chat/completions"},

		// full mode: use upstream verbatim, ignore request
		{"full uses upstream only", "/v1/chat/completions", "/v1/chat/completions", "full", "/v1/chat/completions"},
		{"full ignores request", "/v1/ai/chat", "/v1/chat/completions", "full", "/v1/ai/chat"},
		{"full with trailing slash", "/v1/chat/completions/", "/anything", "full", "/v1/chat/completions"},
		{"full empty upstream", "", "/v1/chat/completions", "full", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CombineUpstreamPath(tt.upstreamPath, tt.requestPath, tt.urlMode); got != tt.want {
				t.Fatalf("CombineUpstreamPath(%q, %q, %q) = %q, want %q",
					tt.upstreamPath, tt.requestPath, tt.urlMode, got, tt.want)
			}
		})
	}
}
