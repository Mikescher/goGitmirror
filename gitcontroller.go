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
	exitcode, _, _, err := this.ExecGitCommandErr("status")

	if err != nil {
		EXIT_ERROR("Error executing command 'git status'\n\n"+err.Error(), EXIT_GIT_ERROR)
	}

	return exitcode == 0
}

func (this *GitController) ExecGitCommandSafe(args ...string) (int, string, string) {
	exitcode, stdout, stderr, err := this.ExecGitCommandErr(args...)

	if err != nil {
		exitcode = -1
		stderr = "Recoverable Error executing command 'git " + args[0] + "'\n\n" + err.Error()
		LOG_OUT("Recoverable Internal Error in command 'git " + args[0] + "'\n\n" + stderr)
	} else if exitcode != 0 {
		LOG_OUT("Recoverable Error in command 'git " + args[0] + "'\n\n" + stderr)
	}

	return exitcode, stdout, stderr
}

func (this *GitController) ExecGitCommandErr(args ...string) (int, string, string, error) {
	return CmdRun(this.Folder, this.Silent, "git", args...)
}

func (this *GitController) ExecGitCommand(args ...string) string {
	exitcode, stdout, stderr, err := this.ExecGitCommandErr(args...)

	if err != nil {
		EXIT_ERROR("Error executing command 'git "+args[0]+"'\n\n"+err.Error(), EXIT_GIT_ERROR)
	}

	if exitcode != 0 {
		EXIT_ERROR("Error in command 'git "+args[0]+"'\n\n"+stderr, EXIT_GIT_ERROR)
	}

	return stdout
}

func (this *GitController) ExecCredGitCommandSafe(cred GGCredentials, forceNetRCClean bool, args ...string) (int, string, string) {

	if IsEmpty(cred.Host) || IsEmpty(cred.Username) || IsEmpty(cred.Password) {
		return this.ExecGitCommandSafe(args...)
	}

	EnterNetRCBlock(cred.Host, cred.Username, cred.Password)

	exitcode, stdout, stderr := this.ExecGitCommandSafe(args...)

	ExitNetRCBlock(forceNetRCClean)

	return exitcode, stdout, stderr
}

func (this *GitController) ExecCredGitCommand(cred GGCredentials, forceNetRCClean bool, args ...string) string {

	if IsEmpty(cred.Host) || IsEmpty(cred.Username) || IsEmpty(cred.Password) {
		return this.ExecGitCommand(args...)
	}

	EnterNetRCBlock(cred.Host, cred.Username, cred.Password)

	stdout := this.ExecGitCommand(args...)

	ExitNetRCBlock(forceNetRCClean)

	return stdout
}

func (this *GitController) RemoveAllRemotes() {
	branches := this.ExecGitCommand("remote")

	for _, remote := range strings.Split(branches, "\n") {
		if !IsEmpty(remote) {
			this.ExecGitCommand("remote", "rm", remote)
		}
	}
}

func (this *GitController) CloneOrPull(branch string, remote string, cred GGCredentials, forceNetRCClean bool) {

	if this.ExistsLocal() {
		this.RemoveAllRemotes()
		this.ExecGitCommand("remote", "add", "origin", remote)

		this.ExecCredGitCommand(cred, forceNetRCClean, "fetch", "--all")
		this.ExecGitCommand("checkout", "-f", branch)
		this.ExecGitCommand("reset", "--hard", "origin/"+branch)

	} else {
		this.ExecCredGitCommand(cred, forceNetRCClean, "clone", remote, ".", "--origin", "origin")
		this.ExecGitCommand("checkout", "-f", "origin/"+branch)
	}

	this.ExecCredGitCommand(cred, forceNetRCClean, "branch", "-u", "origin/"+branch, branch)
	this.ExecCredGitCommand(cred, forceNetRCClean, "clean", "-f", "-d")
}

