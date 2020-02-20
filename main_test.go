package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"testing/quick"
)

func TestGetenvReturnsVal(t *testing.T) {
	f := func(a string, b string, c string) bool {
		os.Clearenv()
		// skip testing empty variable key
		if a == "" {
			return true
		}
		err := os.Setenv(a, b)
		if err != nil {
			t.Error(err)
		}
		// test we get back the randomly generated value
		if getenv(a, c) == b {
			return true
		}
		return false
	}
	err := quick.Check(f, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestGetenvReturnsDefault(t *testing.T) {
	f := func(a string, b string) bool {
		os.Clearenv()
		if getenv(a, b) == b {
			return true
		}
		return false
	}
	err := quick.Check(f, nil)
	if err != nil {
		t.Error(err)
	}
}

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
		t.Errorf("handleHealth returned %v instead of %v", rec.Body.String(), "Good to Serve")
	}
}

func TestRenderNomadJobSucceeds(t *testing.T) {
	f := func(a string, b string, c string, d string, e string, f string) bool {
		jobArgs := NomadJobData{
			ConsulKeyPrefix: a,
			ConsulServerURL: b,
			GitRepoName:     c,
			GitRepoURL:      d,
			HeadSHA:         e,
			VaultToken:      f,
		}
		_, err := renderNomadJob(jobArgs)
		if err != nil {
			t.Error(err)
		}
		return true
	}
	err := quick.Check(f, nil)
	if err != nil {
		t.Error(err)
	}
}
