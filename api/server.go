package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/patrick/go-finance-ci/loan"
)

// Server holds the router (and, later, the stored loans).
type Server struct {
	mux    *http.ServeMux
	mu     sync.Mutex   // a lock; see the note below
	loans  map[int]Loan // ID → loan
	nextID int
}

type Loan struct {
	ID             int     `json:"id"`
	Principal      float64 `json:"principal"`
	Rate           float64 `json:"rate"`
	Months         int     `json:"months"`
	MonthlyPayment float64 `json:"monthly_payment"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

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
	defer s.mu.Unlock() // unlock automatically when the function ends
	in.ID = s.nextID
	s.nextID++
	in.MonthlyPayment = pmt
	s.loans[in.ID] = in
	writeJSON(w, http.StatusCreated, in)
}

// NewServer builds a Server and registers its routes.
func NewServer() *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		loans:  make(map[int]Loan),
		nextID: 1,
	}
	s.mux.HandleFunc("GET /loans", s.listLoans)
	s.mux.HandleFunc("POST /loans", s.createLoan)
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
