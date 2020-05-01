job "axiomatic" {
  meta {
    repo = "http://github.com/code42/axiomatic"
    service = "axiomatic"
  }
  datacenters = ["dc1"]
  group "axiomatic" {
    task "axiomatic" {
      driver = "docker"

      config {
        image = "code42software/axiomatic:v1.0.1"
        port_map {
          http = 8181
        }
      }

      env {
        AXIOMATIC_IP = "0.0.0.0"
        AXIOMATIC_PORT = "8181"
        GITHUB_SECRET = "you-deserve-what-you-get"
        NOMAD_CACERT = "/secrets/certs/nomad-ca.pem"
        NOMAD_CLIENT_CERT = "/secrets/certs/cli.pem"
        NOMAD_CLIENT_KEY = "/secrets/certs/cli-key.pem"
      }
      template {
        data = <<EOH
{{ with secret "pki_int/issue/nomad-cluster" "ttl=24h" }}
{{ .Data.issuing_ca }}
{{ end }}
EOH
        destination = "/secrets/certs/nomad-ca.pem"
      }
      template {
        data = <<EOH
{{ with secret "pki_int/issue/nomad-cluster" "ttl=24h" }}
{{ .Data.certificate }}
{{ end }}
EOH
        destination = "/secrets/certs/cli.pem"
      }
      template {
        data = <<EOH
{{ with secret "pki_int/issue/nomad-cluster" "ttl=24h" }}
{{ .Data.private_key }}
{{ end }}
EOH
        destination = "/secrets/certs/cli-key.pem"
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
    policies = ["tls-policy"]
  }
}
