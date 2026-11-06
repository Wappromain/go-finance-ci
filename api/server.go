package api

import (
	"encoding/json"
	"net/http"
)

// Server holds the router (and, later, the stored loans).
type Server struct {
	mux *http.ServeMux
}

// NewServer builds a Server and registers its routes.
func NewServer() *Server {
	s := &Server{mux: http.NewServeMux()}       // 1: make a Server with a new router
	s.mux.HandleFunc("GET /loans", s.listLoans) // 2: GET /loans → listLoans
	return s
}

// ServeHTTP makes Server an http.Handler, so tests and the real
// server can both use it.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) listLoans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") // 5: tell the client it's JSON
	w.WriteHeader(http.StatusOK)                       // 6: status 200
	json.NewEncoder(w).Encode([]string{})
}
