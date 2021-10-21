// gogitmirror project main.go
package main

import (
	"bufio"
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

	if strings.ToLower(os.Args[1]) == "status" {
		ExecStatus(ParamIsSet("force"))
		return
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

	if strings.ToLower(os.Args[1]) == "credentials" {
		ExecCredHelper()
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
	fmt.Println("   status")
	fmt.Println("       show status of all configured remotes")
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
		LOG_OUT("   > [Credentials.Source] := " + conf.SourceCredentials.Str())
		LOG_OUT("   > [Credentials.Target] := " + conf.TargetCredentials.Str())

		conf.Force = conf.Force || force

		if config.AutoCleanTempFolder {
			LOG_OUT("Testing temp folder for remote " + conf.Target)
			conf.CleanFolder()
		}

		conf.Update(config)

		if config.AutoCleanTempFolder {
			LOG_OUT("Cleaning temp folder for remote " + conf.Target)
			conf.CleanFolder()
		}

		LOG_LINESEP()
	}
}

func ExecAdd() {
	var config GGMConfig

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

	LOG_OUT("Reading config file")
	LOG_LINESEP()
	config.LoadFromFile(ExpandPath(CONFIG_PATH))

	f, err := os.OpenFile(ExpandPath(CONFIG_PATH), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		EXIT_ERROR("ERROR: Could not open file '"+CONFIG_PATH+"'", EXIT_CONFIG_WRITE)
	}

	defer f.Close()

	var text = "\n\n\n[[Remote]] # Added via commandline\nSource = \"" + source + "\"\nTarget = \"" + target + "\"\n"

	if _, err = f.WriteString(text); err != nil {
		EXIT_ERROR("ERROR: Could not write to file '"+CONFIG_PATH+"'", EXIT_CONFIG_WRITE)
	}
}

func ExecCrypt() {
	if len(os.Args) < 3 {
		EXIT_ERROR("ERROR: The comand [Credentials] needs an password supplied as argument", EXIT_ERRONEOUS_CRYPT_ARGS)
	}

	LOG_OUT("aes:" + Encrypt(os.Args[2]))
}

func ExecStatus(force bool) {
	var config GGMConfig

	LOG_LINESEP()
	config.LoadFromFile(ExpandPath(CONFIG_PATH))

	LOG_OUT(" | " + forceStrLen("NAME", STAT_COL_NAME) + "| " + forceStrLen("BRANCH", STAT_COL_BRANCH) + "| " + forceStrLen("SOURCE", STAT_COL_SOURCE) + " | " + forceStrLen("LOCAL", STAT_COL_LOCAL) + " | " + forceStrLen("TARGET", STAT_COL_TARGET) + "")
	LOG_OUT("-|-" + strings.Repeat("-", STAT_COL_NAME) + "|-" + strings.Repeat("-", STAT_COL_BRANCH) + "|-" + strings.Repeat("-", STAT_COL_SOURCE) + "-|-" + strings.Repeat("-", STAT_COL_LOCAL) + "-|-" + strings.Repeat("-", STAT_COL_TARGET) + "-")

	for _, conf := range config.Remote {
		conf.Force = conf.Force || force
		conf.OutputStatus(config)
	}
}

func ExecCredHelper() {
	if len(os.Args) < 3 {
		EXIT_ERROR("ERROR: The comand [Crypt] needs an cred_index", EXIT_ERRONEOUS_CRYPT_ARGS)
	}

	uniqid := os.Args[2]

	var config GGMConfig

	config.LoadFromFile(ExpandPath(CONFIG_PATH))

	stdin := ""
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		txt := scanner.Text()
		stdin += txt + "\n"
		if txt == "" {
			break
		}
	}

	for _, cred := range config.Credentials {
		if cred.UniqID == uniqid {
			fmt.Print("username=" + cred.Username + "\n" + "password=" + cred.Password + "\n\n")
			return
		}
	}

	EXIT_ERROR("Credential '"+uniqid+"' not found", EXIT_ERRONEOUS_CRED_ARGS)
}
