package channel

import (
	"context"
	"gpt-load/internal/models"
)

func init() {
	Register("openrouter", newOpenRouterChannel)
}

type OpenRouterChannel struct {
	*OpenAIChannel
}

func newOpenRouterChannel(f *Factory, group *models.Group) (ChannelProxy, error) {
	base, err := f.newBaseChannel("openrouter", group)
	if err != nil {
		return nil, err
	}

	return &OpenRouterChannel{
		OpenAIChannel: &OpenAIChannel{BaseChannel: base},
	}, nil
}

func (ch *OpenRouterChannel) ValidateKey(ctx context.Context, apiKey *models.APIKey, group *models.Group) (KeyValidationResult, error) {
	result, err := ch.OpenAIChannel.ValidateKey(ctx, apiKey, group)
	if err != nil {
		return result, err
	}

	balance, balanceErr := fetchOpenRouterBalance(ctx, ch.HTTPClient, apiKey.KeyValue)
	if balanceErr != nil {
		result.TierProvider = "openrouter"
		result.TierReason = "balance_error"
		result.TierHost = balanceHost("https://openrouter.ai")
		return result, nil
	}

	return mergeBalanceResult(result, "openrouter", balance, balanceHost("https://openrouter.ai")), nil
}
