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

func (this *GitController) ExecGitCommand(args ...string) {
	exitcode, _, stderr, err := CmdRun(this.Folder, "git", args...)

	if err != nil {
		EXIT_ERROR("Error executing command 'git "+args[0]+"'\n\n"+err.Error(), EXIT_GIT_ERROR)
	}

	if exitcode != 0 {
		EXIT_ERROR("Error in command 'git "+args[0]+"'\n\n"+stderr, EXIT_GIT_ERROR)
	}
}

func (this *GitController) ExecCredGitCommand(cred GGCredentials, args ...string) {

	if IsEmpty(cred.Host) || IsEmpty(cred.Username) || IsEmpty(cred.Password) {
		this.ExecGitCommand(args...)
		return
	}

	credRC := "machine " + cred.Host + " login " + cred.Username + " password " + cred.Password

	EnterNetRCBlock(credRC)

	this.ExecGitCommand(args...)

	ExitNetRCBlock()
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
		this.ExecGitCommand("reset", "--hard", "origin/"+branch)

	} else {
		this.ExecCredGitCommand(cred, "clone", remote, ".", "--origin", "origin")
	}
}

func (this *GitController) PushBack(branch string, remote string, cred GGCredentials, useForce bool) {
	this.RemoveAllRemotes()

	if useForce {
		this.ExecCredGitCommand(cred, "push", remote, "HEAD:"+branch, "--force")
	} else {
		this.ExecCredGitCommand(cred, "push", remote, "HEAD:"+branch)
	}
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

		if !IsEmpty(branch) && !strings.EqualFold(branch, "HEAD") {
			result = AppendIfUniqueCaseInsensitive(result, branch)
		}
	}

	return result
}
