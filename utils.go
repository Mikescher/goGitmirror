package main

import (
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func Contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func ParamIsSet(longArg string) bool {
	for _, s := range os.Args[1:] {
		if s[0:2] == "--" {
			if strings.ToLower(s[2:]) == strings.ToLower(longArg) {
				return true
			}
		}
	}

	return false
}

func ParamIsSet2(longArg string, shortArg string) bool {
	for _, s := range os.Args[1:] {
		if s[0:2] == "--" {
			if strings.ToLower(s[2:]) == strings.ToLower(longArg) {
				return true
			}
		} else if s[0:1] == "-" {
			if strings.ToLower(s[1:]) == strings.ToLower(shortArg) {
				return true
			}
		}
	}

	return false
}

func IsValidURL(uri string) bool {
	_, err := url.Parse(uri)
	if err != nil {
		return false
	}

	return true
}

func ExpandPath(path string) string {
	usr, err := user.Current()
	if err != nil {
		os.Stderr.WriteString("ERROR: Cannot read user home dir\n")

		os.Exit(EXIT_ERROR_INTERNAL)
	}

	if path[:2] == "~/" {
		path = filepath.Join(usr.HomeDir, path[2:])
	}

	return path
}
