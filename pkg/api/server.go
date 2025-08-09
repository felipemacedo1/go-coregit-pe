package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/felipemacedo1/go-coregit-pe/internal/logging"
	"github.com/felipemacedo1/go-coregit-pe/pkg/core"
	"github.com/felipemacedo1/go-coregit-pe/pkg/core/execgit"
)

// Server provides HTTP API for Git operations
type Server struct {
	git    core.CoreGit
	logger *logging.Logger
	server *http.Server
}

// NewServer creates a new API server
func NewServer(addr string) *Server {
	git := execgit.New()
	logger := logging.NewLogger(nil, false)

	s := &Server{
		git:    git,
		logger: logger,
	}

	mux := http.NewServeMux()
	s.setupRoutes(mux)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return s
}

// setupRoutes configures HTTP routes
func (s *Server) setupRoutes(mux *http.ServeMux) {
	// Repository operations
	mux.HandleFunc("/v1/repo", s.handleRepo)
	mux.HandleFunc("/v1/clone", s.handleClone)
	mux.HandleFunc("/v1/status", s.handleStatus)
	mux.HandleFunc("/v1/log", s.handleLog)
	mux.HandleFunc("/v1/diff", s.handleDiff)

	// Sync operations
	mux.HandleFunc("/v1/fetch", s.handleFetch)
	mux.HandleFunc("/v1/pull", s.handlePull)
	mux.HandleFunc("/v1/push", s.handlePush)

	// Raw command execution
	mux.HandleFunc("/v1/raw", s.handleRaw)

	// Health check
	mux.HandleFunc("/health", s.handleHealth)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("Starting API server", map[string]interface{}{
		"addr": s.server.Addr,
	})
	return s.server.ListenAndServe()
}

// Stop stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping API server")
	return s.server.Shutdown(ctx)
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// writeJSON writes a JSON response
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, Response{
		Success: false,
		Error:   message,
	})
}

// writeSuccess writes a success response
func (s *Server) writeSuccess(w http.ResponseWriter, data interface{}) {
	s.writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	s.writeSuccess(w, map[string]string{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// handleRepo handles repository info requests
func (s *Server) handleRepo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		s.writeError(w, http.StatusBadRequest, "path parameter is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	repo, err := s.git.Open(ctx, path)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to open repository: %v", err))
		return
	}

	s.writeSuccess(w, repo)
}

// CloneRequest represents a clone request
type CloneRequest struct {
	URL       string   `json:"url"`
	Path      string   `json:"path"`
	Branch    string   `json:"branch,omitempty"`
	Depth     int      `json:"depth,omitempty"`
	Sparse    []string `json:"sparse,omitempty"`
	Recursive bool     `json:"recursive,omitempty"`
}

// handleClone handles repository clone requests
func (s *Server) handleClone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req CloneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if req.URL == "" || req.Path == "" {
		s.writeError(w, http.StatusBadRequest, "url and path are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	opts := core.CloneOptions{
		URL:       req.URL,
		Path:      req.Path,
		Branch:    req.Branch,
		Depth:     req.Depth,
		Sparse:    req.Sparse,
		Recursive: req.Recursive,
		Progress:  true,
	}

	repo, err := s.git.Clone(ctx, opts)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Clone failed: %v", err))
		return
	}

	s.writeSuccess(w, repo)
}

// handleStatus handles repository status requests
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		s.writeError(w, http.StatusBadRequest, "path parameter is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	repo, err := s.git.Open(ctx, path)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to open repository: %v", err))
		return
	}

	status, err := s.git.GetStatus(ctx, repo)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get status: %v", err))
		return
	}

	s.writeSuccess(w, status)
}

// handleLog handles log requests
func (s *Server) handleLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		s.writeError(w, http.StatusBadRequest, "path parameter is required")
		return
	}

	maxStr := r.URL.Query().Get("max")
	max := 10 // default
	if maxStr != "" {
		if m, err := strconv.Atoi(maxStr); err == nil && m > 0 {
			max = m
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	repo, err := s.git.Open(ctx, path)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to open repository: %v", err))
		return
	}

	commits, err := s.git.Log(ctx, repo, "", max, false)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get log: %v", err))
		return
	}

	s.writeSuccess(w, commits)
}

// handleDiff handles diff requests
func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		s.writeError(w, http.StatusBadRequest, "path parameter is required")
		return
	}

	base := r.URL.Query().Get("base")
	head := r.URL.Query().Get("head")
	stat := r.URL.Query().Get("stat") == "true"

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	repo, err := s.git.Open(ctx, path)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to open repository: %v", err))
		return
	}

	diff, err := s.git.Diff(ctx, repo, base, head, stat)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get diff: %v", err))
		return
	}

	s.writeSuccess(w, map[string]string{"diff": diff})
}

// SyncRequest represents a sync operation request
type SyncRequest struct {
	Path   string `json:"path"`
	Remote string `json:"remote,omitempty"`
	Branch string `json:"branch,omitempty"`
	Force  bool   `json:"force,omitempty"`
	Prune  bool   `json:"prune,omitempty"`
	Tags   bool   `json:"tags,omitempty"`
	Rebase bool   `json:"rebase,omitempty"`
}

// handleFetch handles fetch requests
func (s *Server) handleFetch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if req.Path == "" {
		s.writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	repo, err := s.git.Open(ctx, req.Path)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to open repository: %v", err))
		return
	}

	err = s.git.Fetch(ctx, repo, req.Remote, req.Prune, req.Tags)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Fetch failed: %v", err))
		return
	}

	s.writeSuccess(w, map[string]string{"message": "Fetch completed successfully"})
}

// handlePull handles pull requests
func (s *Server) handlePull(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if req.Path == "" {
		s.writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	repo, err := s.git.Open(ctx, req.Path)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to open repository: %v", err))
		return
	}

	err = s.git.Pull(ctx, repo, req.Remote, req.Branch, req.Rebase)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Pull failed: %v", err))
		return
	}

	s.writeSuccess(w, map[string]string{"message": "Pull completed successfully"})
}

// handlePush handles push requests
func (s *Server) handlePush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if req.Path == "" {
		s.writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	repo, err := s.git.Open(ctx, req.Path)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to open repository: %v", err))
		return
	}

	err = s.git.Push(ctx, repo, req.Remote, req.Branch, req.Force, req.Tags)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Push failed: %v", err))
		return
	}

	s.writeSuccess(w, map[string]string{"message": "Push completed successfully"})
}

// RawRequest represents a raw command request
type RawRequest struct {
	Path string   `json:"path"`
	Args []string `json:"args"`
}

// handleRaw handles raw git command requests
func (s *Server) handleRaw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req RawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON request")
		return
	}

	if req.Path == "" || len(req.Args) == 0 {
		s.writeError(w, http.StatusBadRequest, "path and args are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Minute)
	defer cancel()

	repo, err := s.git.Open(ctx, req.Path)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Failed to open repository: %v", err))
		return
	}

	result, err := s.git.RunRaw(ctx, repo, req.Args)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Command failed: %v", err))
		return
	}

	s.writeSuccess(w, result)
}
