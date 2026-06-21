package channel

import (
	"context"
	"encoding/json"
	"fmt"
	app_errors "gpt-load/internal/errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func fetchDeepSeekBalance(ctx context.Context, client *http.Client, apiKey string) (string, error) {
	return fetchDeepSeekBalanceURL(ctx, client, apiKey, "https://api.deepseek.com/user/balance")
}

func fetchDeepSeekBalanceURL(ctx context.Context, client *http.Client, apiKey, endpoint string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create balance request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send balance request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read balance response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("[status %d] %s", resp.StatusCode, app_errors.ParseUpstreamError(body))
	}

	var payload struct {
		BalanceInfos []struct {
			Currency     string `json:"currency"`
			TotalBalance string `json:"total_balance"`
		} `json:"balance_infos"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("failed to parse balance response: %w", err)
	}

	parts := make([]string, 0, len(payload.BalanceInfos))
	for _, info := range payload.BalanceInfos {
		if amount := formatCurrencyAmount(info.Currency, info.TotalBalance); amount != "" {
			parts = append(parts, amount)
		}
	}
	if len(parts) == 0 {
		return "", nil
	}
	return strings.Join(parts, " / "), nil
}

func fetchOpenRouterBalance(ctx context.Context, client *http.Client, apiKey string) (string, error) {
	return fetchOpenRouterBalanceURLs(ctx, client, apiKey, "https://openrouter.ai/api/v1/credits", "https://openrouter.ai/api/v1/key")
}

func fetchOpenRouterBalanceURLs(ctx context.Context, client *http.Client, apiKey, creditsEndpoint, keyEndpoint string) (string, error) {
	balance, err := fetchOpenRouterCreditsBalanceURL(ctx, client, apiKey, creditsEndpoint)
	if err == nil && balance != "" {
		return balance, nil
	}
	keyBalance, keyErr := fetchOpenRouterKeyBalanceURL(ctx, client, apiKey, keyEndpoint)
	if keyErr != nil {
		if err != nil {
			return "", err
		}
		return "", keyErr
	}
	return keyBalance, nil
}

func fetchOpenRouterKeyBalanceURL(ctx context.Context, client *http.Client, apiKey, endpoint string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create balance request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send balance request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read balance response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("[status %d] %s", resp.StatusCode, app_errors.ParseUpstreamError(body))
	}

	var payload struct {
		Data struct {
			LimitRemaining any `json:"limit_remaining"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("failed to parse balance response: %w", err)
	}

	return formatCurrencyAmount("USD", fmt.Sprint(payload.Data.LimitRemaining)), nil
}

func fetchOpenRouterCreditsBalanceURL(ctx context.Context, client *http.Client, apiKey, endpoint string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create credits request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send credits request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read credits response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("[status %d] %s", resp.StatusCode, app_errors.ParseUpstreamError(body))
	}

	var payload struct {
		Data struct {
			TotalCredits any `json:"total_credits"`
			TotalUsage   any `json:"total_usage"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("failed to parse credits response: %w", err)
	}

	totalCredits, ok := parseFlexibleFloat(fmt.Sprint(payload.Data.TotalCredits))
	if !ok {
		return "", nil
	}
	totalUsage, ok := parseFlexibleFloat(fmt.Sprint(payload.Data.TotalUsage))
	if !ok {
		totalUsage = 0
	}

	return formatCurrencyAmount("USD", strconv.FormatFloat(totalCredits-totalUsage, 'f', -1, 64)), nil
}

func formatCurrencyAmount(currency, rawAmount string) string {
	amount, ok := parseFlexibleFloat(rawAmount)
	if !ok {
		return ""
	}

	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "CNY":
		return "¥" + formatMoney(amount)
	case "USD":
		return "$" + formatMoney(amount)
	default:
		return strings.ToUpper(strings.TrimSpace(currency)) + " " + formatMoney(amount)
	}
}

func parseFlexibleFloat(value string) (float64, bool) {
	value = strings.TrimSpace(value)
	if value == "" || value == "<nil>" {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, false
	}
	return parsed, true
}

func formatMoney(value float64) string {
	formatted := strconv.FormatFloat(value, 'f', 4, 64)
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimRight(formatted, ".")
	if formatted == "" || formatted == "-0" {
		return "0"
	}
	return formatted
}

func balanceHost(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

func mergeBalanceResult(base KeyValidationResult, provider, balance, host string) KeyValidationResult {
	if balance == "" {
		base.TierProvider = provider
		base.TierReason = "missing_balance"
		base.TierHost = host
		return base
	}
	base.Tier = balance
	base.TierUpdated = true
	base.TierProvider = provider
	base.TierReason = "balance"
	base.TierHost = host
	return base
}
