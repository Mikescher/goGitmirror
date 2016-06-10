package main

import (
	"strings"
)

type GitController struct {
	Folder string
}

func (this *GitController) ExistsLocal() bool {

	exitcode, _, _, err := CmdRun(this.Folder, "git", "status")

	if err != nil {
		EXIT_ERROR("Error executing command 'git status'\n\n"+err.Error(), EXIT_GIT_ERROR)
	}

	return exitcode == 0
}

func (this *GitController) ExecGitCommand(args ...string) string {
	exitcode, stdout, stderr, err := CmdRun(this.Folder, "git", args...)

	if err != nil {
		EXIT_ERROR("Error executing command 'git "+args[0]+"'\n\n"+err.Error(), EXIT_GIT_ERROR)
	}

	if exitcode != 0 {
		EXIT_ERROR("Error in command 'git "+args[0]+"'\n\n"+stderr, EXIT_GIT_ERROR)
	}

	return stdout
}

func (this *GitController) ExecCredGitCommand(cred GGCredentials, args ...string) string {

	if IsEmpty(cred.Host) || IsEmpty(cred.Username) || IsEmpty(cred.Password) {
		return this.ExecGitCommand(args...)
	}

	credRC := "machine " + cred.Host + " login " + cred.Username + " password " + cred.Password

	EnterNetRCBlock(credRC)

	stdout := this.ExecGitCommand(args...)

	ExitNetRCBlock()

	return stdout
}

func (this *GitController) QueryGitCommand(args ...string) string {
	exitcode, stdout, stderr, err := CmdRun(this.Folder, "git", args...)

	if err != nil {
		EXIT_ERROR("Error executing command 'git "+args[0]+"'\n\n"+err.Error(), EXIT_GIT_ERROR)
	}

	if exitcode != 0 {
		EXIT_ERROR("Error in command 'git "+args[0]+"'\n\n"+stderr, EXIT_GIT_ERROR)
	}

	return stdout
}

func (this *GitController) RemoveAllRemotes() {
	branches := this.QueryGitCommand("remote")

	for _, remote := range strings.Split(branches, "\n") {
		if !IsEmpty(remote) {
			this.ExecGitCommand("remote", "rm", remote)
		}
	}
}

func (this *GitController) CloneOrPull(branch string, remote string, cred GGCredentials) {

	if this.ExistsLocal() {
		this.RemoveAllRemotes()
		this.ExecGitCommand("remote", "add", "origin", remote)

		this.ExecCredGitCommand(cred, "fetch", "--all")
		this.ExecGitCommand("checkout", "-f", "origin/"+branch)
		this.ExecCredGitCommand(cred, "fetch", "--all")
		this.ExecGitCommand("reset", "--hard", "origin/"+branch)

	} else {
		this.ExecCredGitCommand(cred, "clone", remote, ".", "--origin", "origin")
		this.ExecGitCommand("checkout", "-f", "origin/"+branch)
	}

	this.ExecCredGitCommand(cred, "branch", "-u", "origin/"+branch, branch)
	this.ExecCredGitCommand(cred, "clean", "-f", "-d")
}

func (this *GitController) PushBack(branch string, remote string, cred GGCredentials, useForce bool) {

	this.RemoveAllRemotes()
	this.ExecGitCommand("remote", "add", "origin", remote)

	this.ExecCredGitCommand(cred, "fetch", "--all")
	this.ExecCredGitCommand(cred, "branch", "-u", "origin/"+branch, branch)
	this.ExecCredGitCommand(cred, "checkout", branch)
	status := this.ExecGitCommand("status")
	LOG_OUT(status)

	var commandoutput string

	if useForce {
		commandoutput = this.ExecCredGitCommand(cred, "push", "origin", "HEAD:"+branch, "--force")
	} else {
		commandoutput = this.ExecCredGitCommand(cred, "push", "origin", "HEAD:"+branch)
	}

	LOG_OUT(commandoutput)
}

func (this *GitController) ListLocalBranches() []string {
	stdout := this.QueryGitCommand("branch", "-a", "--list")
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
