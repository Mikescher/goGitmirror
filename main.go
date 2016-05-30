// gogitmirror project main.go
package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/davecgh/go-spew/spew"

	"os"
)

const EXIT_SUCCESS = 0
const EXIT_CONFIG_READ_ERROR = 101
const EXIT_ERRONEOUS_ADD_ARGS = 201

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

	Branches []string // If not set, this is []{"master"}

	SourceCredentials GGCredentials // normally not set via TOML, but auto assigned based on host
	TargetCredentials GGCredentials // normally not set via TOML, but auto assigned based on host
}

type GGCredentials struct {
	Host string // Host is empty string for anonymous login

	Username string
	Password string
}

func main() {
	if len(os.Args) < 2 || Contains(os.Args[1:], "--help") {
		ExecHelp()
		return
	}

	if strings.ToLower(os.Args[1]) == "cron" {
		ExecCron()
		return
	}

	if strings.ToLower(os.Args[1]) == "add" && len(os.Args) >= 4 {
		ExecAdd()
		return
	}

	ExecHelp()
}

func ExecHelp() {
	fmt.Println("TODO: HELP")
}

func ExecCron() {
	var config GGMConfig = LoadConfig()

	spew.Dump(config)

}

func ExecAdd() {
	var source = os.Args[2]
	var target = os.Args[3]

	_, err := url.Parse(source)
	if err != nil {
		os.Stderr.WriteString("ERROR: The Source '" + source + "' is not a valid URL")

		os.Exit(EXIT_ERRONEOUS_ADD_ARGS)
	}

	_, err = url.Parse(target)
	if err != nil {
		os.Stderr.WriteString("ERROR: The Target '" + target + "' is not a valid URL")

		os.Exit(EXIT_ERRONEOUS_ADD_ARGS)
	}
}

func Contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func LoadConfig() GGMConfig {
	var config GGMConfig

	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		os.Stderr.WriteString("ERROR: Cannot load config.toml")
		os.Stderr.WriteString("")
		os.Stderr.WriteString(err.Error())

		os.Exit(EXIT_CONFIG_READ_ERROR)
	}

	for i := 0; i < len(config.Credentials); i++ {
		if config.Credentials[i].Host == "" {
			os.Stderr.WriteString("ERROR: Credeentials must have the property 'Host' set ")

			os.Exit(EXIT_CONFIG_READ_ERROR)
		}
	}

	for i := 0; i < len(config.Remote); i++ {
		if config.Remote[i].Source == "" {
			os.Stderr.WriteString("ERROR: Every remote must have the property 'Source' set")

			os.Exit(EXIT_CONFIG_READ_ERROR)
		}

		if config.Remote[i].Target == "" {
			os.Stderr.WriteString("ERROR: Every remote must have the property 'Target' set")

			os.Exit(EXIT_CONFIG_READ_ERROR)
		}

		if config.Remote[i].Branches == nil {
			config.Remote[i].Branches = []string{"master"} // Default value
		}

		urlSource, err := url.Parse(config.Remote[i].Source)
		if err != nil {
			os.Stderr.WriteString("ERROR: The Source '" + config.Remote[i].Source + "' is not a valid URL")

			os.Exit(EXIT_CONFIG_READ_ERROR)
		}

		urlTarget, err := url.Parse(config.Remote[i].Target)
		if err != nil {
			os.Stderr.WriteString("ERROR: The Target '" + config.Remote[i].Target + "' is not a valid URL")

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

	return config
}