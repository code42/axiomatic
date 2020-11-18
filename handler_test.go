package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleHealth(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Error(err)
	}
	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(handleHealth)
	handler.ServeHTTP(rec, req)
	if status := rec.Code; status != http.StatusOK {
		t.Errorf("handleHealth returned %v instead of %v", status, http.StatusOK)
	}
	if rec.Body.String() != "Good to Serve" {
		t.Errorf("handleHealth returned (%v) instead of (%v)", rec.Body.String(), "Good to Serve")
	}
}
