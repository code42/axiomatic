job "dir2consul-{{ .GitRepoName }}" {
    datacenters = ["dc1"]
    region = "global"
    group "dir2consul" {
        task "dir2consul" {
            artifact {
                destination = "local/{{ .GitRepoName }}"
				source = "{{ .GitRepoURL }}"
                options {
                    sshkey = "{{ .SshKey }}"
                }
            }
            config {
                image = "code42software/dir2consul:v1.4.1"
            }
            driver = "docker"
            env {
                D2C_CONSUL_KEY_PREFIX = "services/{{ .GitRepoName }}/config"
                D2C_DIRECTORY = "local/{{ .GitRepoName }}"
			{{ range .Environment}}
				{{.}}
			{{ end }}
            }
            meta {
                commit-SHA = "{{ .HeadSHA }}"
            }
            vault = {
                policies = ["consul-{{ .GitRepoName }}-write"]
            }
        }
    }
    type = "batch"
}