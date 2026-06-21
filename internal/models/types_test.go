package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAPIKeyJSONIncludesTierWhenEmpty(t *testing.T) {
	data, err := json.Marshal(APIKey{})
	if err != nil {
		t.Fatalf("failed to marshal APIKey: %v", err)
	}

	if !strings.Contains(string(data), `"tier":""`) {
		t.Fatalf("expected tier to be present, got %s", data)
	}
	if strings.Contains(string(data), `"openai_tier"`) {
		t.Fatalf("openai_tier should not be exposed in APIKey JSON, got %s", data)
	}
}
