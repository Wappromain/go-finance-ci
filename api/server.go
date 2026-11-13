// Package api exposes the loan calculator as a JSON REST API.
//
//	GET    /loans       200 list of loans
//	POST   /loans       201 created loan, 400 bad input
//	GET    /loans/{id}  200 loan, 400 bad id, 404 not found
//	PUT    /loans/{id}  200 updated loan, 400 bad id or input, 404 not found
//	DELETE /loans/{id}  204 no content, 400 bad id, 404 not found
package api

import (
	"encoding/json"
	"math"
	"net/http"
	"sort"
	"strconv"
	"sync"

	"github.com/patrick/go-finance-ci/loan"
)

// Loan is what clients send and receive. The json tags set the field names.
type Loan struct {
	ID             int     `json:"id"`
	Principal      float64 `json:"principal"`
	Rate           float64 `json:"rate"`
	Months         int     `json:"months"`
	MonthlyPayment float64 `json:"monthly_payment"`
}

// Server holds the router and the stored loans (in memory).
type Server struct {
	mux    *http.ServeMux
	mu     sync.Mutex   // makes simultaneous requests take turns with the map
	loans  map[int]Loan // ID → loan
	nextID int
}

// NewServer sets up empty storage and registers the routes.
func NewServer() *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		loans:  make(map[int]Loan),
		nextID: 1,
	}
	s.mux.HandleFunc("GET /loans", s.listLoans)
	s.mux.HandleFunc("POST /loans", s.createLoan)
	s.mux.HandleFunc("GET /loans/{id}", s.getLoan)
	s.mux.HandleFunc("PUT /loans/{id}", s.updateLoan)
	s.mux.HandleFunc("DELETE /loans/{id}", s.deleteLoan)
	return s
}

// ServeHTTP makes Server an http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// writeJSON sends a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// pathID reads {id} from the URL. ok is false (and a 400 has been sent)
// when the id is not a number.
func pathID(w http.ResponseWriter, r *http.Request) (id int, ok bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return 0, false
	}
	return id, true
}

// decodeLoan reads and validates a loan from the request body, filling in
// MonthlyPayment (rounded to cents). ok is false (and a 400 has been sent) on bad input.
func decodeLoan(w http.ResponseWriter, r *http.Request) (in Loan, ok bool) {
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return Loan{}, false
	}
	pmt, err := loan.MonthlyPayment(in.Principal, in.Rate, in.Months)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return Loan{}, false
	}
	in.MonthlyPayment = math.Round(pmt*100) / 100 // round to cents
	return in, true
}

// GET /loans → 200 with every loan, sorted by ID
func (s *Server) listLoans(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := make([]Loan, 0, len(s.loans))
	for _, l := range s.loans {
		list = append(list, l)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	writeJSON(w, http.StatusOK, list)
}

// POST /loans → 201 with the new loan, or 400 on bad input
func (s *Server) createLoan(w http.ResponseWriter, r *http.Request) {
	in, ok := decodeLoan(w, r)
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	in.ID = s.nextID
	s.nextID++
	s.loans[in.ID] = in
	writeJSON(w, http.StatusCreated, in)
}

// GET /loans/{id} → 200, 400 or 404
func (s *Server) getLoan(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	l, found := s.loans[id]
	if !found {
		writeError(w, http.StatusNotFound, "loan not found")
		return
	}
	writeJSON(w, http.StatusOK, l)
}

// PUT /loans/{id} → 200 with the replaced loan, 400 or 404
func (s *Server) updateLoan(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	in, ok := decodeLoan(w, r)
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, found := s.loans[id]; !found {
		writeError(w, http.StatusNotFound, "loan not found")
		return
	}
	in.ID = id // the URL decides the ID, not the body
	s.loans[id] = in
	writeJSON(w, http.StatusOK, in)
}

// DELETE /loans/{id} → 204 with no body, 400 or 404
func (s *Server) deleteLoan(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, found := s.loans[id]; !found {
		writeError(w, http.StatusNotFound, "loan not found")
		return
	}
	delete(s.loans, id)
	w.WriteHeader(http.StatusNoContent)
}
