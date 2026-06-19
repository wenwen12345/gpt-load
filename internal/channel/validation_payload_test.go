package channel

import (
	"gpt-load/internal/models"
	"net/url"
	"testing"
)

func TestRegisteredChannelsIncludesOpenAIImageGeneration(t *testing.T) {
	channels := GetChannels()
	for _, channelType := range channels {
		if channelType == "openai-image-generation" {
			return
		}
	}

	t.Fatalf("expected openai-image-generation to be registered, got %#v", channels)
}

func TestBuildOpenAIValidationPayloadUsesResponsesFormat(t *testing.T) {
	payload := buildOpenAIValidationPayload("/v1/responses", "gpt-5.5")

	if payload["model"] != "gpt-5.5" {
		t.Fatalf("unexpected model: %v", payload["model"])
	}
	if payload["input"] != "hi" {
		t.Fatalf("expected Responses API input payload, got %v", payload["input"])
	}
	if _, exists := payload["messages"]; exists {
		t.Fatal("Responses API validation payload must not include messages")
	}
}

func TestBuildOpenAIValidationPayloadUsesChatCompletionsFormat(t *testing.T) {
	payload := buildOpenAIValidationPayload("/v1/chat/completions", "gpt-4.1-nano")

	if payload["model"] != "gpt-4.1-nano" {
		t.Fatalf("unexpected model: %v", payload["model"])
	}
	if _, exists := payload["input"]; exists {
		t.Fatal("Chat Completions validation payload must not include input")
	}
	messages, ok := payload["messages"].([]map[string]string)
	if !ok || len(messages) != 1 {
		t.Fatalf("expected one chat message, got %#v", payload["messages"])
	}
	if messages[0]["role"] != "user" || messages[0]["content"] != "hi" {
		t.Fatalf("unexpected message payload: %#v", messages[0])
	}
}

func TestBuildOpenAIImageGenerationValidationPayloadUsesPrompt(t *testing.T) {
	payload := buildOpenAIImageGenerationValidationPayload("gpt-image-2")

	if payload["model"] != "gpt-image-2" {
		t.Fatalf("unexpected model: %v", payload["model"])
	}
	if payload["prompt"] != "cat" {
		t.Fatalf("expected image generation prompt cat, got %v", payload["prompt"])
	}
	if _, exists := payload["messages"]; exists {
		t.Fatal("Image generation validation payload must not include messages")
	}
	if _, exists := payload["input"]; exists {
		t.Fatal("Image generation validation payload must not include input")
	}
}

func TestBuildAnthropicValidationPayloadUsesMessagesFormat(t *testing.T) {
	payload := buildAnthropicValidationPayload("claude-sonnet-4-20250514")

	if payload["model"] != "claude-sonnet-4-20250514" {
		t.Fatalf("unexpected model: %v", payload["model"])
	}
	if payload["max_tokens"] != 100 {
		t.Fatalf("expected max_tokens, got %v", payload["max_tokens"])
	}
	if _, exists := payload["input"]; exists {
		t.Fatal("Anthropic validation payload must not include Responses API input")
	}
	messages, ok := payload["messages"].([]map[string]string)
	if !ok || len(messages) != 1 {
		t.Fatalf("expected one Anthropic message, got %#v", payload["messages"])
	}
	if messages[0]["role"] != "user" || messages[0]["content"] != "hi" {
		t.Fatalf("unexpected Anthropic message payload: %#v", messages[0])
	}
}

func TestGeminiBuildValidationRequestNativeFormat(t *testing.T) {
	ch := &GeminiChannel{
		BaseChannel: &BaseChannel{
			Name:               "gemini",
			Upstreams:          []UpstreamInfo{{URL: mustParseURL(t, "https://generativelanguage.googleapis.com")}},
			TestModel:          "gemini-2.0-flash-lite",
			ValidationEndpoint: "",
		},
	}

	reqURL, payload, authMode, err := ch.buildValidationRequest(&models.APIKey{KeyValue: "sk-test"})
	if err != nil {
		t.Fatalf("buildValidationRequest returned error: %v", err)
	}
	if authMode != "query" {
		t.Fatalf("expected query auth mode, got %q", authMode)
	}
	if reqURL != "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite:generateContent?key=sk-test" {
		t.Fatalf("unexpected request URL: %s", reqURL)
	}
	if _, exists := payload["contents"]; !exists {
		t.Fatalf("expected native Gemini contents payload, got %#v", payload)
	}
}

func TestGeminiBuildValidationRequestOpenAICompatibleFormat(t *testing.T) {
	ch := &GeminiChannel{
		BaseChannel: &BaseChannel{
			Name:               "gemini",
			Upstreams:          []UpstreamInfo{{URL: mustParseURL(t, "https://generativelanguage.googleapis.com")}},
			TestModel:          "gemini-2.0-flash-lite",
			ValidationEndpoint: "/v1beta/openai/chat/completions",
		},
	}

	reqURL, payload, authMode, err := ch.buildValidationRequest(&models.APIKey{KeyValue: "sk-test"})
	if err != nil {
		t.Fatalf("buildValidationRequest returned error: %v", err)
	}
	if authMode != "bearer" {
		t.Fatalf("expected bearer auth mode, got %q", authMode)
	}
	if reqURL != "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions" {
		t.Fatalf("unexpected request URL: %s", reqURL)
	}
	if _, exists := payload["messages"]; !exists {
		t.Fatalf("expected OpenAI-compatible messages payload, got %#v", payload)
	}
	if _, exists := payload["contents"]; exists {
		t.Fatal("OpenAI-compatible validation payload must not include Gemini contents")
	}
}

func mustParseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("failed to parse test URL: %v", err)
	}
	return parsed
}
