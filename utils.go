package main

import (
	"bytes"
	"errors"
	"strings"
	"syscall"

	"net/url"

	"os"
	"os/exec"
	"os/user"

	"path/filepath"

	"io/ioutil"
)

var netRCBlock bool
var netRCBackup []byte

func EnterNetRCBlock(host string, usr string, pass string) {
	content := "machine " + host + "\nlogin " + usr + "\npassword " + pass
	//content := "default " + " login " + usr + " password " + pass

	netRC_read := true
	oldNetRC, err := ioutil.ReadFile(ExpandPath(NETRCPATH))
	if err != nil {
		netRC_read = false
	}

	err = ioutil.WriteFile(ExpandPath(NETRCPATH), []byte(content), 0600)
	if err != nil {
		EXIT_ERROR("Cannot write to "+ExpandPath(NETRCPATH), EXIT_GIT_ERROR)
	}

	if netRC_read && len(oldNetRC) > 0 && !bytes.Equal([]byte(content), oldNetRC) {
		netRCBackup = oldNetRC
		netRCBlock = true
	} else {
		netRCBackup = nil
		netRCBlock = true
	}
}

func ExitNetRCBlock(forceClean bool) {
	if !netRCBlock {
		return
	}

	if netRCBackup == nil && !forceClean {
		ioutil.WriteFile(ExpandPath(NETRCPATH), netRCBackup, 0600)
		netRCBlock = false
	} else {
		os.Remove(ExpandPath(NETRCPATH))
		netRCBlock = false
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

	if strings.ContainsRune(fragment, '.') {
		fragment = fragment[:strings.LastIndex(fragment, ".")]
	}

	return fragment
}

func EXIT_ERROR(msg string, code int) {
	os.Stderr.WriteString(msg + "\n")

	ExitNetRCBlock(false)

	os.Exit(code)
}

func LOG_OUT(msg string) {
	os.Stdout.WriteString(msg + "\n")
}

func LOG_LINESEP() {
	os.Stdout.WriteString("\n")
}

func PathIsValid(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return true
	}

	return false
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}

	EXIT_ERROR("Internal Error: Invalid folder :"+path, EXIT_ERROR_INTERNAL)
	return false
}

func CleanFolder(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func CmdRun(folder string, command string, args ...string) (int, string, string, error) {

	//IF DEBUG
	LOG_OUT("   > " + command + " " + Join(" ", args))

	wout := new(bytes.Buffer)
	werr := new(bytes.Buffer)

	binary, err := exec.LookPath(command)

	if err != nil {
		return 0, "", "", errors.New("The binary '" + command + "' was not found on this system")
	}

	cmd := exec.Command(binary, args...)

	cmd.Dir = folder
	cmd.Stdout = wout
	cmd.Stderr = werr

	if err := cmd.Start(); err != nil {
		return 0, "", "", err
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return status.ExitStatus(), wout.String(), werr.String(), nil
			}
		} else {
			return 0, "", "", err
		}
	}

	//IF DEBUG
	//for _, s := range strings.Split(wout.String(), "\n") {
	//	LOG_OUT("      :   " + s)
	//}

	return 0, wout.String(), werr.String(), nil
}

func Join(sep string, arr []string) string {
	var buffer bytes.Buffer

	var first = true
	for _, piece := range arr {
		if first {
			first = false
			buffer.WriteString(piece)
		} else {
			buffer.WriteString(sep)
			buffer.WriteString(piece)
		}
	}

	return buffer.String()
}

func AppendIfUniqueCaseInsensitive(slice []string, i string) []string {
	for _, ele := range slice {
		if strings.EqualFold(ele, i) {
			return slice
		}
	}
	return append(slice, i)
}

func IsEmpty(value string) bool {
	return strings.TrimSpace(value) == ""
}
