package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
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

// VaultToken is the token used to access the Nomad server
var VaultToken = getenv("VAULT_TOKEN", "")

var jobTemplate *template.Template

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

	jobTemplate = template.Must(template.New("job").Parse(templateNomadJob()))

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

		job, err := templateToJob(jobArgs)
		if err != nil {
			log.Println("template to job failed:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = submitNomadJob(job)
		if err != nil {
			log.Println("submit job failed:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		log.Printf("WARN: unknown event type %s\n", github.WebHookType(r))
		return
	}
}

func templateToJob(jobArgs NomadJobData) (*api.Job, error) {
	var buf bytes.Buffer

	// execute template with given data and output to io pipe
	err := jobTemplate.Execute(&buf, jobArgs)
	if err != nil {
		return nil, err
	}

	// create a Nomad job struct by parsing data from the io pipe
	var job *api.Job
	job, err = jobspec.Parse(&buf)
	if err != nil {
		return nil, err
	}

	return job, nil
}

// submitNomadJob sends a job to a Nomad server
func submitNomadJob(job *api.Job) error {
	nomadClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Println("Error establishing Nomad Client:", err)
		return err
	}

	var jobResp *api.JobRegisterResponse
	jobResp, _, err = nomadClient.Jobs().Register(job, nil)
	if jobResp.Warnings != "" {
		log.Printf("Eval Warning (%s) %s", jobResp.EvalID, jobResp.Warnings)
		return errors.New(jobResp.Warnings)
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
				destination = "local/{{ .GitRepoName }}"
				source = "{{ .GitRepoURL }}"
			}
			config {
				image = "jimrazmus/dir2consul:v1.3.0"
			}
			driver = "docker"
			env {
				D2C_CONSUL_KEY_PREFIX = "services/{{ .GitRepoName }}/config"
				D2C_CONSUL_SERVER = "{{ .ConsulServerURL }}"
			}
			meta {
				commit-SHA = "{{ .HeadSHA }}"
			}
			vault = {
				policies = ["consul-{{ .GitRepoName }}-write"]
			}
		}
	}
	type = "batch"
}
`
	return jobTemplate
}
