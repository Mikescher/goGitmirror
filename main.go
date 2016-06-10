// gogitmirror project main.go
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {

	if len(os.Args) < 2 || ParamIsSet("help") {
		ExecHelp()
		return
	}

	if ParamIsSet2("version", "v") {
		ExecVersion()
	}

	if strings.ToLower(os.Args[1]) == "cron" {
		ExecCron(ParamIsSet("force"))
		return
	}

	if strings.ToLower(os.Args[1]) == "add" {
		ExecAdd()
		return
	}

	if strings.ToLower(os.Args[1]) == "crypt" {
		ExecCrypt()
		return
	}

	ExecHelp()
}

func ExecVersion() {
	fmt.Println(PROGNAME + " " + PROGVERSION)
}

func ExecHelp() {
	fmt.Println("usage: gogitmirror [--version] [--help] <command> [<args>]")
	fmt.Println("")
	fmt.Println("These are the possible commands:")
	fmt.Println("")
	fmt.Println("   add $source $target [--force]")
	fmt.Println("       add a new source-target pair to the configuration")
	fmt.Println("")
	fmt.Println("   cron [--force]")
	fmt.Println("       update all targets, optionally specify --force to")
	fmt.Println("       force push all remotes")
	fmt.Println("")
	fmt.Println("   cyrpt $password")
	fmt.Println("       encrypt an password for use in config file")
}

func ExecCron(force bool) {
	var config GGMConfig

	LOG_OUT("Reading config file")
	LOG_LINESEP()
	config.LoadFromFile(ExpandPath(CONFIG_PATH))

	for _, conf := range config.Remote {
		LOG_OUT("Processing remote " + conf.Target)

		conf.Force = conf.Force || force

		if config.AutoCleanTempFolder {
			LOG_OUT("Testing temp folder for remote " + conf.Target)
			conf.CleanFolder()
		}

		conf.Update()

		if config.AutoCleanTempFolder {
			LOG_OUT("Cleaning temp folder for remote " + conf.Target)
			conf.CleanFolder()
		}

		LOG_LINESEP()
	}
}

func ExecAdd() {
	if len(os.Args) < 4 {
		EXIT_ERROR("ERROR: The comand [Add] needs at least two arguments (source & target)", EXIT_ERRONEOUS_ADD_ARGS)
	}

	var source = os.Args[2]
	var target = os.Args[3]

	if !IsValidURL(source) {
		EXIT_ERROR("ERROR: The Source '"+source+"' is not a valid URL", EXIT_ERRONEOUS_ADD_ARGS)
	}

	if !IsValidURL(target) {
		EXIT_ERROR("ERROR: The Target '"+target+"' is not a valid URL", EXIT_ERRONEOUS_ADD_ARGS)
	}
}

func ExecCrypt() {
	if len(os.Args) < 3 {
		EXIT_ERROR("ERROR: The comand [Crypt] needs an password supplied as argument", EXIT_ERRONEOUS_CRYPT_ARGS)
	}

	LOG_OUT("aes:" + Encrypt(os.Args[2]))
}
