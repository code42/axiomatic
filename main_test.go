package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"testing"
)

// go test -update
var update = flag.Bool("update", false, "update .golden files")

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
