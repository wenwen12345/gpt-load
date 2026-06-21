package channel

import (
	"net/http"
	"net/url"
	"strings"
)

var anthropicTierByRequestsLimit = map[int64]string{
	50:   "T1",
	1000: "T2",
	2000: "T3",
	4000: "T4",
}

func buildAnthropicValidationResult(endpoint *url.URL, model string, headers http.Header) KeyValidationResult {
	result := KeyValidationResult{
		IsValid:           true,
		TierProvider:      "anthropic",
		TierModel:         model,
		RequestsLimit:     headers.Get("anthropic-ratelimit-requests-limit"),
		InputTokensLimit:  headers.Get("anthropic-ratelimit-input-tokens-limit"),
		OutputTokensLimit: headers.Get("anthropic-ratelimit-output-tokens-limit"),
		TokensLimit:       headers.Get("anthropic-ratelimit-tokens-limit"),
	}
	if endpoint != nil {
		result.TierHost = endpoint.Hostname()
	}

	if !isOfficialAnthropicAPI(endpoint) {
		result.TierReason = "not_official_anthropic"
		return result
	}

	result.TierUpdated = true
	result.Tier = inferAnthropicTierFromHeaders(headers)
	result.TierReason = anthropicTierReason(result.Tier, headers)
	return result
}

func isOfficialAnthropicAPI(endpoint *url.URL) bool {
	if endpoint == nil {
		return false
	}
	return strings.EqualFold(endpoint.Hostname(), "api.anthropic.com")
}

func inferAnthropicTierFromHeaders(headers http.Header) string {
	requests, ok := parseOpenAIRateLimitHeader(headers.Get("anthropic-ratelimit-requests-limit"))
	if !ok {
		return ""
	}
	return anthropicTierByRequestsLimit[requests]
}

func anthropicTierReason(tier string, headers http.Header) string {
	if tier != "" {
		return "inferred"
	}
	requestsLimit := strings.TrimSpace(headers.Get("anthropic-ratelimit-requests-limit"))
	if requestsLimit == "" {
		return "missing_requests_header"
	}
	if _, ok := parseOpenAIRateLimitHeader(requestsLimit); !ok {
		return "invalid_requests_header"
	}
	return "unmapped_limits"
}
