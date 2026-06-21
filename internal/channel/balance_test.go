package channel

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFormatCurrencyAmount(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		amount   string
		want     string
	}{
		{name: "cny", currency: "CNY", amount: "12.3400", want: "¥12.34"},
		{name: "usd", currency: "USD", amount: "0.5000", want: "$0.5"},
		{name: "zero", currency: "USD", amount: "0", want: "$0"},
		{name: "unknown currency", currency: "EUR", amount: "3.20", want: "EUR 3.2"},
		{name: "invalid", currency: "USD", amount: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatCurrencyAmount(tt.currency, tt.amount); got != tt.want {
				t.Fatalf("unexpected formatted amount: got %q want %q", got, tt.want)
			}
		})
	}
}

func TestDeepSeekAndOpenRouterChannelsRegistered(t *testing.T) {
	channels := GetChannels()
	seen := make(map[string]bool, len(channels))
	for _, channelType := range channels {
		seen[channelType] = true
	}

	for _, channelType := range []string{"deepseek", "openrouter"} {
		if !seen[channelType] {
			t.Fatalf("expected %s channel to be registered, got %#v", channelType, channels)
		}
	}
}

func TestFetchDeepSeekBalanceFormatsMultipleCurrencies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			t.Fatalf("unexpected authorization header: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"balance_infos": [
				{"currency": "CNY", "total_balance": "12.3400"},
				{"currency": "USD", "total_balance": "5.0000"}
			]
		}`))
	}))
	defer server.Close()

	balance, err := fetchDeepSeekBalanceURL(context.Background(), server.Client(), "sk-test", server.URL)
	if err != nil {
		t.Fatalf("fetchDeepSeekBalanceURL returned error: %v", err)
	}
	if balance != "¥12.34 / $5" {
		t.Fatalf("unexpected balance: got %q want %q", balance, "¥12.34 / $5")
	}
}

func TestFetchOpenRouterBalanceFormatsLimitRemaining(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			t.Fatalf("unexpected authorization header: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"limit_remaining": 7.25}}`))
	}))
	defer server.Close()

	balance, err := fetchOpenRouterKeyBalanceURL(context.Background(), server.Client(), "sk-test", server.URL)
	if err != nil {
		t.Fatalf("fetchOpenRouterKeyBalanceURL returned error: %v", err)
	}
	if balance != "$7.25" {
		t.Fatalf("unexpected balance: got %q want %q", balance, "$7.25")
	}
}

func TestFetchOpenRouterCreditsBalanceFormatsRemainingCredits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			t.Fatalf("unexpected authorization header: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"total_credits": 20.5, "total_usage": 3.25}}`))
	}))
	defer server.Close()

	balance, err := fetchOpenRouterCreditsBalanceURL(context.Background(), server.Client(), "sk-test", server.URL)
	if err != nil {
		t.Fatalf("fetchOpenRouterCreditsBalanceURL returned error: %v", err)
	}
	if balance != "$17.25" {
		t.Fatalf("unexpected balance: got %q want %q", balance, "$17.25")
	}
}

func TestFetchOpenRouterBalancePrefersCreditsBalance(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/credits":
			_, _ = w.Write([]byte(`{"data":{"total_credits": 20, "total_usage": 3}}`))
		case "/api/v1/key":
			_, _ = w.Write([]byte(`{"data":{"limit_remaining": 2}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	balance, err := fetchOpenRouterBalanceURLs(
		context.Background(),
		server.Client(),
		"sk-test",
		server.URL+"/api/v1/credits",
		server.URL+"/api/v1/key",
	)
	if err != nil {
		t.Fatalf("fetchOpenRouterBalanceURLs returned error: %v", err)
	}
	if balance != "$17" {
		t.Fatalf("unexpected balance: got %q want %q", balance, "$17")
	}
	if len(requests) != 1 || requests[0] != "/api/v1/credits" {
		t.Fatalf("unexpected requests: %#v", requests)
	}
}
