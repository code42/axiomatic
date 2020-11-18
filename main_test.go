package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"testing/quick"
	"text/template"
)

// go test -update
var update = flag.Bool("update", false, "update .golden files")

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

func TestTemplateToJob(t *testing.T) {
	jobTemplate = template.Must(template.New("job").Parse(templateNomadJob()))
	f := func(a string, b string, c string, d string) bool {
		jobArgs := NomadJobData{
			GitRepoName: a,
			GitRepoURL:  b,
			HeadSHA:     c,
			DeployKey:   d,
			Environment: map[string]string{"CONSUL_TEST": "1"},
		}
		_, err := templateToJob(jobArgs)
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

func TestStartupMessage(t *testing.T) {
	os.Clearenv()
	err := os.Setenv("TEST", "TestStartupMessage")
	if err != nil {
		t.Fatal(err)
	}
	actual := []byte(startupMessage())
	auFile := "testdata/TestStartupMessage.golden"
	if *update {
		err = ioutil.WriteFile(auFile, actual, 0644)
		if err != nil {
			t.Fatal(err)
		}
	}
	golden, err := ioutil.ReadFile(auFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(golden, actual) {
		t.Errorf("failed\nexpected:\n%s\ngot:\n%s", string(golden[:]), string(actual[:]))
	}

}
