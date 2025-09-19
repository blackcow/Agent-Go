package mock

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"agentgo/internal/model"
	"agentgo/internal/provider"
)

// Provider 使用预置数据模拟知乎、公众号、小红书等平台。
type Provider struct {
	name      string
	data      []model.Result
	mu        sync.Mutex
	callCount int
}

// New 创建一个带有默认数据的 MockProvider。
func New() *Provider {
	now := time.Now()
	data := []model.Result{
		{
			Title:       "知乎用户热议 AI 在内容运营中的应用",
			URL:         "https://www.zhihu.com/question/ai-content-ops",
			Summary:     "讨论了自动化生成工具如何帮助运营人员提升效率，并分享了最新案例。",
			Author:      "知乎用户",
			Source:      "zhihu",
			PublishedAt: now.Add(-2 * time.Hour),
			Metrics: map[string]int64{
				"likes":    1200,
				"comments": 132,
			},
			Tags: []string{"AI", "内容运营"},
		},
		{
			Title:       "微信公众平台：品牌在私域中的增长策略",
			URL:         "https://mp.weixin.qq.com/s/brand-growth-2025",
			Summary:     "文章提出了3 个构建私域流量池的关键步骤，并分析了案例数据。",
			Author:      "新媒体观察",
			Source:      "wechat",
			PublishedAt: now.Add(-3 * time.Hour),
			Metrics: map[string]int64{
				"reads": 5800,
				"likes": 280,
			},
			Tags: []string{"品牌", "私域"},
		},
		{
			Title:       "小红书运营实战：如何抓住年轻用户的注意力",
			URL:         "https://www.xiaohongshu.com/explore/ops-tips",
			Summary:     "从社区氛围、内容调性和互动机制三个角度拆解小红书增长策略。",
			Author:      "运营小红",
			Source:      "xiaohongshu",
			PublishedAt: now.Add(-90 * time.Minute),
			Metrics: map[string]int64{
				"likes": 860,
				"saves": 210,
			},
			Tags: []string{"增长", "社区运营"},
		},
		{
			Title:       "知乎圆桌：品牌该如何进行声誉监测",
			URL:         "https://www.zhihu.com/roundtable/brand-monitor",
			Summary:     "参与嘉宾分享了从数据抓取到情绪分析的全流程方法。",
			Author:      "数据绽放",
			Source:      "zhihu",
			PublishedAt: now.Add(-6 * time.Hour),
			Metrics: map[string]int64{
				"likes":    420,
				"comments": 65,
			},
			Tags: []string{"品牌", "舆情"},
		},
		{
			Title:       "微信·有数：2025 年内容消费趋势洞察",
			URL:         "https://mp.weixin.qq.com/s/data-trend-2025",
			Summary:     "分析了不同城市用户在短视频和图文内容上的消费变化。",
			Author:      "数据有数",
			Source:      "wechat",
			PublishedAt: now.Add(-5 * time.Hour),
			Metrics: map[string]int64{
				"reads": 8700,
				"likes": 490,
			},
			Tags: []string{"数据", "趋势"},
		},
	}

	return &Provider{name: "mock", data: data}
}

// Name 返回 Provider 名称。
func (p *Provider) Name() string {
	return p.name
}

// Search 在预置数据中执行简单的文本匹配。
func (p *Provider) Search(_ context.Context, query string, opts provider.SearchOptions) ([]model.Result, error) {
	p.mu.Lock()
	p.callCount++
	p.mu.Unlock()

	query = strings.ToLower(strings.TrimSpace(query))
	matched := make([]model.Result, 0)
	for _, item := range p.data {
		if query == "" || strings.Contains(strings.ToLower(item.Title), query) || strings.Contains(strings.ToLower(item.Summary), query) {
			matched = append(matched, item)
		}
	}

	sort.Slice(matched, func(i, j int) bool {
		return matched[i].PublishedAt.After(matched[j].PublishedAt)
	})

	limit := opts.Limit
	if limit <= 0 || limit > len(matched) {
		limit = len(matched)
	}
	return matched[:limit], nil
}

// CallCount 返回被调用次数，主要用于测试。
func (p *Provider) CallCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.callCount
}
