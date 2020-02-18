# Nomad job definition to run the Axiomatic service
job "axiomatic" {
  datacenters = ["dc1"]
  meta {
    repo = "http://github.com/jimrazmus/axiomatic"
  }
  region = "global"
  type = "service"
  update {
    auto_revert = true
    canary = 1
    healthy_deadline = "3m"
    min_healthy_time = "10s"
    progress_deadline = "10m"
  }

  group "axiomatic" {
    service {
      check {
        interval = "30s"
        path    = "/health"
        timeout  = "2s"
        type     = "http"
      }
      connect {
        sidecar_service {
          proxy {
            upstreams {
              destination_name = "nomad"
              local_bind_port = "4646"
            }
          }
        }
      }
      meta {
        repo = "http://github.com/jimrazmus/axiomatic"
      }
      name = "axiomatic"
      port = "8181"
      tags = ["global", "consul", "configuration"]
    }
    task "axiomatic" {
      config {
        image = "jimrazmus/axiomatic:rc"
      }
      driver = "docker"
      env {
        AXIOMATIC_IP = "0.0.0.0"
        AXIOMATIC_PORT = "8181"
        GITHUB_SECRET = "you-deserve-what-you-get"
      }
      resources {
        network {
          port "http" {}
        }
      }
      # vault {
      #   policies      = ["axiomatic"]
      #   change_mode   = "signal"
      #   change_signal = "SIGHUP"
      # }
    }
  }
}
