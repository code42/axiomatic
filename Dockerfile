############################
# STEP 1 build executable binary
############################
FROM golang:1.14.0-alpine3.11 as builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create appuser
ENV USER=appuser
ENV UID=10001

# See https://stackoverflow.com/a/55757473/12429735
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"
WORKDIR $GOPATH/src/mypackage/myapp/

# Fetch dependencies.
COPY go.* ./
RUN go mod download

COPY . ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags='-w -s -extldflags "-static"' -a \
      -o /go/bin/axiomatic .

############################
# STEP 2 build a small image
############################
FROM scratch

ARG BUILD_DATE
ARG VCS_REF
ARG VCS_URL
ARG VERSION

LABEL "org.opencontainers.image.authors"="Empower Rangers <empower-rangers@code42.com>"
LABEL "org.opencontainers.image.created"=$BUILD_DATE
LABEL "org.opencontainers.image.licenses"="https://github.com/code42/axiomatic/blob/master/LICENSE.md"
LABEL "org.opencontainers.image.revision"=$VCS_REF
LABEL "org.opencontainers.image.source"=$VCS_URL
LABEL "org.opencontainers.image.version"=$VERSION

# Import from builder.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy our static executable
COPY --from=builder /go/bin/axiomatic /go/bin/axiomatic

# Use an unprivileged user.
USER appuser:appuser

# Run the binary.
ENTRYPOINT ["/go/bin/axiomatic"]