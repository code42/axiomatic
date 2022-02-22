job "axiomatic" {
  meta {
    repo = "http://github.com/code42/axiomatic"
    service = "axiomatic"
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
        image = "code42software/axiomatic:v1.1.0"
        port_map {
          http = 8181
        }
      }

      env {
        AXIOMATIC_IP = "0.0.0.0"
        AXIOMATIC_PORT = "8181"
        GITHUB_SECRET = "you-deserve-what-you-get"
      }
      template {
        data = <<EOH
NOMAD_TOKEN={{ with secret "secrets/team/empower-rangers/nomad-bootstrap-token" }}
{{ .Data.token }}
{{ end }}
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
        name = "axiomatic"
        port = "http"
        tags = [ "proxy" ]

        connect {
          sidecar_service {
          }
        }
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
