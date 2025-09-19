package model

import "time"

// Result 表示从任意平台抓取的一条内容。
type Result struct {
	Title       string            `json:"title"`
	URL         string            `json:"url"`
	Summary     string            `json:"summary"`
	Author      string            `json:"author"`
	Source      string            `json:"source"`
	PublishedAt time.Time         `json:"published_at"`
	Metrics     map[string]int64  `json:"metrics,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Extras      map[string]string `json:"extras,omitempty"`
}

// Summary 聚合后的综述结果。
type Summary struct {
	Query           string         `json:"query"`
	Overview        string         `json:"overview"`
	Highlights      []string       `json:"highlights"`
	Keywords        []string       `json:"keywords"`
	Sentiment       string         `json:"sentiment"`
	SourceBreakdown map[string]int `json:"source_breakdown"`
	GeneratedAt     time.Time      `json:"generated_at"`
}
