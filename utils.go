package main

import (
	"bytes"
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
		EXIT_ERROR("ERROR: Cannot read user home dir", EXIT_ERROR_INTERNAL)
	}

	if path[:2] == "~/" {
		path = filepath.Join(usr.HomeDir, path[2:])
	}

	return path
}

func InRange(needle rune, start rune, end rune) bool {
	return needle >= start && needle <= end
}

func NormalizeStringToFilePath(str string) string {
	var buffer bytes.Buffer

	for _, c := range str {
		if InRange(c, '0', '9') || InRange(c, 'A', 'Z') || InRange(c, 'a', 'z') || c == '-' || c == '.' {
			buffer.WriteRune(c)
		}
	}

	var fragment string
	fragment = buffer.String()
	fragment = strings.ToLower(fragment)
	fragment = strings.TrimSpace(fragment)

	for strings.ContainsRune(fragment, '.') {
		fragment = fragment[:strings.IndexRune(fragment, '.')]
	}

	return fragment
}

func EXIT_ERROR(msg string, code int) {
	os.Stderr.WriteString(msg + "\n")

	os.Exit(code)
}
