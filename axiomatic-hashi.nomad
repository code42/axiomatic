job "axiomatic-hashi" {
  meta {
    repo = "http://github.com/code42/axiomatic"
    service = "axiomatic"
    repo_serving = "http://github.com/code42/cfg-hashi-versions"
  }
  constraint {
    attribute = "${node.class}"
    value     = "default"
  }
  datacenters = ["dc1"]
  group "axiomatic" {
    task "axiomatic" {
      driver = "docker"

      config {
        image = "code42software/axiomatic:v2.0.0"
        port_map {
          http = 8181
        }
      }

      env {
        AXIOMATIC_IP = "0.0.0.0"
        AXIOMATIC_PORT = "8181"
        AXIOMATIC_GITHUB_SECRET = "redacted"
        AXIOMATIC_SSH_PRIV_KEY = "redacted"
        AXIOMATIC_SSH_PUB_KEY = "redacted"
        AXIOMATIC_VERBOSE = "true"
        NOMAD_ADDR = "https://default-nomad-servers.us-east-1.us1-config-prod-02.cloud.code42.com"
      }
      template {
        data = <<EOH
NOMAD_TOKEN={{ with secret "secrets/team/empower-rangers/nomad-bootstrap-token" }}{{ .Data.token }}{{ end }}
EOH
        destination = "local/secrets.env"
        env         = true
      }

      resources {
        network {
          mode = "bridge"
          port "http" { }
        }
      }

      service {
        name = "axiomatic-hashi"
        port = "http"
        tags = [ "proxy" ]
      }
      meta {
        repo = "http://github.com/code42/axiomatic"
      }
    }
  }
  type = "service"

  vault = {
    policies = ["secrets-team-empower-rangers-read"]
  }
}