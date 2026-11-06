package api

import (
	"encoding/json"
	"net/http"
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

// Server holds the router and the stored loans.
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
	s.mux.HandleFunc("GET /loans/{id}", s.getLoan) // NEW in Lesson 3
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// writeJSON sends a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// GET /loans → 200 with every loan
func (s *Server) listLoans(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := make([]Loan, 0, len(s.loans))
	for _, l := range s.loans {
		list = append(list, l)
	}
	writeJSON(w, http.StatusOK, list)
}

// POST /loans → 201 with the new loan, or 400 on bad input
func (s *Server) createLoan(w http.ResponseWriter, r *http.Request) {
	var in Loan
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	pmt, err := loan.MonthlyPayment(in.Principal, in.Rate, in.Months)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	in.ID = s.nextID
	s.nextID++
	in.MonthlyPayment = pmt
	s.loans[in.ID] = in
	writeJSON(w, http.StatusCreated, in)
}

// GET /loans/{id} → 200 with the loan, 404 if missing, 400 if id isn't a number
// NEW in Lesson 3
func (s *Server) getLoan(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id")) // "7" → 7; "abc" → error
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "id must be a number"})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	l, ok := s.loans[id] // ok is false if no loan has this id
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "loan not found"})
		return
	}
	writeJSON(w, http.StatusOK, l)
}
