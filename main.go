package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/google/go-github/github"
	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
	"github.com/spf13/viper"
)

var jobTemplate *template.Template

// NomadJobData contains data for job template rendering
type NomadJobData struct {
	GitRepoName string
	GitRepoURL  string
	HeadSHA     string
	SshKey      string
	Environment []string
}

func main() {
	setupEnvironment()
	if isMissingConfiguration() {
		log.Fatal("Shutting down.")
	}
	fmt.Println(startupMessage())

	jobTemplate = template.Must(template.New("job").Parse(templateNomadJob()))

	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/webhook", handleWebhook)

	log.Fatal(http.ListenAndServe(viper.GetString("IP")+":"+viper.GetString("PORT"), nil))
	return
}

// filterConsul returns a slice of strings from ss that begin with "CONSUL_"
func filterConsul(ss []string) []string {
	r := []string{}
	for _, s := range ss {
		if strings.HasPrefix(s, "CONSUL_") {
			r = append(r, s)
		}
	}
	return r
}

func setupEnvironment() {
	viper.SetEnvPrefix("AXIOMATIC")
	viper.SetDefault("GITHUB_SECRET", "")
	viper.SetDefault("IP", "127.0.0.1")
	viper.SetDefault("PORT", "8181")
	viper.SetDefault("SSH_PRIV_KEY", "")
	viper.SetDefault("SSH_PUB_KEY", "")
	viper.AutomaticEnv()
	viper.BindEnv("GITHUB_SECRET")
	viper.BindEnv("IP")
	viper.BindEnv("PORT")
	viper.BindEnv("SSH_PRIV_KEY")
	viper.BindEnv("SSH_PUB_KEY")
}

func isMissingConfiguration() bool {
	r := false
	vs := []string{"GITHUB_SECRET", "SSH_PRIV_KEY", "SSH_PUB_KEY"}
	for _, v := range vs {
		if viper.GetString(v) == "" {
			log.Printf("You must configure AXIOMATIC_%s!", v)
			r = true
		}
	}
	return r
}

func startupMessage() string {
	banner := "\n------------\n Axiomatic \n------------\n"

	config := fmt.Sprintf("Configuration\n\tAXIOMATIC_IP: %s\n\tAXIOMATIC_PORT: %s", viper.GetString("IP"), viper.GetString("PORT"))

	env := os.Environ()
	sort.Strings(env)
	environment := fmt.Sprintf("\nEnvironment\n\t%s", strings.Join(env, "\n\t"))

	return banner + config + environment
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	log.Println("Good to Serve")
	fmt.Fprintf(w, "Good to Serve")
	return
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	payload, err := github.ValidatePayload(r, []byte(viper.GetString("GITHUB_SECRET")))
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
			GitRepoName: e.Repo.GetFullName(),
			GitRepoURL:  e.Repo.GetSSHURL(),
			HeadSHA:     e.GetAfter(),
			SshKey:      viper.GetString("SSH_PRIV_KEY"),
			Environment: filterConsul(os.Environ()),
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
	if err != nil {
		return err
	}
	if jobResp == (*api.JobRegisterResponse)(nil) {
		return errors.New("jobResp is nil")
	}
	if jobResp.Warnings != "" {
		log.Printf("Nomad job response: %+v", jobResp)
		return errors.New(jobResp.Warnings)
	}

	return nil
}

// templateNomadJob returns a Nomad job definition as a string
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
                options {
                    sshkey = "{{ .SshKey }}"
                }
            }
            config {
                image = "jimrazmus/dir2consul:v1.4.1"
            }
            driver = "docker"
            env {
                D2C_CONSUL_KEY_PREFIX = "services/{{ .GitRepoName }}/config"
                D2C_DIRECTORY = "local/{{ .GitRepoName }}"
			{{ range .Environment}}
				{{.}}
			{{ end }}
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
