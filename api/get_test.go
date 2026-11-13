package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// createTestLoan POSTs a loan and fails the test if that doesn't work.
func createTestLoan(t *testing.T, srv *Server) Loan {
	t.Helper()
	body := strings.NewReader(`{"principal": 25000, "rate": 0.045, "months": 60}`)
	req := httptest.NewRequest(http.MethodPost, "/loans", body)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup: POST /loans status = %d, want 201", rec.Code)
	}
	var l Loan
	json.NewDecoder(rec.Body).Decode(&l)
	return l
}

func TestGetLoanReturns200(t *testing.T) {
	srv := NewServer()
	created := createTestLoan(t, srv)

	req := httptest.NewRequest(http.MethodGet, "/loans/1", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /loans/1 status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got Loan
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if got != created {
		t.Errorf("got %+v, want %+v", got, created)
	}
}

func TestGetLoanErrors(t *testing.T) {
	tests := []struct {
		name string
		path string
		want int
	}{
		{"id that does not exist", "/loans/99", http.StatusNotFound},
		{"id that is not a number", "/loans/abc", http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := NewServer()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Errorf("GET %s status = %d, want %d", tc.path, rec.Code, tc.want)
			}
		})
	}
}
