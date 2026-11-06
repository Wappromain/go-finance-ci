package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/patrick/go-finance-ci/loan" // calculator package
)

// The shape of a loan, in Go and in JSON.
// The `json:"..."` tags set the field names clients see.
type Loan struct {
	ID             int     `json:"id"`
	Principal      float64 `json:"principal"`
	Rate           float64 `json:"rate"`
	Months         int     `json:"months"`
	MonthlyPayment float64 `json:"monthly_payment"`
}

// The server now remembers loans.
type Server struct {
	mux    *http.ServeMux
	mu     sync.Mutex   // makes simultaneous requests take turns with the map
	loans  map[int]Loan // ID → loan
	nextID int          // the ID the next new loan gets
}

// Set up empty storage and register both routes.
func NewServer() *Server {
	s := &Server{
		mux:    http.NewServeMux(),
		loans:  make(map[int]Loan), // a map must be made before use
		nextID: 1,
	}
	s.mux.HandleFunc("GET /loans", s.listLoans)
	s.mux.HandleFunc("POST /loans", s.createLoan)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// One helper for every JSON response: header, status, body.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// List every stored loan.
func (s *Server) listLoans(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock() // "defer" = run this when the function returns

	list := make([]Loan, 0, len(s.loans)) // empty slice, not nil, so JSON is [] not null
	for _, l := range s.loans {
		list = append(list, l)
	}
	writeJSON(w, http.StatusOK, list)
}

// Create a loan from the request body.
func (s *Server) createLoan(w http.ResponseWriter, r *http.Request) {
	// Read the JSON body into a Loan.
	var in Loan
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return // stop here; don't save anything
	}

	// Let the loan package validate and calculate.
	pmt, err := loan.MonthlyPayment(in.Principal, in.Rate, in.Months)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Save it with a fresh ID.
	s.mu.Lock()
	defer s.mu.Unlock()
	in.ID = s.nextID
	s.nextID++
	in.MonthlyPayment = pmt
	s.loans[in.ID] = in

	// Reply 201 Created with the saved loan.
	writeJSON(w, http.StatusCreated, in)
}
