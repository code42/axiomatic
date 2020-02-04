VERSION=`git rev-parse HEAD`
BUILD=`date +%FT%T%z`
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

.PHONY: build
build:
	@export DOCKER_CONTENT_TRUST=1 && docker build -f Dockerfile -t axiomatic .

.PHONY: run
run:
	@docker run -p 127.0.0.1:8181:8181 --env-file=.env axiomatic:latest