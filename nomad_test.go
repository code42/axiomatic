package main

import (
	"testing"
	"testing/quick"
	"text/template"
)

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
