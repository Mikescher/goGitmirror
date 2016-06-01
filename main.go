// gogitmirror project main.go
package main

import (
	"fmt"
	"os"
	"strings"
)

const EXIT_SUCCESS = 0
const EXIT_CONFIG_READ_ERROR = 101
const EXIT_ERRONEOUS_ADD_ARGS = 201
const EXIT_ERROR_INTERNAL = 999

var CONFIG_PATH = "~/.config/gogitmirror.toml"

var PROGNAME = "goGitmirror"
var PROGVERSION = "0.1"

func main() {
	Init()

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

	ExecHelp()
}

func Init() {
	CONFIG_PATH = ExpandPath(CONFIG_PATH)
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
}

func ExecCron(force bool) {
	var config GGMConfig
	config.LoadFromFile(CONFIG_PATH)

	for _, conf := range config.Remote {
		conf.Force = conf.Force || force

		if config.AutoCleanTempFolder {
			conf.CleanFolder()
		}

		conf.Update()

		if config.AutoCleanTempFolder {
			conf.CleanFolder()
		}
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
