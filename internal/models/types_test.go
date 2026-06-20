package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAPIKeyJSONIncludesOpenAITierWhenEmpty(t *testing.T) {
	data, err := json.Marshal(APIKey{})
	if err != nil {
		t.Fatalf("failed to marshal APIKey: %v", err)
	}

	if !strings.Contains(string(data), `"openai_tier":""`) {
		t.Fatalf("expected openai_tier to be present, got %s", data)
	}
}
