package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListLoansReturns200(t *testing.T) {
	srv := NewServer()
	req := httptest.NewRequest(http.MethodGet, "/loans", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /loans status = %d, want %d", rec.Code, http.StatusOK)
	}
}
