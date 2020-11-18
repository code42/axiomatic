package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"text/template"

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
