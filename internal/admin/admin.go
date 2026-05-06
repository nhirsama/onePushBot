package admin

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed web/*
var webFS embed.FS

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

func NewHandler(token string, controller Controller) http.Handler {
	mux := http.NewServeMux()
	rootFS, err := fs.Sub(webFS, "web")
	if err != nil {
		panic(err)
	}
	staticFS := http.FileServer(http.FS(rootFS))
	mux.Handle("/", staticFS)
	handler := &handler{
		token:      token,
		controller: controller,
	}
	mux.Handle("/api/status", handler.auth(http.HandlerFunc(handler.handleStatus)))
	mux.Handle("/api/logs", handler.auth(http.HandlerFunc(handler.handleLogs)))
	mux.Handle("/api/config", handler.auth(http.HandlerFunc(handler.handleConfig)))
	mux.Handle("/api/reload", handler.auth(http.HandlerFunc(handler.handleReload)))
	mux.Handle("/api/restart", handler.auth(http.HandlerFunc(handler.handleRestart)))
	return mux
}

type handler struct {
	token      string
	controller Controller
}

func (h *handler) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.token == "" {
			next.ServeHTTP(w, r)
			return
		}
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			token = r.URL.Query().Get("token")
		}
		if token != h.token {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	status, err := h.controller.Status(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, status)
}

func (h *handler) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	since, limit := parseUintQuery(r, "since"), int(parseUintQuery(r, "limit"))
	if limit <= 0 {
		limit = 500
	}
	items, err := h.controller.Logs(r.Context(), since, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]any{"logs": items})
}

func (h *handler) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg, err := h.controller.Config(r.Context())
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
		if err := h.controller.UpdateConfig(r.Context(), patch); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, map[string]any{"ok": true})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *handler) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := h.controller.Reload(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func (h *handler) handleRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := h.controller.Restart(r.Context()); err != nil {
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
