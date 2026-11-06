package api

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateLoanReturns201(t *testing.T) {
	srv := NewServer()
	// The request body: a JSON loan, wrapped so it can be read like a file
	body := strings.NewReader(`{"principal": 25000, "rate": 0.045, "months": 60}`)
	req := httptest.NewRequest(http.MethodPost, "/loans", body)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	// 1. Check the status code
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /loans status = %d, want %d", rec.Code, http.StatusCreated)
	}

	// 2. Turn the JSON response back into a Go value so we can check it
	var got Loan
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if got.ID != 1 {
		t.Errorf("ID = %d, want 1", got.ID)
	}
	if math.Abs(got.MonthlyPayment-466.08) > 0.01 {
		t.Errorf("MonthlyPayment = %.2f, want 466.08", got.MonthlyPayment)
	}
}
