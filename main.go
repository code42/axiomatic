package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"

	"github.com/google/go-github/github"
)

// conditionally compile in or out the debug prints
const debug = false

// AxiomaticIP is the IP address to bind
var AxiomaticIP = getenv("AXIOMATIC_IP", "127.0.0.1")

// AxiomaticPort is the port number to bind
var AxiomaticPort = getenv("AXIOMATIC_PORT", "8181")

// ConsulKeyPrefix is the path prefix to prepend to all consul keys
var ConsulKeyPrefix = getenv("D2C_CONSUL_KEY_PREFIX", "")

// ConsulServerURL is the URL of the Consul server kv store
var ConsulServerURL = getenv("D2C_CONSUL_SERVER", "http://localhost:8500/v1/kv")

// GithubWebhookSecret is the secret token for validating webhook requests
var GithubWebhookSecret = getenv("GITHUB_SECRET", "")

// NomadServerURL is the URL of the Nomad server that will handle job submissions
var NomadServerURL = getenv("NOMAD_SERVER", "http://localhost:4646")

// VaultToken is the token used to access the Nomad server
var VaultToken = getenv("VAULT_TOKEN", "")

// NomadJobData contains data for job template rendering
type NomadJobData struct {
	ConsulKeyPrefix string
	ConsulServerURL string
	GitRepoName     string
	GitRepoURL      string
	HeadSHA         string
	VaultToken      string
}

func main() {
	log.Println("Axiomatic Server Starting")
	if GithubWebhookSecret == "" {
		log.Fatal("You must configure GITHUB_SECRET! Axiomatic shutting down.")
	}
	log.Println("AXIOMATIC_IP:", AxiomaticIP)
	log.Println("AXIOMATIC_PORT:", AxiomaticPort)
	log.Println("NOMAD_SERVER:", NomadServerURL)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/webhook", handleWebhook)
	serverAddr := strings.Join([]string{AxiomaticIP, AxiomaticPort}, ":")
	log.Fatal(http.ListenAndServe(serverAddr, nil))
	return
}

// getenv returns the environment value for the given key or the default value when not found
func getenv(key string, _default string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return _default
	}
	return val
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	log.Println("Good to Serve")
	fmt.Fprintf(w, "Good to Serve")
	return
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	payload, err := github.ValidatePayload(r, []byte(GithubWebhookSecret))
	if err != nil {
		log.Printf("error validating request body: err=%s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("could not parse webhook: err=%s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch e := event.(type) {
	case *github.PingEvent:
		log.Println("GitHub Pinged the Webhook")
	case *github.PushEvent:
		jobArgs := NomadJobData{
			ConsulKeyPrefix: ConsulKeyPrefix,
			ConsulServerURL: ConsulServerURL,
			GitRepoName:     e.Repo.GetFullName(),
			GitRepoURL:      e.Repo.GetCloneURL(),
			HeadSHA:         e.GetAfter(),
			VaultToken:      VaultToken,
		}
		if debug {
			log.Printf("jobArgs: %+v\n", jobArgs)
		}

		jobText, err := renderNomadJob(jobArgs)
		if err != nil {
			log.Println("renderNomamdJob Error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if debug {
			log.Println("jobText:", jobText)
		}

		err = submitNomadJob(jobArgs.GitRepoName, jobText)
		if err != nil {
			log.Println("submitJob Error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("submitJob Success: %s (%s)", jobArgs.GitRepoName, jobArgs.HeadSHA)
		fmt.Fprintln(w, "Nomad Job Submitted")
	default:
		log.Printf("WARN: unknown event type %s\n", github.WebHookType(r))
		return
	}
}

// renderNomadJob combines a template with supplied args and returns a Nomad job definition as a string
func renderNomadJob(jobArgs NomadJobData) (*bytes.Buffer, error) {
	t := template.Must(template.New("job").Parse(templateNomadJob()))
	buf := &bytes.Buffer{}
	err := t.Execute(buf, jobArgs)
	if err != nil {
		return buf, err
	}
	return buf, nil
}

// submitNomadJob sends a job to a Nomad server REST API
func submitNomadJob(jobName string, jobBody *bytes.Buffer) error {
	url := strings.Join([]string{NomadServerURL, "v1/job", url.PathEscape(jobName)}, "/")
	if debug {
		log.Println("URL:", url)
	}

	request, err := http.NewRequest("POST", url, jobBody)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if debug {
		body, _ := ioutil.ReadAll(response.Body)
		log.Println("response Body:", string(body))
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return errors.New(response.Status)
	}
	return nil
}

// templateNomadJob returns a templated,json formatted, Nomad job definition as a string
func templateNomadJob() string {
	const jobTemplate = `
job "dir2consul-{{ .GitRepoName }}" {
	datacenters = ["dc1"]
	region = "global"
	group "dir2consul" {
		task "dir2consul" {
			artifact {
				destination = "local/"
				source = "{{ .GitRepoURL }}"
			}
			config {
				image = "jimrazmus/dir2consul:v1.1.0"
			}
			driver = "docker"
			env {
				D2C_CONSUL_KEY_PREFIX = "{{ .ConsulKeyPrefix }}"
				D2C_CONSUL_SERVER = "{{ .ConsulServerURL }}"
			}
			meta {
				commit-SHA = "{{ .HeadSHA }}"
			}
			# vault {}
		}
	}
	type = "batch"
	# vault = {}
}
`
	return jobTemplate
}
