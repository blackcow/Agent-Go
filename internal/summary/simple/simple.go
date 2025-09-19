package simple

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"agentgo/internal/model"
)

// Summarizer 使用朴素统计算法生成摘要与关键词。
type Summarizer struct{}

// New 创建简单摘要器。
func New() *Summarizer {
	return &Summarizer{}
}

// Summarize 生成概要信息。
func (s *Summarizer) Summarize(_ context.Context, query string, results []model.Result) (model.Summary, error) {
	if len(results) == 0 {
		return model.Summary{Query: query, GeneratedAt: time.Now()}, nil
	}

	keywords := s.extractKeywords(results)
	highlights := s.buildHighlights(results)
	sentiment := s.estimateSentiment(results)
	sourceCount := map[string]int{}
	for _, r := range results {
		sourceCount[r.Source]++
	}

	overview := s.buildOverview(query, results, highlights)
	return model.Summary{
		Query:           query,
		Overview:        overview,
		Highlights:      highlights,
		Keywords:        keywords,
		Sentiment:       sentiment,
		SourceBreakdown: sourceCount,
		GeneratedAt:     time.Now(),
	}, nil
}

func (s *Summarizer) extractKeywords(results []model.Result) []string {
	freq := map[string]int{}
	for _, r := range results {
		tokens := tokenize(r.Title + " " + r.Summary)
		seen := map[string]struct{}{}
		for _, token := range tokens {
			if len(token) < 2 {
				continue
			}
			if _, ok := seen[token]; ok {
				continue
			}
			freq[token]++
			seen[token] = struct{}{}
		}
	}

	type kv struct {
		key string
		val int
	}
	pairs := make([]kv, 0, len(freq))
	for k, v := range freq {
		pairs = append(pairs, kv{key: k, val: v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].val == pairs[j].val {
			return pairs[i].key < pairs[j].key
		}
		return pairs[i].val > pairs[j].val
	})

	limit := 6
	if len(pairs) < limit {
		limit = len(pairs)
	}
	result := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, pairs[i].key)
	}
	return result
}

func (s *Summarizer) buildHighlights(results []model.Result) []string {
	limit := 3
	if len(results) < limit {
		limit = len(results)
	}
	highlights := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		r := results[i]
		highlight := r.Title
		if r.Summary != "" {
			highlight += " - " + truncate(r.Summary, 80)
		}
		highlights = append(highlights, highlight)
	}
	return highlights
}

func (s *Summarizer) buildOverview(query string, results []model.Result, highlights []string) string {
	if len(results) == 0 {
		return ""
	}
	builder := strings.Builder{}
	builder.WriteString("围绕“")
	builder.WriteString(query)
	builder.WriteString("”共找到")
	builder.WriteString(fmtInt(len(results)))
	builder.WriteString("条相关讨论，主要集中在")
	if len(highlights) > 0 {
		builder.WriteString(highlights[0])
	} else {
		builder.WriteString(results[0].Title)
	}
	builder.WriteString("等话题。")
	return builder.String()
}

func (s *Summarizer) estimateSentiment(results []model.Result) string {
	if len(results) == 0 {
		return "neutral"
	}
	var positive, negative int
	for _, r := range results {
		text := strings.ToLower(r.Summary + " " + r.Title)
		positive += countContains(text, []string{"good", "great", "上升", "突破", "增长", "机遇"})
		negative += countContains(text, []string{"bad", "risk", "下降", "危机", "挑战", "下滑"})
	}
	switch {
	case positive > negative:
		return "positive"
	case negative > positive:
		return "negative"
	default:
		return "neutral"
	}
}

func tokenize(text string) []string {
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsNumber(r))
	})
	tokens := make([]string, 0, len(fields))
	for _, f := range fields {
		token := strings.TrimSpace(strings.ToLower(f))
		if token != "" {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

func truncate(text string, length int) string {
	if len([]rune(text)) <= length {
		return text
	}
	runes := []rune(text)
	return string(runes[:length]) + "…"
}

func countContains(text string, dictionary []string) int {
	count := 0
	for _, word := range dictionary {
		if strings.Contains(text, strings.ToLower(word)) {
			count++
		}
	}
	return count
}

func fmtInt(n int) string {
	if n < 0 {
		return "0"
	}
	return strconv.Itoa(n)
}
