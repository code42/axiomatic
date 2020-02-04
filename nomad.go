package main

// jobTemplate is a Nomad job definition is json format with template variables
const jobTemplate = `
{
	"Job": {
		"Datacenters": [
		"dc1"
		],
		"ID": "{{ .Name }}",
		"Name": "{{ .Name }}",
		"Region": "global",
		"TaskGroups": [
		{
			"Name": "dir2consul",
			"Tasks": [
			{
				"Artifacts": [
				{
					"GetterMode": "any",
					"GetterOptions": null,
					"GetterSource": "{{ .GitRepoURL }}",
					"RelativeDest": "local/"
				}
				],
				"Config": {
					"image": "jimrazmus/awscli",
					"args": [
						"aws",
						"--version"
					]
				},
				"Driver": "docker",
				"Env": null,
				"Meta": {
					"commit-SHA": "{{ .HeadSHA }}"
				},
				"Name": "dir2consul",
				"Vault": null
			}
			]
		}
		],
		"Type": "batch",
		"VaultToken": ""
	}
}
`
