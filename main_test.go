package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
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

func TestTemplateToJob(t *testing.T) {
	jobTemplate = template.Must(template.New("job").Parse(templateNomadJob()))
	f := func(a string, b string, c string, d string, e string) bool {
		jobArgs := NomadJobData{
			ConsulKeyPrefix: a,
			GitRepoName:     b,
			GitRepoURL:      c,
			HeadSHA:         d,
			VaultToken:      e,
			Environment:     []string{"CONSUL_TEST=1"},
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

func TestFilterConsul(t *testing.T) {
	cases := []struct {
		name string
		ss   []string
		rs   []string
	}{
		{
			"One Element Match",
			[]string{"CONSUL_A=a"},
			[]string{"CONSUL_A=a"},
		},
		{
			"Two Element Match",
			[]string{"CONSUL_A=a", "CONSUL_b=b"},
			[]string{"CONSUL_A=a", "CONSUL_b=b"},
		},
		{
			"Leading Element Match",
			[]string{"CONSUL_1=1", "a=a", "b=b"},
			[]string{"CONSUL_1=1"},
		},
		{
			"Nested Element Match",
			[]string{"a=a", "CONSUL_1=1", "b=b"},
			[]string{"CONSUL_1=1"},
		},
		{
			"Trailing Element Match",
			[]string{"a=a", "b=b", "CONSUL_1=1"},
			[]string{"CONSUL_1=1"},
		},
		{
			"No Element Match",
			[]string{"a=a", "b=b", "CONSULA_A=a"},
			[]string{},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			got := filterConsul(tc.ss)
			if !reflect.DeepEqual(got, tc.rs) {
				t.Errorf("got (%+v) want (%+v)", got, tc.rs)
			}
		})
	}
}
