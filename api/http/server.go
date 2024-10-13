package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/umgbhalla/gokv/internal/query"
	"github.com/umgbhalla/gokv/internal/store"
)

type Server struct {
	store  *store.Store
	query  *query.Query
	router *mux.Router
	server *http.Server
}

func NewServer(store *store.Store, query *query.Query) *Server {
	s := &Server{
		store:  store,
		query:  query,
		router: mux.NewRouter(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) Start(addr string) error {
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) Router() *mux.Router {
	return s.router
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/get/{key}", s.handleGet).Methods("GET")
	s.router.HandleFunc("/set", s.handleSet).Methods("POST")
	s.router.HandleFunc("/delete/{key}", s.handleDelete).Methods("DELETE")
	s.router.HandleFunc("/query", s.handleQuery).Methods("GET")
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	value, ok := s.store.Get(key)

	if !ok {
		s.errorResponse(w, "Key not found", http.StatusNotFound)
		return
	}
	s.jsonResponse(w, map[string]interface{}{"value": value}, http.StatusOK)
}

func (s *Server) handleSet(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		s.errorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	key, ok := data["key"].(string)
	if !ok {
		s.errorResponse(w, "Invalid key", http.StatusBadRequest)
		return
	}

	value := data["value"]
	ttl := time.Duration(0)
	if ttlSeconds, ok := data["ttl"].(float64); ok {
		ttl = time.Duration(ttlSeconds) * time.Second
	}

	if err := s.store.Set(key, value, ttl); err != nil {
		s.errorResponse(w, "Error setting value", http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, map[string]string{"status": "ok"}, http.StatusOK)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	if err := s.store.Delete(key); err != nil {
		s.errorResponse(w, "Error deleting key", http.StatusInternalServerError)
		return
	}
	s.jsonResponse(w, map[string]string{"status": "ok"}, http.StatusOK)
}

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	queryString := r.URL.Query().Get("q")
	if queryString == "" {
		s.errorResponse(w, "Missing query parameter", http.StatusBadRequest)
		return
	}

	result, err := s.query.Execute(queryString)
	if err != nil {
		s.errorResponse(w, "Error executing query", http.StatusInternalServerError)
		return
	}

	s.jsonResponse(w, result, http.StatusOK)
}

func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) errorResponse(w http.ResponseWriter, message string, status int) {
	s.jsonResponse(w, map[string]string{"error": message}, status)
}
