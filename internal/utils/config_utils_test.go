package utils

import (
	"gpt-load/internal/models"
	"testing"
)

func TestGetValidationEndpointUsesOpenAIImageGenerationDefault(t *testing.T) {
	group := &models.Group{ChannelType: "openai-image-generation"}

	if endpoint := GetValidationEndpoint(group); endpoint != "/v1/images/generations" {
		t.Fatalf("unexpected validation endpoint: %s", endpoint)
	}
}
