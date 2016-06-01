package main

import (
	"net/url"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type GGMConfig struct {
	TemporaryPath       string
	AutoCleanTempFolder bool

	Credentials []GGCredentials

	Remote []GGMirror
}

type GGMirror struct {
	Source string
	Target string

	Force bool

	Branches            []string // If not set AutoBranchDiscovery becomes true
	AutoBranchDiscovery bool     // normally not set via TOML, but auto assigned based on host

	SourceCredentials GGCredentials // normally not set via TOML, but auto assigned based on host
	TargetCredentials GGCredentials // normally not set via TOML, but auto assigned based on host
}

type GGCredentials struct {
	Host string // Host is empty string for anonymous login

	Username string
	Password string
}

func (config GGMConfig) LoadFromFile(path string) {

	if _, err := toml.DecodeFile(path, &config); err != nil {
		os.Stderr.WriteString("ERROR: Cannot load config from " + path + "\n")
		os.Stderr.WriteString("\n")
		os.Stderr.WriteString(err.Error())

		os.Exit(EXIT_CONFIG_READ_ERROR)
	}

	for i := 0; i < len(config.Credentials); i++ {
		if config.Credentials[i].Host == "" {
			os.Stderr.WriteString("ERROR: Credeentials must have the property 'Host' set\n")

			os.Exit(EXIT_CONFIG_READ_ERROR)
		}
	}

	for i := 0; i < len(config.Remote); i++ {
		if config.Remote[i].Source == "" {
			os.Stderr.WriteString("ERROR: Every remote must have the property 'Source' set\n")

			os.Exit(EXIT_CONFIG_READ_ERROR)
		}

		if config.Remote[i].Target == "" {
			os.Stderr.WriteString("ERROR: Every remote must have the property 'Target' set\n")

			os.Exit(EXIT_CONFIG_READ_ERROR)
		}

		if config.Remote[i].Branches == nil {
			config.Remote[i].Branches = []string{} // Default value
			config.Remote[i].AutoBranchDiscovery = true
		} else {
			config.Remote[i].AutoBranchDiscovery = false
		}

		urlSource, err := url.Parse(config.Remote[i].Source)
		if err != nil {
			os.Stderr.WriteString("ERROR: The Source '" + config.Remote[i].Source + "' is not a valid URL\n")

			os.Exit(EXIT_CONFIG_READ_ERROR)
		}

		urlTarget, err := url.Parse(config.Remote[i].Target)
		if err != nil {
			os.Stderr.WriteString("ERROR: The Target '" + config.Remote[i].Target + "' is not a valid URL\n")

			os.Exit(EXIT_CONFIG_READ_ERROR)
		}

		for _, cred := range config.Credentials {
			if strings.ToUpper(cred.Host) == strings.ToUpper(urlSource.Host) && config.Remote[i].SourceCredentials.Host == "" {
				config.Remote[i].SourceCredentials = cred
			}

			if strings.ToUpper(cred.Host) == strings.ToUpper(urlTarget.Host) && config.Remote[i].TargetCredentials.Host == "" {
				config.Remote[i].TargetCredentials = cred
			}
		}
	}
}
