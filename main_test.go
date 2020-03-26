package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/spf13/viper"
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
		t.Errorf("handleHealth returned %v instead of %v", rec.Body.String(), "Good to Serve")
	}
}

func TestTemplateToJob(t *testing.T) {
	jobTemplate = template.Must(template.New("job").Parse(templateNomadJob()))
	f := func(a string, b string, c string, d string) bool {
		jobArgs := NomadJobData{
			GitRepoName: a,
			GitRepoURL:  b,
			HeadSHA:     c,
			SshKey:      d,
			Environment: []string{"CONSUL_TEST=1"},
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

func TestSetupEnvironment(t *testing.T) {
	os.Clearenv()
	setupEnvironment()
	if viper.GetString("GITHUB_SECRET") != "" {
		t.Error("AXIOMATIC_GITHUB_SECRET is not an empty string")
	}
	if viper.GetString("IP") != "127.0.0.1" {
		t.Error("AXIOMATIC_IP != 127.0.0.1")
	}
	if viper.GetString("PORT") != "8181" {
		t.Error("AXIOMATIC_PORT != 8181")
	}
	if viper.GetString("SSH_KEY") != "" {
		t.Error("AXIOMATIC_SSH_KEY is not an empty string")
	}
}

func TestStartupMessage(t *testing.T) {
	os.Clearenv()
	os.Setenv("TEST", "TestStartupMessage")
	actual := []byte(startupMessage())
	auFile := "testdata/TestStartupMessage.golden"
	if *update {
		ioutil.WriteFile(auFile, actual, 0644)
	}
	golden, err := ioutil.ReadFile(auFile)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(golden, actual) {
		t.Errorf("failed\nexpected:\n%s\ngot:\n%s", string(golden[:]), string(actual[:]))
	}

}
