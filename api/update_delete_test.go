package api

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// do sends one request to the server and returns the recorded response.
func do(srv *Server, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func TestListLoansReturnsCreatedLoans(t *testing.T) {
	srv := NewServer()
	createTestLoan(t, srv)
	createTestLoan(t, srv)

	rec := do(srv, http.MethodGet, "/loans", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var list []Loan
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if len(list) != 2 || list[0].ID != 1 || list[1].ID != 2 {
		t.Errorf("got %+v, want loans 1 and 2 in order", list)
	}
}

func TestUpdateLoanReturns200(t *testing.T) {
	srv := NewServer()
	createTestLoan(t, srv)

	rec := do(srv, http.MethodPut, "/loans/1", `{"principal": 1200, "rate": 0, "months": 12}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /loans/1 status = %d, want %d", rec.Code, http.StatusOK)
	}
	var got Loan
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
	if math.Abs(got.MonthlyPayment-100) > 0.01 {
		t.Errorf("MonthlyPayment = %.2f, want 100.00", got.MonthlyPayment)
	}

	// The change must stick: a later GET sees the new values.
	rec = do(srv, http.MethodGet, "/loans/1", "")
	var fetched Loan
	json.NewDecoder(rec.Body).Decode(&fetched)
	if fetched != got {
		t.Errorf("GET after PUT = %+v, want %+v", fetched, got)
	}
}

func TestUpdateLoanErrors(t *testing.T) {
	valid := `{"principal": 1000, "rate": 0.05, "months": 12}`
	tests := []struct {
		name string
		path string
		body string
		want int
	}{
		{"id that does not exist", "/loans/99", valid, http.StatusNotFound},
		{"id that is not a number", "/loans/abc", valid, http.StatusBadRequest},
		{"body is not json", "/loans/1", `hello`, http.StatusBadRequest},
		{"invalid loan values", "/loans/1", `{"principal": 0, "rate": 0.05, "months": 12}`, http.StatusBadRequest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := NewServer()
			createTestLoan(t, srv) // loan 1 exists in every case
			rec := do(srv, http.MethodPut, tc.path, tc.body)
			if rec.Code != tc.want {
				t.Errorf("PUT %s status = %d, want %d", tc.path, rec.Code, tc.want)
			}
		})
	}
}

func TestDeleteLoanReturns204(t *testing.T) {
	srv := NewServer()
	createTestLoan(t, srv)

	rec := do(srv, http.MethodDelete, "/loans/1", "")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE /loans/1 status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("204 response should have no body, got %q", rec.Body.String())
	}

	// The loan must really be gone.
	if rec := do(srv, http.MethodGet, "/loans/1", ""); rec.Code != http.StatusNotFound {
		t.Errorf("GET after DELETE status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDeleteLoanErrors(t *testing.T) {
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
			rec := do(srv, http.MethodDelete, tc.path, "")
			if rec.Code != tc.want {
				t.Errorf("DELETE %s status = %d, want %d", tc.path, rec.Code, tc.want)
			}
		})
	}
}

func TestUnsupportedMethodReturns405(t *testing.T) {
	srv := NewServer()
	rec := do(srv, http.MethodPatch, "/loans/1", "")
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("PATCH /loans/1 status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestResponsesAreJSON(t *testing.T) {
	srv := NewServer()
	rec := do(srv, http.MethodGet, "/loans", "")
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}
