package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"testing/quick"
	"text/template"

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
		t.Errorf("handleHealth returned (%v) instead of (%v)", rec.Body.String(), "Good to Serve")
	}
}

func TestPublicKey(t *testing.T) {
	answer := "an ssh public key"
	err := os.Setenv("AXIOMATIC_SSH_PUB_KEY", answer)
	if err != nil {
		t.Fatal(err)
	}
	setupEnvironment()
	req, err := http.NewRequest("GET", "/publickey", nil)
	if err != nil {
		t.Error(err)
	}
	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(handlePublicKey)
	handler.ServeHTTP(rec, req)
	if status := rec.Code; status != http.StatusOK {
		t.Errorf("handlePublicKey returned %v instead of %v", status, http.StatusOK)
	}
	if rec.Body.String() != "an ssh public key" {
		t.Errorf("handlePublicKey returned (%v) instead of (%v)", rec.Body.String(), answer)
	}
}

func TestTemplateToJob(t *testing.T) {
	jobTemplate = template.Must(template.New("job").Parse(templateNomadJob()))
	f := func(a string, b string, c string, d string) bool {
		jobArgs := NomadJobData{
			GitRepoName: a,
			GitRepoURL:  b,
			HeadSHA:     c,
			SSHKey:      d,
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

func TestFilterEnvironmentSucceeds(t *testing.T) {
	cases := []struct {
		name string
		ss   []string
		rs   map[string]string
	}{
		{
			"One Element Match",
			[]string{"CONSUL_A=a"},
			map[string]string{"CONSUL_A": "a"},
		},
		{
			"Two Element Match",
			[]string{"CONSUL_A=a", "CONSUL_b=b"},
			map[string]string{"CONSUL_A": "a", "CONSUL_b": "b"},
		},
		{
			"Leading Element Match",
			[]string{"CONSUL_1=1", "a=a", "b=b"},
			map[string]string{"CONSUL_1": "1"},
		},
		{
			"Nested Element Match",
			[]string{"a=a", "CONSUL_1=1", "b=b"},
			map[string]string{"CONSUL_1": "1"},
		},
		{
			"Trailing Element Match",
			[]string{"a=a", "b=b", "CONSUL_1=1"},
			map[string]string{"CONSUL_1": "1"},
		},
		{
			"No Element Match",
			[]string{"a=a", "b=b", "CONSULA_A=a"},
			map[string]string{},
		},
		{
			"One d2c Element Match",
			[]string{"a=a", "b=b", "D2C_A=a"},
			map[string]string{"D2C_A": "a"},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			got, _ := filterEnvironment(tc.ss)
			if !reflect.DeepEqual(got, tc.rs) {
				t.Errorf("got (%+v) want (%+v)", got, tc.rs)
			}
		})
	}
}

func TestFilterEnvironmentErrors(t *testing.T) {
	cases := []struct {
		name string
		ss   []string
	}{
		{
			"Two equals errors",
			[]string{"CONSUL_A=a=c"},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			_, err := filterEnvironment(tc.ss)
			if err == nil {
				t.Error("expected an error")
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
	if viper.GetString("SSH_PRIV_KEY") != "" {
		t.Error("AXIOMATIC_SSH_PRIV_KEY is not an empty string")
	}
	if viper.GetString("SSH_PUB_KEY") != "" {
		t.Error("AXIOMATIC_SSH_PUB_KEY is not an empty string")
	}
}

func TestIsMissingConfiguration(t *testing.T) {
	os.Clearenv()
	if isMissingConfiguration() != true {
		t.Error("expected: (true) got: (false)")
	}
	err := os.Setenv("AXIOMATIC_GITHUB_SECRET", "testing")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Setenv("AXIOMATIC_SSH_PRIV_KEY", "testing")
	if err != nil {
		t.Fatal(err)
	}
	err = os.Setenv("AXIOMATIC_SSH_PUB_KEY", "testing")
	if err != nil {
		t.Fatal(err)
	}

	if isMissingConfiguration() != false {
		t.Error("expected: (false) got: (true)")
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
