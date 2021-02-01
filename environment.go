package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

func setupEnvironment() {
	envDefaults := map[string]string{
		"GITHUB_SECRET": "",
		"IP":            "127.0.0.1",
		"PORT":          "8181",
		"SSH_PRIV_KEY":  "",
		"SSH_PUB_KEY":   "",
	}

	viper.SetEnvPrefix("AXIOMATIC")

	for key, val := range envDefaults {
		viper.SetDefault(key, val)
	}

	viper.AutomaticEnv()

	for key := range envDefaults {
		err := viper.BindEnv(key)
		if err != nil {
			log.Fatalf("Error setting up environment: %s", err)
		}
	}
}

// filterEnvironment returns a map of strings from ss that begin with "CONSUL_" or "D2C_"
func filterEnvironment(ss []string) (map[string]string, error) {
	r := make(map[string]string)

	for _, s := range ss {
		if strings.HasPrefix(s, "CONSUL_") || strings.HasPrefix(s, "D2C_") {
			kv := strings.Split(s, "=")

			if len(kv) != 2 {
				return nil, fmt.Errorf("Error parsing environment variable: '%s'", s)
			}

			r[kv[0]] = kv[1]
		}
	}

	return r, nil
}

func isMissingConfiguration() bool {
	r := false
	vs := []string{"GITHUB_SECRET", "SSH_PRIV_KEY", "SSH_PUB_KEY"}
	for _, v := range vs {
		if viper.GetString(v) == "" {
			log.Printf("You must configure AXIOMATIC_%s!", v)
			r = true
		}
	}
	return r
}
