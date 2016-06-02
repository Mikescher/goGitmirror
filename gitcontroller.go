package main

import (
	"strings"
)

type GitController struct {
	Folder string
	Remote string
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
		if strings.TrimSpace(remote) != "" {
			this.ExecGitCommand("remote", "rm", remote)
		}
	}
}

func (this *GitController) CloneOrPull(branch string) {

	if this.ExistsLocal() {
		this.RemoveAllRemotes()
		this.ExecGitCommand("remote", "add", "origin", this.Remote)

		this.ExecGitCommand("fetch", "--all")
		this.ExecGitCommand("reset", "--hard", "origin/"+branch)

	} else {
		LOG_OUT("Cloning fresh remote " + this.Remote + " to " + this.Folder)
		this.ExecGitCommand("clone", this.Remote, ".", "--origin", "origin")
		this.ExecGitCommand("pull", this.Remote)
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

		if strings.HasPrefix(strings.ToLower(line), "remotes/origin/") {
			result = AppendIfUniqueCaseInsensitive(result, line[15:])
		} else if line != "" {
			result = AppendIfUniqueCaseInsensitive(result, line)
		}
	}

	return result
}
