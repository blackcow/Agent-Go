package provider

import (
	"context"
	"time"

	"agentgo/internal/model"
)

// SearchOptions 控制 provider 的检索行为。
type SearchOptions struct {
	Limit     int
	StartTime time.Time
	EndTime   time.Time
}

// Provider 统一所有平台的检索能力。
type Provider interface {
	Name() string
	Search(ctx context.Context, query string, opts SearchOptions) ([]model.Result, error)
}

// ErrNotConfigured 当 provider 因缺少配置而不可用时返回。
type ErrNotConfigured struct {
	Provider string
}

func (e ErrNotConfigured) Error() string {
	return e.Provider + " provider is not configured"
}
