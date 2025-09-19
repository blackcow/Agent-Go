package aggregator

import (
	"context"
	"testing"
	"time"

	"agentgo/internal/provider"
	"agentgo/internal/provider/mock"
	simplesummary "agentgo/internal/summary/simple"
)

func TestAggregatorSearchAndCache(t *testing.T) {
	mockProvider := mock.New()
	providers := map[string]provider.Provider{
		mockProvider.Name(): mockProvider,
	}

	agg := New(providers, simplesummary.New(), Config{CacheTTL: time.Minute})

	ctx := context.Background()
	resp, err := agg.Search(ctx, "运营", Options{Limit: 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Results) == 0 {
		t.Fatalf("expected results, got 0")
	}
	if resp.Metadata.Cached {
		t.Fatalf("first response should not be cached")
	}
	if mockProvider.CallCount() != 1 {
		t.Fatalf("expected provider call count 1, got %d", mockProvider.CallCount())
	}

	history := agg.History(5)
	if len(history) != 1 {
		t.Fatalf("expected history size 1, got %d", len(history))
	}
	if history[0].Query != "运营" {
		t.Fatalf("unexpected history query: %s", history[0].Query)
	}

	resp2, err := agg.Search(ctx, "运营", Options{Limit: 3})
	if err != nil {
		t.Fatalf("unexpected error on cache fetch: %v", err)
	}
	if !resp2.Metadata.Cached {
		t.Fatalf("expected cached response")
	}
	if mockProvider.CallCount() != 1 {
		t.Fatalf("expected provider not called again, got %d", mockProvider.CallCount())
	}
}
