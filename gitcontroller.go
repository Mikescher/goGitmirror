package main

import (
	"strings"
)

type GitController struct {
	Folder string
	Silent bool
}

func (this *GitController) SetSilent() {
	this.Silent = true
}

func (this *GitController) ExistsLocal() bool {

	if !PathExists(this.Folder) {
		return false
	}
	exitcode, _, _, err := this.ExecGitCommandErr(false, "status")

	if err != nil {
		EXIT_ERROR("Error executing command 'git status'\n\n"+err.Error(), EXIT_GIT_ERROR)
	}

	return exitcode == 0
}

func (this *GitController) ExecGitCommandSafe(nosslverify bool, args ...string) (int, string, string) {
	exitcode, stdout, stderr, err := this.ExecGitCommandErr(nosslverify, args...)

	if err != nil {
		exitcode = -1
		stderr = "Recoverable Error executing command 'git " + args[0] + "'\n\n" + err.Error()
		LOG_OUT("Recoverable Internal Error in command 'git " + args[0] + "'\n\n" + stderr)
	} else if exitcode != 0 {
		LOG_OUT("Recoverable Error in command 'git " + args[0] + "'\n\n" + stderr)
	}

	return exitcode, stdout, stderr
}

func (this *GitController) ExecGitCommandErr(nosslverify bool, args ...string) (int, string, string, error) {
	allargs := args
	if nosslverify {
		allargs = append([]string{"-c", "http.sslVerify=false"}, allargs...)
	}
	return CmdRun(this.Folder, this.Silent, "git", allargs...)
}

func (this *GitController) ExecGitCommand(nosslverify bool, args ...string) string {
	exitcode, stdout, stderr, err := this.ExecGitCommandErr(nosslverify, args...)

	if err != nil {
		EXIT_ERROR("Error executing command 'git "+args[0]+"'\n\n"+err.Error(), EXIT_GIT_ERROR)
	}

	if exitcode != 0 {
		EXIT_ERROR("Error in command 'git "+args[0]+"'\n\n"+stderr, EXIT_GIT_ERROR)
	}

	return stdout
}

func (this *GitController) ExecCredGitCommandSafe(cred GGCredentials, mode CredMode, forceNetRCClean bool, nosslverify bool, args ...string) (int, string, string) {

	if IsEmpty(cred.Host) || IsEmpty(cred.Username) || IsEmpty(cred.Password) {
		return this.ExecGitCommandSafe(nosslverify, args...)
	}

	if mode == CredModeNetRC {
		EnterNetRCBlock(cred.Host, cred.Username, cred.Password)
		exitcode, stdout, stderr := this.ExecGitCommandSafe(nosslverify, args...)
		ExitNetRCBlock(forceNetRCClean)
		return exitcode, stdout, stderr
	} else if mode == CredModeHelper {
		gitargs := []string{"-c", "credential.helper=\"!" + BINARY_PATH + " credentials " + cred.UniqID + "\""}
		gitargs = append(gitargs, args...)
		exitcode, stdout, stderr := this.ExecGitCommandSafe(nosslverify, gitargs...)
		return exitcode, stdout, stderr
	} else if mode == CredModeCFile {
		tf, cleanup := CreateCredTempFile(cred.Host, cred.Username, cred.Password)
		defer cleanup()
		gitargs := []string{"-c", "credential.helper=store --file " + tf}
		gitargs = append(gitargs, args...)
		exitcode, stdout, stderr := this.ExecGitCommandSafe(nosslverify, gitargs...)
		return exitcode, stdout, stderr
	}

	EXIT_ERROR("Invalid CredMode: "+string(mode), EXIT_CONFIG_VALUE_ERROR)
	return 0, "", ""
}

func (this *GitController) ExecCredGitCommand(cred GGCredentials, mode CredMode, forceNetRCClean bool, nosslverify bool, args ...string) string {

	if IsEmpty(cred.Host) || IsEmpty(cred.Username) || IsEmpty(cred.Password) {
		return this.ExecGitCommand(nosslverify, args...)
	}

	if mode == CredModeNetRC {
		EnterNetRCBlock(cred.Host, cred.Username, cred.Password)
		stdout := this.ExecGitCommand(nosslverify, args...)
		ExitNetRCBlock(forceNetRCClean)
		return stdout
	} else if mode == CredModeHelper {
		gitargs := []string{"-c", "credential.helper=\"!" + BINARY_PATH + " credentials " + cred.UniqID + "\""}
		gitargs = append(gitargs, args...)
		stdout := this.ExecGitCommand(nosslverify, gitargs...)
		return stdout
	} else if mode == CredModeCFile {
		tf, cleanup := CreateCredTempFile(cred.Host, cred.Username, cred.Password)
		defer cleanup()
		gitargs := []string{"-c", "credential.helper=store --file " + tf}
		gitargs = append(gitargs, args...)
		stdout := this.ExecGitCommand(nosslverify, gitargs...)
		return stdout
	}

	EXIT_ERROR("Invalid CredMode: "+string(mode), EXIT_CONFIG_VALUE_ERROR)
	return ""
}

