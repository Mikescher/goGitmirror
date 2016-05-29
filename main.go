// gogitmirror project main.go
package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type GGMConfig struct {
	TemporaryPath string
	Remote        []GGMirror
}

type GGMirror struct {
	Source string
	Target string

	Branches []string
}

func main() {
	var config GGMConfig
	if _, err := toml.DecodeFile("config.toml", &config); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("GoGitMirror path: " + config.TemporaryPath)

	for _, cfgRemote := range config.Remote {
		fmt.Println(cfgRemote.Source + "  ->  " + cfgRemote.Target)

		for _, cfgBranch := range cfgRemote.Branches {
			fmt.Println("    | " + cfgBranch)
		}

		fmt.Println("")
	}
}
