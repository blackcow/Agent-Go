package history

import (
	"sort"
	"sync"
	"time"
)

// Record 表示一次搜索请求的元数据。
type Record struct {
	Query     string        `json:"query"`
	Providers []string      `json:"providers"`
	Results   int           `json:"results"`
	Took      time.Duration `json:"took"`
	Time      time.Time     `json:"time"`
}

// Store 持久化最近的搜索记录。
type Store struct {
	limit   int
	records []Record
	mu      sync.RWMutex
}

// NewStore 构建历史记录仓库。
func NewStore(limit int) *Store {
	if limit <= 0 {
		limit = 50
	}
	return &Store{
		limit: limit,
	}
}

// Add 追加一条记录。
func (s *Store) Add(record Record) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append([]Record{record}, s.records...)
	if len(s.records) > s.limit {
		s.records = s.records[:s.limit]
	}
}

// List 返回最近的若干条记录。
func (s *Store) List(limit int) []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > len(s.records) {
		limit = len(s.records)
	}
	result := make([]Record, limit)
	copy(result, s.records[:limit])
	sort.Slice(result, func(i, j int) bool {
		return result[i].Time.After(result[j].Time)
	})
	return result
}
