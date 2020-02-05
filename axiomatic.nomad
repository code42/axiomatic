# Nomad job definition to run the Axiomatic service
job "axiomatic" {
  datacenters = ["dc1"]
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
    task "axiomatic" {
      config {
        image = "jimrazmus/axiomatic:rc"
      }
      driver = "docker"
      env {
        AXIOMATIC_IP = "0.0.0.0"
        AXIOMATIC_PORT = "8181"
      }
      resources {
        network {
          port "http" {
            static = 8181
          }
        }
      }
      service {
        name = "axiomatic"
        tags = ["global", "consul", "configuration"]
        port = "http"

        check {
          interval = "30s"
          path    = "/health"
          timeout  = "2s"
          type     = "http"
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