func (this *GitController) PushBack(branch string, remote string, cred GGCredentials, useForce bool, forceFallback bool, forceNetRCClean bool) {

	this.RemoveAllRemotes()
	this.ExecGitCommand("remote", "add", "origin", remote)

	if this.HasRemoteBranch(branch) {
		LOG_OUT("Branch " + branch + " does exist on remote " + remote)
		this.PushBackExistingBranch(branch, remote, cred, useForce, forceFallback, forceNetRCClean)
	} else {
		LOG_OUT("Branch " + branch + " does not exist on remote " + remote)
		this.PushBackNewBranch(branch, remote, cred, useForce, forceFallback, forceNetRCClean)
	}
}

func (this *GitController) PushBackExistingBranch(branch string, remote string, cred GGCredentials, useForce bool, forceFallback bool, forceNetRCClean bool) {

	this.RemoveAllRemotes()
	this.ExecGitCommand("remote", "add", "origin", remote)

	this.ExecCredGitCommand(cred, forceNetRCClean, "fetch", "--all")
	this.ExecCredGitCommand(cred, forceNetRCClean, "branch", "-u", "origin/"+branch, branch)
	this.ExecCredGitCommand(cred, forceNetRCClean, "checkout", branch)
	status := this.ExecGitCommand("status")
	LOG_OUT(status)

	var commandoutput string

	if useForce {
		commandoutput = this.ExecCredGitCommand(cred, forceNetRCClean, "push", "origin", "HEAD:"+branch, "--follow-tags", "--force")
	} else if forceFallback {
		exitcode, stdout, _ := this.ExecCredGitCommandSafe(cred, forceNetRCClean, "push", "origin", "HEAD:"+branch, "--follow-tags")
		commandoutput = stdout
		if exitcode != 0 {
			LOG_OUT("Command in normal mode failed - falling back to force-push")
			commandoutput = this.ExecCredGitCommand(cred, forceNetRCClean, "push", "origin", "HEAD:"+branch, "--follow-tags", "--force")
		}
	} else {
		commandoutput = this.ExecCredGitCommand(cred, forceNetRCClean, "push", "origin", "HEAD:"+branch, "--follow-tags")
	}

	LOG_OUT(commandoutput)
}

func (this *GitController) PushBackNewBranch(branch string, remote string, cred GGCredentials, useForce bool, forceFallback bool, forceNetRCClean bool) {

	this.RemoveAllRemotes()
	this.ExecGitCommand("remote", "add", "origin", remote)

	this.ExecCredGitCommand(cred, forceNetRCClean, "fetch", "--all")
	this.ExecCredGitCommand(cred, forceNetRCClean, "checkout", branch)
	status := this.ExecGitCommand("status")
	LOG_OUT(status)

	var commandoutput string

	if useForce {
		commandoutput = this.ExecCredGitCommand(cred, forceNetRCClean, "push", "origin", "HEAD:"+branch, "--tags", "--force")
	} else if forceFallback {
		exitcode, stdout, _ := this.ExecCredGitCommandSafe(cred, forceNetRCClean, "push", "origin", "HEAD:"+branch, "--tags")
		commandoutput = stdout
		if exitcode != 0 {
			LOG_OUT("Command in normal mode failed - falling back to force-push")
			commandoutput = this.ExecCredGitCommand(cred, forceNetRCClean, "push", "origin", "HEAD:"+branch, "--tags", "--force")
		}
	} else {
		commandoutput = this.ExecCredGitCommand(cred, forceNetRCClean, "push", "origin", "HEAD:"+branch, "--tags")
	}

	LOG_OUT(commandoutput)
}

func (this *GitController) GarbageCollect() {
	this.ExecGitCommand("gc")
}

func (this *GitController) ListLocalBranches() []string {
	stdout := this.ExecGitCommand("branch", "-a", "--list")
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
			branch = line
		}

		branch = strings.TrimSpace(branch)

		if !IsEmpty(branch) && !strings.EqualFold(branch, "HEAD") && branch[0:1] != "(" {
			result = AppendIfUniqueCaseInsensitive(result, branch)
		}
	}

	return result
}

func (this *GitController) HasRemoteBranch(branchname string) bool {
	stdout := this.ExecGitCommand("branch", "-r", "--list")
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