func (this *GitController) RemoveRemoteIfExists(name string) {
	branches := this.ExecGitCommand(false, "remote")

	for _, remote := range strings.Split(branches, "\n") {
		if !IsEmpty(remote) && (remote == name || name == "") {
			this.ExecGitCommand(false, "remote", "rm", remote)
		}
	}
}

func (this *GitController) CloneOrPull(branch string, remote string, cred GGCredentials, credmode CredMode, forceNetRCClean bool) {

	if this.ExistsLocal() {
		this.RemoveRemoteIfExists("origin")
		this.ExecGitCommand(false, "remote", "add", "origin", remote)

		this.ExecCredGitCommand(cred, credmode, forceNetRCClean, cred.NoSSLVerify, "fetch", "--all")

		this.ExecGitCommand(false, "checkout", "--force", "-B", branch)
		this.ExecGitCommand(false, "reset", "--hard", "origin/"+branch)

	} else {
		this.ExecCredGitCommand(cred, credmode, forceNetRCClean, cred.NoSSLVerify, "clone", remote, ".", "--origin", "origin")

		this.ExecCredGitCommand(cred, credmode, forceNetRCClean, cred.NoSSLVerify, "fetch", "--all")

		this.ExecGitCommand(false, "checkout", "--force", "origin/"+branch)
	}

	this.ExecCredGitCommand(cred, credmode, forceNetRCClean, false, "branch", "--set-upstream-to=origin/"+branch, branch)
	this.ExecCredGitCommand(cred, credmode, forceNetRCClean, false, "clean", "--force", "-d")
}

func (this *GitController) FetchAltRemote(name string, remote string, cred GGCredentials, credmode CredMode, forceNetRCClean bool) {
	this.RemoveRemoteIfExists(name)
	this.ExecCredGitCommand(cred, credmode, forceNetRCClean, false, "remote", "add", name, remote)
	this.ExecCredGitCommand(cred, credmode, forceNetRCClean, false, "fetch", name, "--prune", "--prune-tags", "--tags")
}

func (this *GitController) GetHeadHash(originName string, branch string, hashlen int) string {

	exitcode, stdout, _ := this.ExecGitCommandSafe(false, "show-ref", originName+"/"+branch)

	if exitcode == 0 {
		hash := strings.TrimSpace(stdout)
		if len(hash) <= hashlen {
			return hash
		} else {
			return stdout[:hashlen]
		}
	} else {
		return ""
	}
}

func (this *GitController) PushBack(branch string, remote string, cred GGCredentials, credmode CredMode, useForce bool, forceFallback bool, forceNetRCClean bool) {

	this.RemoveRemoteIfExists("origin")
	this.ExecGitCommand(false, "remote", "add", "origin", remote)

	if this.HasRemoteBranch(branch, cred.NoSSLVerify) {
		LOG_OUT("Branch " + branch + " does exist on remote " + remote)
		this.PushBackExistingBranch(branch, remote, cred, credmode, useForce, forceFallback, forceNetRCClean, cred.NoSSLVerify)
	} else {
		LOG_OUT("Branch " + branch + " does not exist on remote " + remote)
		this.PushBackNewBranch(branch, remote, cred, credmode, useForce, forceFallback, forceNetRCClean, cred.NoSSLVerify)
	}
}

