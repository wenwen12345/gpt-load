package channel

import (
	"context"
	"gpt-load/internal/models"
)

func init() {
	Register("deepseek", newDeepSeekChannel)
}

type DeepSeekChannel struct {
	*OpenAIChannel
}

func newDeepSeekChannel(f *Factory, group *models.Group) (ChannelProxy, error) {
	base, err := f.newBaseChannel("deepseek", group)
	if err != nil {
		return nil, err
	}

	return &DeepSeekChannel{
		OpenAIChannel: &OpenAIChannel{BaseChannel: base},
	}, nil
}

func (ch *DeepSeekChannel) ValidateKey(ctx context.Context, apiKey *models.APIKey, group *models.Group) (KeyValidationResult, error) {
	result, err := ch.OpenAIChannel.ValidateKey(ctx, apiKey, group)
	if err != nil {
		return result, err
	}

	balance, balanceErr := fetchDeepSeekBalance(ctx, ch.HTTPClient, apiKey.KeyValue)
	if balanceErr != nil {
		result.TierProvider = "deepseek"
		result.TierReason = "balance_error"
		result.TierHost = balanceHost("https://api.deepseek.com")
		return result, nil
	}

	return mergeBalanceResult(result, "deepseek", balance, balanceHost("https://api.deepseek.com")), nil
}
