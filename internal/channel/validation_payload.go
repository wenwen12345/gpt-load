package channel

import "strings"

func buildOpenAIValidationPayload(endpointPath, model string) map[string]any {
	if isOpenAIResponsesEndpoint(endpointPath) {
		return map[string]any{
			"model": model,
			"input": "hi",
		}
	}

	return map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": "hi"},
		},
	}
}

func buildGeminiNativeValidationPayload() map[string]any {
	return map[string]any{
		"contents": []map[string]any{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": "hi"},
				},
			},
		},
	}
}

func buildAnthropicValidationPayload(model string) map[string]any {
	return map[string]any{
		"model":      model,
		"max_tokens": 100,
		"messages": []map[string]string{
			{"role": "user", "content": "hi"},
		},
	}
}

func isOpenAIResponsesEndpoint(endpointPath string) bool {
	normalized := strings.ToLower(strings.TrimRight(endpointPath, "/"))
	return normalized == "/v1/responses" || strings.HasSuffix(normalized, "/responses")
}

func isGeminiOpenAICompatibleEndpoint(endpointPath string) bool {
	return strings.Contains(strings.ToLower(endpointPath), "/openai/")
}
