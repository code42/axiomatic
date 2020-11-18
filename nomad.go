package main

import (
	"bytes"
	"errors"
	"log"

	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/jobspec"
)

// NomadJobData contains data for job template rendering
type NomadJobData struct {
	GitRepoName string
	GitRepoURL  string
	HeadSHA     string
	DeployKey   string
	Environment map[string]string
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
