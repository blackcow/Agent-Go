package summary

import (
	"context"

	"agentgo/internal/model"
)

// Summarizer 负责根据聚合结果生成概要。
type Summarizer interface {
	Summarize(ctx context.Context, query string, results []model.Result) (model.Summary, error)
}
