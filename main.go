package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/google/go-github/github"
	"github.com/spf13/viper"
)

var jobTemplate *template.Template

func main() {
	setupEnvironment()
	if isMissingConfiguration() {
		log.Fatal("Shutting down.")
	}
	fmt.Println(startupMessage())

	jobTemplate = template.Must(template.New("job").Parse(templateNomadJob()))

	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/webhook", handleWebhook)

	log.Fatal(http.ListenAndServe(viper.GetString("IP")+":"+viper.GetString("PORT"), nil))
}

func startupMessage() string {
	banner := "\n------------\n Axiomatic \n------------\n"

	config := fmt.Sprintf("Configuration\n\tAXIOMATIC_IP: %s\n\tAXIOMATIC_PORT: %s", viper.GetString("IP"), viper.GetString("PORT"))

	env := os.Environ()
	sort.Strings(env)
	environment := fmt.Sprintf("\nEnvironment\n\t%s", strings.Join(env, "\n\t"))

	return banner + config + environment
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "Good to Serve")
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Good to Serve")
	}
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := r.Body.Close()
		if err != nil {
			log.Printf("Error closing body: %s", err)
		}
	}()

	payload, err := github.ValidatePayload(r, []byte(viper.GetString("GITHUB_SECRET")))
	if err != nil {
		log.Printf("error validating request body: err=%s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		log.Printf("could not parse webhook: err=%s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch e := event.(type) {
	case *github.PingEvent:
		log.Println("GitHub Pinged the Webhook")
	case *github.PushEvent:
		env, err := filterEnvironment(os.Environ())
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		jobArgs := NomadJobData{
			GitRepoName: e.Repo.GetName(),
			GitRepoURL:  e.Repo.GetSSHURL(),
			HeadSHA:     e.GetAfter(),
			DeployKey:   viper.GetString("SSH_PRIV_KEY"),
			Environment: env,
		}

		job, err := templateToJob(jobArgs)
		if err != nil {
			log.Println("template to job failed:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = submitNomadJob(job)
		if err != nil {
			log.Println("submit job failed:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		log.Printf("WARN: unknown event type %s\n", github.WebHookType(r))
		return
	}
}
