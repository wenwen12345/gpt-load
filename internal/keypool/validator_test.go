package keypool

import (
	"gpt-load/internal/models"
	"gpt-load/internal/types"
	"testing"
	"time"
)

func TestKeyValidationTimeoutUsesConfiguredValueForTextChannels(t *testing.T) {
	group := &models.Group{
		ChannelType: "openai",
		EffectiveConfig: types.SystemSettings{
			KeyValidationTimeoutSeconds: 20,
		},
	}

	if timeout := keyValidationTimeout(group); timeout != 20*time.Second {
		t.Fatalf("unexpected timeout: %v", timeout)
	}
}

func TestKeyValidationTimeoutRaisesOpenAIImageGenerationMinimum(t *testing.T) {
	group := &models.Group{
		ChannelType: "openai-image-generation",
		EffectiveConfig: types.SystemSettings{
			KeyValidationTimeoutSeconds: 20,
		},
	}

	if timeout := keyValidationTimeout(group); timeout != 300*time.Second {
		t.Fatalf("unexpected timeout: %v", timeout)
	}
}

func TestKeyValidationTimeoutKeepsHigherOpenAIImageGenerationValue(t *testing.T) {
	group := &models.Group{
		ChannelType: "openai-image-generation",
		EffectiveConfig: types.SystemSettings{
			KeyValidationTimeoutSeconds: 600,
		},
	}

	if timeout := keyValidationTimeout(group); timeout != 600*time.Second {
		t.Fatalf("unexpected timeout: %v", timeout)
	}
}
