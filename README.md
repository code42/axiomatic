# Axiomatic

Axiomatic is a GitHub webhook handler that launches dir2consul as a Nomad batch job.

## Configuration

Axiomatic uses environment variables to override the default configuration values. The variables are:

* AXIOMATIC_IP is the IP address to bind. Default: 127.0.0.1
* AXIOMATIC_PORT is the port number to bind. Default: 8181
* GITHUB_SECRET is the secret token for validating webhook requests. Please set it to something unique. Default: you-deserve-what-you-get
* NOMAD_SERVER is the URL of the Nomad server that will handle job submissions. Default: http://localhost:4646
* VAULT_TOKEN is the token value that will be added to submitted batch jobs. Default: ""

## Running with Docker

```bash
#> docker pull jimrazmus/axiomatic:vN.N.N
#> docker run -p 80:8181 --env-file=.env jimrazmus/axiomatic:vN.N.N
```

## Running with Nomad

```bash
#> nomad job plan axiomatic.nomad
#> nomad job run -check-index 0 axiomatic.nomad
```

## Health Check

Service health can be confirmed by making a web request to the '/health' path of the service.

## Add a GitHub Repo Webhook

TBD
