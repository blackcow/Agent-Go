package aggregator

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"agentgo/internal/cache"
	"agentgo/internal/history"
	"agentgo/internal/model"
	"agentgo/internal/provider"
	"agentgo/internal/summary"
)

// Options 控制聚合行为。
type Options struct {
	Providers    []string
	Limit        int
	ForceRefresh bool
}

// Metadata 描述一次聚合的额外信息。
type Metadata struct {
	Cached           bool             `json:"cached"`
	GeneratedAt      time.Time        `json:"generated_at"`
	Took             time.Duration    `json:"took"`
	ProviderStatuses []ProviderStatus `json:"provider_statuses"`
}

// ProviderStatus 记录单个平台的执行情况。
type ProviderStatus struct {
	Name   string `json:"name"`
	Count  int    `json:"count"`
	Error  string `json:"error,omitempty"`
	Cached bool   `json:"cached"`
}

// Response 为聚合搜索的完整返回。
type Response struct {
	Query    string         `json:"query"`
	Results  []model.Result `json:"results"`
	Summary  model.Summary  `json:"summary"`
	Metadata Metadata       `json:"metadata"`
}

// Config 聚合器基础配置。
type Config struct {
	CacheTTL       time.Duration
	RequestTimeout time.Duration
	HistorySize    int
}

// Aggregator 负责并发调度多个 provider 并汇总结果。
type Aggregator struct {
	providers  map[string]provider.Provider
	cache      *cache.Cache[string, Response]
	summarizer summary.Summarizer
	history    *history.Store
	timeout    time.Duration
	mu         sync.RWMutex
}

// New 创建聚合器。
func New(providers map[string]provider.Provider, summarizer summary.Summarizer, cfg Config) *Aggregator {
	ttl := cfg.CacheTTL
	if ttl <= 0 {
		ttl = 3 * time.Minute
	}
	timeout := cfg.RequestTimeout
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	historySize := cfg.HistorySize
	if historySize <= 0 {
		historySize = 50
	}

	return &Aggregator{
		providers:  providers,
		cache:      cache.New[string, Response](ttl),
		summarizer: summarizer,
		history:    history.NewStore(historySize),
		timeout:    timeout,
	}
}

// ProviderNames 返回聚合器中注册的 provider 名称。
func (a *Aggregator) ProviderNames() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	names := make([]string, 0, len(a.providers))
	for name := range a.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// RegisterProvider 允许动态增加 provider。
func (a *Aggregator) RegisterProvider(p provider.Provider) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.providers == nil {
		a.providers = map[string]provider.Provider{}
	}
	a.providers[p.Name()] = p
}

// Search 执行聚合查询。
func (a *Aggregator) Search(ctx context.Context, query string, opts Options) (Response, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return Response{}, errors.New("query is required")
	}

	start := time.Now()

	providers := a.selectProviders(opts.Providers)
	if len(providers) == 0 {
		return Response{}, errors.New("no providers configured")
	}

	cacheKey := a.buildCacheKey(query, providers, opts.Limit)
	if !opts.ForceRefresh {
		if resp, ok := a.cache.Get(cacheKey); ok {
			resp.Metadata.Cached = true
			resp.Metadata.Took = time.Since(start)
			return resp, nil
		}
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 10
	}

	type resultEnvelope struct {
		provider string
		results  []model.Result
		err      error
	}

	resultCh := make(chan resultEnvelope, len(providers))
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	for _, name := range providers {
		prov := a.getProvider(name)
		if prov == nil {
			resultCh <- resultEnvelope{provider: name, err: fmt.Errorf("provider %s not found", name)}
			continue
		}
		wg.Add(1)
		go func(p provider.Provider) {
			defer wg.Done()
			res, err := p.Search(ctx, query, provider.SearchOptions{Limit: limit})
			resultCh <- resultEnvelope{provider: p.Name(), results: res, err: err}
		}(prov)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	aggregated := make([]model.Result, 0)
	statuses := make([]ProviderStatus, 0, len(providers))
	providerNames := make([]string, 0, len(providers))

	for envelope := range resultCh {
		providerNames = append(providerNames, envelope.provider)
		if envelope.err != nil {
			statuses = append(statuses, ProviderStatus{Name: envelope.provider, Error: envelope.err.Error()})
			continue
		}
		aggregated = append(aggregated, envelope.results...)
		statuses = append(statuses, ProviderStatus{Name: envelope.provider, Count: len(envelope.results)})
	}

	sort.Slice(aggregated, func(i, j int) bool {
		return aggregated[i].PublishedAt.After(aggregated[j].PublishedAt)
	})

	summaryResult, err := a.summarizer.Summarize(ctx, query, aggregated)
	if err != nil {
		statuses = append(statuses, ProviderStatus{Name: "summary", Error: err.Error()})
	}

	resp := Response{
		Query:   query,
		Results: aggregated,
		Summary: summaryResult,
		Metadata: Metadata{
			GeneratedAt:      time.Now(),
			Took:             time.Since(start),
			ProviderStatuses: statuses,
		},
	}

	a.cache.Set(cacheKey, resp)
	a.history.Add(history.Record{
		Query:     query,
		Providers: providerNames,
		Results:   len(aggregated),
		Took:      resp.Metadata.Took,
		Time:      resp.Metadata.GeneratedAt,
	})

	return resp, nil
}

// History 返回最近的查询记录。
func (a *Aggregator) History(limit int) []history.Record {
	return a.history.List(limit)
}

func (a *Aggregator) buildCacheKey(query string, providers []string, limit int) string {
	cloned := append([]string(nil), providers...)
	sort.Strings(cloned)
	return fmt.Sprintf("%s|%s|%d", strings.ToLower(query), strings.Join(cloned, ","), limit)
}

func (a *Aggregator) selectProviders(requested []string) []string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if len(requested) == 0 {
		names := make([]string, 0, len(a.providers))
		for name := range a.providers {
			names = append(names, name)
		}
		sort.Strings(names)
		return names
	}

	names := make([]string, 0, len(requested))
	for _, name := range requested {
		name = strings.TrimSpace(name)
		if _, ok := a.providers[name]; ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func (a *Aggregator) getProvider(name string) provider.Provider {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.providers[name]
}