func (this *GitController) PushBackExistingBranch(branch string, remote string, cred GGCredentials, credmode CredMode, useForce bool, forceFallback bool, forceNetRCClean bool, nosslverify bool) {

	this.RemoveRemoteIfExists("origin")
	this.ExecGitCommand(false, "remote", "add", "origin", remote)

	this.ExecCredGitCommand(cred, credmode, forceNetRCClean, nosslverify, "fetch", "--all")
	this.ExecCredGitCommand(cred, credmode, forceNetRCClean, false, "branch", "--set-upstream-to=origin/"+branch, branch)
	this.ExecCredGitCommand(cred, credmode, forceNetRCClean, false, "checkout", branch)
	status := this.ExecGitCommand(false, "status")
	LOG_OUT(status)

	var commandoutput string

	if useForce {
		commandoutput = this.ExecCredGitCommand(cred, credmode, forceNetRCClean, nosslverify, "push", "origin", "HEAD:"+branch, "--follow-tags", "--force")
	} else if forceFallback {
		exitcode, stdout, _ := this.ExecCredGitCommandSafe(cred, credmode, forceNetRCClean, nosslverify, "push", "origin", "HEAD:"+branch, "--follow-tags")
		commandoutput = stdout
		if exitcode != 0 {
			LOG_OUT("Command in normal mode failed - falling back to force-push")
			commandoutput = this.ExecCredGitCommand(cred, credmode, forceNetRCClean, nosslverify, "push", "origin", "HEAD:"+branch, "--follow-tags", "--force")
		}
	} else {
		commandoutput = this.ExecCredGitCommand(cred, credmode, forceNetRCClean, nosslverify, "push", "origin", "HEAD:"+branch, "--follow-tags")
	}

	LOG_OUT(commandoutput)
}

func (this *GitController) PushBackNewBranch(branch string, remote string, cred GGCredentials, credmode CredMode, useForce bool, forceFallback bool, forceNetRCClean bool, nosslverify bool) {

	this.RemoveRemoteIfExists("origin")
	this.ExecGitCommand(false, "remote", "add", "origin", remote)

	this.ExecCredGitCommand(cred, credmode, forceNetRCClean, nosslverify, "fetch", "--all")
	this.ExecCredGitCommand(cred, credmode, forceNetRCClean, false, "checkout", branch)
	status := this.ExecGitCommand(false, "status")
	LOG_OUT(status)

	var commandoutput string

	if useForce {
		commandoutput = this.ExecCredGitCommand(cred, credmode, forceNetRCClean, nosslverify, "push", "origin", "HEAD:"+branch, "--tags", "--force")
	} else if forceFallback {
		exitcode, stdout, _ := this.ExecCredGitCommandSafe(cred, credmode, forceNetRCClean, nosslverify, "push", "origin", "HEAD:"+branch, "--tags")
		commandoutput = stdout
		if exitcode != 0 {
			LOG_OUT("Command in normal mode failed - falling back to force-push")
			commandoutput = this.ExecCredGitCommand(cred, credmode, forceNetRCClean, nosslverify, "push", "origin", "HEAD:"+branch, "--tags", "--force")
		}
	} else {
		commandoutput = this.ExecCredGitCommand(cred, credmode, forceNetRCClean, nosslverify, "push", "origin", "HEAD:"+branch, "--tags")
	}

	LOG_OUT(commandoutput)
}

func (this *GitController) GarbageCollect() {
	this.ExecGitCommand(false, "gc")
}

func (this *GitController) ListLocalBranches() []string {
	stdout := this.ExecGitCommand(false, "branch", "--all", "--list")
	lines := strings.Split(stdout, "\n")

	result := make([]string, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, "*")
		line = strings.TrimSpace(line)

		branch := ""

		if strings.Contains(line, " -> ") {
			line = line[:strings.Index(line, " -> ")]
		}

		if strings.HasPrefix(strings.ToLower(line), "remotes/origin/") {
			branch = line[15:]
		} else {
			//branch = line
			branch = ""
		}

		branch = strings.TrimSpace(branch)

		if !IsEmpty(branch) && !strings.EqualFold(branch, "HEAD") && branch[0:1] != "(" {
			result = AppendIfUniqueCaseInsensitive(result, branch)
		}
	}

	return result
}

func (this *GitController) HasRemoteBranch(branchname string, nosslverify bool) bool {
	stdout := this.ExecGitCommand(nosslverify, "branch", "--remotes", "--list")
	lines := strings.Split(stdout, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, "*")
		line = strings.TrimSpace(line)

		branch := ""

		if strings.Contains(line, " -> ") {
			line = line[:strings.Index(line, " -> ")]
		}

		if strings.HasPrefix(strings.ToLower(line), "remotes/origin/") {
			branch = line[15:]
		} else if strings.HasPrefix(strings.ToLower(line), "origin/") {
			branch = line[7:]
		} else {
			branch = line
		}

		branch = strings.TrimSpace(branch)

		if strings.EqualFold(branch, branchname) {
			return true
		}
	}

	return false
}
