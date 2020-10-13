# Change Log

## [Unreleased](#unreleased)

### Build

- **deps:** bump golang from 1.15.0-alpine3.12 to 1.15.2-alpine3.12
- **deps:** bump github.com/hashicorp/nomad from 0.12.3 to 0.12.4
- **deps:** bump golang from 1.14.6-alpine3.12 to 1.15.0-alpine3.12
- **deps:** bump actions/setup-go from v1 to v2.1.2
- **deps:** bump github.com/hashicorp/nomad from 0.12.1 to 0.12.3

### Chore

- automate releases via workflow
- uplift golang and libs
- add dependabot configuration

## [v1.1.0](#v1.1.0) - 2020-08-06

### Build

- **deps:** bump github.com/spf13/viper from 1.7.0 to 1.7.1
- **deps:** bump github.com/spf13/viper from 1.6.3 to 1.7.0

### Chore

- release new version
- update libraries and golang version ([#31](https://github.com/code42/axiomatic/issues/31))

### Docs

- address some markdown linting issues

### Fix

- Added dir2consul environment variable for the host-local consul agent and fixed the default vault path for consul ACL creds.
- s/jimrazmus/code42/

### Reverts

- Added two things:

### Pull Requests

- Merge pull request [#33](https://github.com/code42/axiomatic/issues/33) from code42/dependabot/go_modules/github.com/spf13/viper-1.7.1
- Merge pull request [#28](https://github.com/code42/axiomatic/issues/28) from code42/dependabot/go_modules/github.com/spf13/viper-1.7.0
- Merge pull request [#27](https://github.com/code42/axiomatic/issues/27) from code42/dm-final-dir2consul-job-adjustment
- Merge pull request [#24](https://github.com/code42/axiomatic/issues/24) from code42/adjusting-repo-artifact-stanza
- Merge pull request [#22](https://github.com/code42/axiomatic/issues/22) from code42/code42
- Merge pull request [#23](https://github.com/code42/axiomatic/issues/23) from code42/change-private-repo

## [v1.0.1](#v1.0.1) - 2020-04-27

### Chore

- release new version
- add container labels

### Pull Requests

- Merge pull request [#21](https://github.com/code42/axiomatic/issues/21) from code42/label-container

## [v1.0.0](#v1.0.0) - 2020-04-27

### Build

- add lint and errcheck to workflow
- Adopt simpler workflow and Docker Hub.
- remove debug var in favor of IDE debugging
- no codecov token required for public repos
- Remove redundant triggers. Add all subdirs to code coverage to future proof.
- Change workflow to Go 1.14
- Add labels to the docker image.
- **deps:** bump github.com/spf13/viper from 1.6.2 to 1.6.3

### Chore

- release new version
- update and tidy go modules
- go mod tidy
- update both codecov URLs
- axiomatic is now a code42 os project.
- remove obsolete Makefile

### Docs

- Remove extra header.
- Add verbiage for SSH key usage.
- make configuration section more clear
- sync with currect configuration behavior
- more link updates post move to Code42
- Add dir2consul configuration verbiage. Make other general improvements while here.
- Add Nomad config options now available with the API.
- Add webhook setup instructions.

### Feat

- Pass consul and d2c env vars to dir2consul
- Use ssh for repo cloning. Use an ssh key pair for accessing private repos.
- Pass CONSUL_* env vars to dir2consul.
- Switch handler to use the new functions.
- Use Nomad API for registering a Job.
- Use Nomad API for creating a Job
- update to golang 1.14.0
- Require github secret configuration.
- Begin using dir2consul
- Pass repo name as the Consul key prefix.

### Fix

- Update to dir2consul job and related processing ([#15](https://github.com/code42/axiomatic/issues/15))
- add err checking
- Update test to match linter change
- appease golint
- Use the new environment filter function
- make the ssh public key easy to fetch
- Pass the ssh private key in to the template
- Move config testing to a function.
- use Viper for env var handling
- Remove obsolete configuration variable.
- correct the destination and dirctory values
- use leading spaces, not tabs, in the job definition.
- repo cloning location Put it in local/repo to avoid potential collisions. Make dir2consul chdir to this repo location.
- detect and return err
- improve detecting/reporting on job response.
- log environment vars for debugging
- clean up imports
- remove obsolete variable and add whitespace.
- Compile the template once and reuse it.
- Change template to HCL
- Remove a trailing brace that was missed.

### Test

- update AU file for startupMessage test
- add test for startup message
- Adjust test for new template function.

### Pull Requests

- Merge pull request [#20](https://github.com/code42/axiomatic/issues/20) from code42/update-modules
- Merge pull request [#17](https://github.com/code42/axiomatic/issues/17) from code42/lint-errcheck
- Merge pull request [#16](https://github.com/code42/axiomatic/issues/16) from code42/workflow-overhaul
- Merge pull request [#13](https://github.com/code42/axiomatic/issues/13) from code42/dependabot/go_modules/github.com/spf13/viper-1.6.3
- Merge pull request [#12](https://github.com/code42/axiomatic/issues/12) from code42/d2c-env
- Merge pull request [#11](https://github.com/code42/axiomatic/issues/11) from code42/ssh
- Merge pull request [#10](https://github.com/code42/axiomatic/issues/10) from code42/use-viper
- Merge pull request [#9](https://github.com/code42/axiomatic/issues/9) from code42/inc-cov
- Merge pull request [#8](https://github.com/code42/axiomatic/issues/8) from jimrazmus/consul-env-vars
- Merge pull request [#7](https://github.com/code42/axiomatic/issues/7) from c42-dana-mckiernan/master
- Merge pull request [#6](https://github.com/code42/axiomatic/issues/6) from jimrazmus/use-nomad-api

## [0.10.0](#0.10.0) - 2020-02-18

[Unreleased]: https://github.com/code42/axiomatic/compare/v1.1.0...HEAD
[v1.1.0]: https://github.com/code42/axiomatic/compare/v1.0.1...v1.1.0
[v1.0.1]: https://github.com/code42/axiomatic/compare/v1.0.0...v1.0.1
[v1.0.0]: https://github.com/code42/axiomatic/compare/0.10.0...v1.0.0
