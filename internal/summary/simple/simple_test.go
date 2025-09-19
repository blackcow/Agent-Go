package simple

import (
	"context"
	"testing"
	"time"

	"agentgo/internal/model"
)

func TestSummarizer(t *testing.T) {
	s := New()
	results := []model.Result{
		{
			Title:       "示例标题一",
			Summary:     "这是一个关于增长的积极案例。",
			Source:      "zhihu",
			PublishedAt: time.Now(),
		},
		{
			Title:       "示例标题二",
			Summary:     "讨论中提到了挑战和风险。",
			Source:      "wechat",
			PublishedAt: time.Now().Add(-time.Hour),
		},
	}

	summary, err := s.Summarize(context.Background(), "增长", results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Query != "增长" {
		t.Fatalf("unexpected query in summary")
	}
	if len(summary.Highlights) == 0 {
		t.Fatalf("expected highlights")
	}
	if len(summary.SourceBreakdown) != 2 {
		t.Fatalf("expected source breakdown for 2 sources")
	}
}
