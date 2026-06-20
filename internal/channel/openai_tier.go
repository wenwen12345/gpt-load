package channel

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type openAIRateLimitPair struct {
	requests int64
	tokens   int64
}

var openAIRateLimitTiersByFamily = map[string]map[openAIRateLimitPair]string{
	"gpt-4.1": {
		{requests: 500, tokens: 30_000}:        "T1",
		{requests: 5_000, tokens: 450_000}:     "T2",
		{requests: 5_000, tokens: 800_000}:     "T3",
		{requests: 10_000, tokens: 2_000_000}:  "T4",
		{requests: 10_000, tokens: 30_000_000}: "T5",
	},
	"gpt-4.1-small": {
		{requests: 500, tokens: 200_000}:        "T1",
		{requests: 5_000, tokens: 2_000_000}:    "T2",
		{requests: 5_000, tokens: 4_000_000}:    "T3",
		{requests: 10_000, tokens: 10_000_000}:  "T4",
		{requests: 30_000, tokens: 150_000_000}: "T5",
	},
	"gpt-5": {
		{requests: 500, tokens: 500_000}:       "T1",
		{requests: 5_000, tokens: 1_000_000}:   "T2",
		{requests: 5_000, tokens: 2_000_000}:   "T3",
		{requests: 10_000, tokens: 4_000_000}:  "T4",
		{requests: 15_000, tokens: 40_000_000}: "T5",
	},
	"gpt-5-mini": {
		{requests: 500, tokens: 500_000}:        "T1",
		{requests: 5_000, tokens: 2_000_000}:    "T2",
		{requests: 5_000, tokens: 4_000_000}:    "T3",
		{requests: 10_000, tokens: 10_000_000}:  "T4",
		{requests: 30_000, tokens: 180_000_000}: "T5",
	},
	"gpt-5-nano": {
		{requests: 500, tokens: 200_000}:        "T1",
		{requests: 5_000, tokens: 2_000_000}:    "T2",
		{requests: 5_000, tokens: 4_000_000}:    "T3",
		{requests: 10_000, tokens: 10_000_000}:  "T4",
		{requests: 30_000, tokens: 180_000_000}: "T5",
	},
}

func buildOpenAIValidationResult(endpoint *url.URL, model string, headers http.Header) KeyValidationResult {
	result := KeyValidationResult{
		IsValid: true,
	}

	if !isOfficialOpenAIAPI(endpoint) {
		return result
	}

	result.OpenAITierUpdated = true
	result.OpenAITier = inferOpenAITierFromHeaders(model, headers)
	return result
}

func isOfficialOpenAIAPI(endpoint *url.URL) bool {
	if endpoint == nil {
		return false
	}
	return strings.EqualFold(endpoint.Hostname(), "api.openai.com")
}

func inferOpenAITierFromHeaders(model string, headers http.Header) string {
	requests, ok := parseOpenAIRateLimitHeader(headers.Get("x-ratelimit-limit-requests"))
	if !ok {
		return ""
	}
	tokens, ok := parseOpenAIRateLimitHeader(headers.Get("x-ratelimit-limit-tokens"))
	if !ok {
		return ""
	}

	pair := openAIRateLimitPair{requests: requests, tokens: tokens}
	if tiers, ok := openAIRateLimitTiersByFamily[openAIModelRateLimitFamily(model)]; ok {
		return tiers[pair]
	}

	return inferOpenAITierFromKnownLimitPair(pair)
}

func openAIModelRateLimitFamily(model string) string {
	model = strings.ToLower(strings.TrimSpace(model))

	switch {
	case strings.HasPrefix(model, "gpt-4.1-mini"),
		strings.HasPrefix(model, "gpt-4.1-nano"),
		strings.HasPrefix(model, "gpt-4o-mini"):
		return "gpt-4.1-small"
	case strings.HasPrefix(model, "gpt-4.1"):
		return "gpt-4.1"
	case strings.HasPrefix(model, "gpt-5.4-mini"),
		strings.HasPrefix(model, "gpt-5-mini"):
		return "gpt-5-mini"
	case strings.HasPrefix(model, "gpt-5.4-nano"),
		strings.HasPrefix(model, "gpt-5-nano"):
		return "gpt-5-nano"
	case strings.HasPrefix(model, "gpt-5.5"),
		strings.HasPrefix(model, "gpt-5.4"),
		strings.HasPrefix(model, "gpt-5"):
		return "gpt-5"
	default:
		return ""
	}
}

func inferOpenAITierFromKnownLimitPair(pair openAIRateLimitPair) string {
	var inferredTier string
	for _, tiers := range openAIRateLimitTiersByFamily {
		tier, ok := tiers[pair]
		if !ok {
			continue
		}
		if inferredTier != "" && inferredTier != tier {
			return ""
		}
		inferredTier = tier
	}
	return inferredTier
}

func parseOpenAIRateLimitHeader(value string) (int64, bool) {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, ",", "")
	if value == "" {
		return 0, false
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, false
	}
	return parsed, true
}
