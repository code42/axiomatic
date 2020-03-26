# Axiomatic

[![Go Report Card](https://goreportcard.com/badge/github.com/code42/axiomatic)](https://goreportcard.com/report/github.com/code42/axiomatic)
![Build](https://github.com/code42/axiomatic/workflows/Go/badge.svg?branch=master)
[![CodeCov](https://codecov.io/gh/code42/axiomatic/branch/master/graph/badge.svg)](https://codecov.io/gh/code42/axiomatic)
[![License](http://img.shields.io/:license-mit-blue.svg?style=flat-square)](http://badges.mit-license.org)

## Summary

Axiomatic is a GitHub webhook handler that launches [dir2consul](https://github.com/code42/dir2consul) as a Nomad batch job.

## Configuration

Axiomatic uses environment variables to override the default configuration values. The Nomad job definition for the Axiomatic service should be adjusted to set these variables.

### Axiomatic

* AXIOMATIC_GITHUB_SECRET (**required**) is the secret token for validating webhook requests. There is no default value, only sorrow.
* AXIOMATIC_IP is the IP address to bind. Default = 127.0.0.1
* AXIOMATIC_PORT is the port number to bind. Default = 8181
* AXIOMATIC_SSH_PRIV_KEY (**required**) is the private ssh key used for cloning repositories. It must be base64 encoded.
* AXIOMATIC_SSH_PUB_KEY (**required**) is the public ssh key used for cloning repositories.
* NOMAD_ADDR is the address of the Nomad server. Default = http://127.0.0.1:4646
* NOMAD_CACERT is the path to a PEM encoded CA cert file to use to verify the Nomad server SSL certificate.
* NOMAD_CAPATH is the path to a directory of PEM encoded CA cert files to verify the Nomad server SSL certificate.
* NOMAD_CLIENT_CERT Path to a PEM encoded client certificate for TLS authentication to the Nomad server.
* NOMAD_CLIENT_KEY Path to an unencrypted PEM encoded private key matching the client certificate.
* NOMAD_NAMESPACE is the target namespace for queries and actions. Default = "default"
* NOMAD_REGION is region of the Nomad servers to forward commands.
* NOMAD_TOKEN is the SecretID of an ACL token to use to authenticate API requests.

### dir2consul

The following configuration variables are passed to dir2consul

* D2C_CONSUL_KEY_PREFIX is the path prefix to prepend to all consul keys. Default: ""

Axiomatic passes any environment variables beginning with "CONSUL_" through to dir2consul. This provides a convenient way to configure the dir2consul batch jobs launched by Axiomatic.

## Installation

Axiomatic requires no installation. It ships as a Docker container meant for running as a service.

### Run the service on Nomad

```bash
#> nomad job plan axiomatic.nomad
#> nomad job run -check-index 0 axiomatic.nomad
```

### Setup the GitHub Repo Webhook and SSH Key

1. Open the Settings tab for your repo and choose "Deploy keys" in the menu
1. Add the ssh public key as a repository deploy key
1. Open the Settings tab for your repo and choose "Webhooks" in the menu
1. Press the "Add webhook" button
1. Enter your Axiomatic service URL in the Payload URL field. E.g. "https://axiomatic.example.com/webhook"
1. Enter your Axiomatic github secret in the Secret field
1. Press the "Add webhook" button

## Health Check

Service health can be confirmed by making a web request to the '/health' path of the service.

## Vault Policy

Axiomatic needs a Vault policy that allows the service to submit batch jobs.

*example policy TBD*

## Contributing

Please follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification for your commit messages. Commit type options include: feat, fix, build, chore, ci, docs, style, refactor, perf, and test.

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct, and the process for submitting pull requests to us.

## Author

Jim Razmus II

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

