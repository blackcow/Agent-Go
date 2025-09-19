package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"agentgo/internal/aggregator"
)

// Server 封装 HTTP 接口。
type Server struct {
	aggregator *aggregator.Aggregator
	mux        *http.ServeMux
}

// New 创建 HTTP Server。
func New(agg *aggregator.Aggregator) *Server {
	srv := &Server{
		aggregator: agg,
		mux:        http.NewServeMux(),
	}
	srv.registerRoutes()
	return srv
}

// Handler 返回底层 HTTP 处理器。
func (s *Server) Handler() http.Handler {
	return s.mux
}

// Run 启动 HTTP 服务。
func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.Handler())
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/healthz", s.handleHealth)
	s.mux.HandleFunc("/v1/search", s.handleSearch)
	s.mux.HandleFunc("/v1/history", s.handleHistory)
	s.mux.HandleFunc("/v1/providers", s.handleProviders)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "query parameter `q` is required"})
		return
	}

	limit := parseInt(r.URL.Query().Get("limit"), 10)
	providers := parseList(r.URL.Query().Get("providers"))
	forceRefresh := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("fresh")), "true")

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	resp, err := s.aggregator.Search(ctx, query, aggregator.Options{
		Providers:    providers,
		Limit:        limit,
		ForceRefresh: forceRefresh,
	})
	if err != nil {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	s.writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	limit := parseInt(r.URL.Query().Get("limit"), 20)
	records := s.aggregator.History(limit)
	s.writeJSON(w, http.StatusOK, map[string]any{"records": records})
}

func (s *Server) handleProviders(w http.ResponseWriter, r *http.Request) {
	names := s.aggregator.ProviderNames()
	s.writeJSON(w, http.StatusOK, map[string]any{"providers": names})
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(payload)
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return v
}

func parseList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
