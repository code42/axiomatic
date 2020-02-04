package main

import (
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

func TestRenderNomadJobSucceeds(t *testing.T) {
	f := func(a string, b string, c string) bool {
		jobArgs := NomadJobData{
			GitRepoURL: a,
			HeadSHA:    b,
			Name:       c,
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
