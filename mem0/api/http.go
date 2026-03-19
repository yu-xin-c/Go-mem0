package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"go-mem0/mem0/model"
	"go-mem0/mem0/service"
)

type Server struct {
	service *service.Service
	mux     *http.ServeMux
}

func NewServer(service *service.Service) (*Server, error) {
	if service == nil {
		return nil, errors.New("service is nil")
	}
	s := &Server{
		service: service,
		mux:     http.NewServeMux(),
	}
	s.routes()
	return s, nil
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("POST /v1/memories/update", s.handleUpdateMemories)
	s.mux.HandleFunc("GET /v1/memories/search", s.handleSearchMemories)
	s.mux.HandleFunc("GET /v1/memories", s.handleListMemories)
}

type updateMemoriesRequest struct {
	UserID   string          `json:"user_id"`
	Messages []model.Message `json:"messages"`
}

type updateMemoriesResponse struct {
	Memories []model.Memory `json:"memories"`
}

func (s *Server) handleUpdateMemories(w http.ResponseWriter, r *http.Request) {
	var req updateMemoriesRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	memories, err := s.service.UpdateFromMessages(r.Context(), req.UserID, req.Messages)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, updateMemoriesResponse{Memories: memories})
}

type searchMemoriesResponse struct {
	Results []model.ScoredMemory `json:"results"`
}

func (s *Server) handleSearchMemories(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := 10
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			limit = v
		}
	}
	results, err := s.service.Search(r.Context(), userID, q, limit)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, searchMemoriesResponse{Results: results})
}

type listMemoriesResponse struct {
	Memories []model.Memory `json:"memories"`
}

func (s *Server) handleListMemories(w http.ResponseWriter, r *http.Request) {
	userID := strings.TrimSpace(r.URL.Query().Get("user_id"))
	limit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			limit = v
		}
	}
	memories, err := s.service.List(r.Context(), userID, limit)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, listMemoriesResponse{Memories: memories})
}

func readJSON(r *http.Request, out any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return err
	}
	if dec.More() {
		return errors.New("invalid json body")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, errorResponse{Error: err.Error()})
}
