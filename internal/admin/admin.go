package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Status struct {
	Running   bool              `json:"running"`
	Platforms map[string]string `json:"platforms"`
	Process   ProcessStatus     `json:"process"`
	Config    map[string]any    `json:"config"`
}

type ProcessStatus struct {
	Uptime      string `json:"uptime"`
	MemoryAlloc string `json:"memory_alloc"`
	MemorySys   string `json:"memory_sys"`
	HeapObjects uint64 `json:"heap_objects"`
	Goroutines  int    `json:"goroutines"`
	CPUPercent  string `json:"cpu_percent"`
	CPUTime     string `json:"cpu_time"`
}

type Controller interface {
	Status(ctx context.Context) (Status, error)
	Logs(ctx context.Context, since uint64, limit int) ([]LogEntry, error)
	Config(ctx context.Context) (map[string]any, error)
	UpdateConfig(ctx context.Context, patch map[string]any) error
	Reload(ctx context.Context) error
	Restart(ctx context.Context) error
}

type LogEntry struct {
	ID      uint64 `json:"id"`
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

type Server struct {
	addr       string
	token      string
	controller Controller
	server     *http.Server
}

func New(addr string, token string, controller Controller) *Server {
	if addr == "" {
		addr = "127.0.0.1:8090"
	}
	s := &Server{
		addr:       addr,
		token:      token,
		controller: controller,
	}
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.routes(),
	}
	return s
}

func (s *Server) Addr() string {
	return s.addr
}

func (s *Server) Start() error {
	err := s.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Close(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.Handle("/api/status", s.auth(http.HandlerFunc(s.handleStatus)))
	mux.Handle("/api/logs", s.auth(http.HandlerFunc(s.handleLogs)))
	mux.Handle("/api/config", s.auth(http.HandlerFunc(s.handleConfig)))
	mux.Handle("/api/reload", s.auth(http.HandlerFunc(s.handleReload)))
	mux.Handle("/api/restart", s.auth(http.HandlerFunc(s.handleRestart)))
	return mux
}

func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.token == "" {
			next.ServeHTTP(w, r)
			return
		}
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			token = r.URL.Query().Get("token")
		}
		if token != s.token {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	status, err := s.controller.Status(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, status)
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	since, limit := parseUintQuery(r, "since"), int(parseUintQuery(r, "limit"))
	if limit <= 0 {
		limit = 500
	}
	items, err := s.controller.Logs(r.Context(), since, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]any{"logs": items})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg, err := s.controller.Config(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, cfg)
	case http.MethodPut:
		var patch map[string]any
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := s.controller.UpdateConfig(r.Context(), patch); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, map[string]any{"ok": true})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := s.controller.Reload(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := s.controller.Restart(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func parseUintQuery(r *http.Request, key string) uint64 {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return 0
	}
	var result uint64
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return 0
		}
		result = result*10 + uint64(ch-'0')
	}
	return result
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": message})
}
