package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/template"

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
	DeployKey   string
	Environment map[string]string
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

// filterEnvironment returns a map of strings from ss that begin with "CONSUL_" or "D2C_"
func filterEnvironment(ss []string) (map[string]string, error) {
	r := make(map[string]string)

	for _, s := range ss {
		if strings.HasPrefix(s, "CONSUL_") || strings.HasPrefix(s, "D2C_") {
			kv := strings.Split(s, "=")

			if len(kv) != 2 {
				return nil, fmt.Errorf("Error parsing environment variable: '%s'", s)
			}

			r[kv[0]] = kv[1]
		}
	}

	return r, nil
}

func setupEnvironment() {
	envDefaults := map[string]string{
		"GITHUB_SECRET": "",
		"IP":            "127.0.0.1",
		"PORT":          "8181",
		"SSH_PRIV_KEY":  "",
		"SSH_PUB_KEY":   "",
	}

	viper.SetEnvPrefix("AXIOMATIC")

	for key, val := range envDefaults {
		viper.SetDefault(key, val)
	}

	viper.AutomaticEnv()

	for key := range envDefaults {
		err := viper.BindEnv(key)
		if err != nil {
			log.Fatalf("Error setting up environment: %s", err)
		}
	}
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
	_, err := fmt.Fprintf(w, "Good to Serve")
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Good to Serve")
	}
	return
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.Printf("Error closing body: %s", err)
		}
	}()

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
		env, err := filterEnvironment(os.Environ())
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		jobArgs := NomadJobData{
			GitRepoName: e.Repo.GetName(),
			GitRepoURL:  e.Repo.GetSSHURL(),
			HeadSHA:     e.GetAfter(),
			DeployKey:   viper.GetString("SSH_PRIV_KEY"),
			Environment: env,
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
                    sshkey = "{{ .DeployKey }}"
                }
            }
            config {
                image = "code42software/dir2consul:v1.5.0"
            }
            driver = "docker"
            env {
                D2C_CONSUL_KEY_PREFIX = "services/{{ .GitRepoName }}/config"
                D2C_DIRECTORY = "/local/{{ .GitRepoName }}"
                CONSUL_HTTP_ADDR = "http://${attr.unique.network.ip-address}:8500"
            {{- range $key, $val := .Environment }}
                {{ $key }} = "{{ $val }}"
            {{- end }}
            }
            meta {
                commit-SHA = "{{ .HeadSHA }}"
            }
            vault = {
                policies = [ "consul-{{ .GitRepoName }}-write" ]
            }
            template {

                data = "CONSUL_HTTP_TOKEN={_ with secret \"consul/creds/{{ .GitRepoName }}-role\" _}{_ .Data.token _}{_ end _}"

                left_delimiter = "{_"
                right_delimiter = "_}"
                destination = "secrets/{{ .GitRepoName }}-token.env"
                env = true
            }

        }
    }
    type = "batch"
}
`
	return jobTemplate
}
