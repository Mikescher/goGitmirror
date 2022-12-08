package main

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type GGMConfig struct {
	TemporaryPath       string
	AutoCleanTempFolder bool
	AutoForceFallback   bool
	AlwaysCleanNetRC    bool
	CredentialMode      CredMode

	Credentials []GGCredentials

	Remote []GGMirror

	AutoMirror []GGAutoMirror
}

type CredMode string

const (
	CredModeNetRC  CredMode = "NETRC"
	CredModeHelper CredMode = "CREDHELPER"
	CredModeCFile  CredMode = "CREDFILE"
)

type GGMirror struct {
	ID string

	Source string
	Target string

	Force bool

	PrimaryBranch string // Used for AutoBranchDiscovery (default == master)

	Branches            []string // If not set AutoBranchDiscovery becomes true
	AutoBranchDiscovery bool     // normally not set via TOML, but auto assigned based on host

	SourceCredentials GGCredentials // normally not set via TOML, but auto assigned based on host
	TargetCredentials GGCredentials // normally not set via TOML, but auto assigned based on host

	SourceCredentialsID string // if set, use these credentials (by-id)
	TargetCredentialsID string // if set, use these credentials (by-id)

	TempBaseFolder string // normally not set via TOML, but auto assigned from root config
}

type GGAutoMirror struct {
	Source GGAutoMirrorConfig
	Target GGAutoMirrorConfig

	OnlyMasterBranch bool
}

type GGAutoMirrorConfig struct {
	Type string // [Github, Gitea, Gitlab, Bitbucket]

	RootURL  string // e.g. https://gilab.com
	Username string

	Credentials GGCredentials // normally not set via TOML, but auto assigned based on host
}

type GGCredentials struct {
	ID string // if set credentials are not auto-assigned but can be directly referenced

	Host string // Host is empty string for anonymous login

	Username string
	Password string

	NoSSLVerify bool   // default = false
	UniqID      string // set by code
}

func (this GGCredentials) Str() string {
	if this.ID != "" {
		return "{{" + this.ID + "}}"
	} else {
		return "[" + this.Username + "@" + this.Host + "]"
	}
}

func (this *GGMConfig) LoadFromFile(path string) {

	if _, err := toml.DecodeFile(path, &this); err != nil {
		EXIT_ERROR("ERROR: Cannot load config from "+path+"\n\n"+err.Error(), EXIT_CONFIG_READ_ERROR)
	}

	if this.CredentialMode == "" {
		this.CredentialMode = CredModeCFile
	}

	for i := 0; i < len(this.Credentials); i++ {
		if this.Credentials[i].Host == "" {
			EXIT_ERROR("ERROR: Credentials must have the property 'Host' set", EXIT_CONFIG_READ_ERROR)
		}

		if len(this.Credentials[i].Password) > 5 && this.Credentials[i].Password[:4] == "aes:" {
			this.Credentials[i].Password = Decrypt(this.Credentials[i].Password[4:])
		} else if this.Credentials[i].Password != "" {
			LOG_OUT("WARNING: password for host " + this.Credentials[i].Host + " is unencrypted")
		}

		this.Credentials[i].UniqID = "GGMCRED_" + strconv.Itoa(i)
	}

	for i := 0; i < len(this.Remote); i++ {
		if this.Remote[i].Source == "" {
			EXIT_ERROR("ERROR: Every remote must have the property 'Source' set", EXIT_CONFIG_READ_ERROR)
		}

		if this.Remote[i].Target == "" {
			EXIT_ERROR("ERROR: Every remote must have the property 'Target' set", EXIT_CONFIG_READ_ERROR)
		}

		if this.Remote[i].TempBaseFolder == "" {
			this.Remote[i].TempBaseFolder = this.TemporaryPath
		}

		if this.Remote[i].Branches == nil {
			this.Remote[i].Branches = []string{} // Default value
			this.Remote[i].AutoBranchDiscovery = true
		} else {
			this.Remote[i].AutoBranchDiscovery = false
		}

		if this.Remote[i].PrimaryBranch == "" {
			this.Remote[i].PrimaryBranch = "master"
		}

		urlSource, err := url.Parse(this.Remote[i].Source)
		if err != nil {
			EXIT_ERROR("ERROR: The Source '"+this.Remote[i].Source+"' is not a valid URL", EXIT_CONFIG_READ_ERROR)
		}

		urlTarget, err := url.Parse(this.Remote[i].Target)
		if err != nil {
			EXIT_ERROR("ERROR: The Target '"+this.Remote[i].Target+"' is not a valid URL", EXIT_CONFIG_READ_ERROR)
		}

		if this.Remote[i].SourceCredentials.Host == "" {
			for _, cred := range this.Credentials {
				if cred.ID == "" && strings.ToUpper(cred.Host) == strings.ToUpper(urlSource.Host) {
					this.Remote[i].SourceCredentials = cred
					this.Remote[i].SourceCredentials.Host = urlSource.Host
				}
			}
		}
		if this.Remote[i].TargetCredentials.Host == "" {
			for _, cred := range this.Credentials {
				if cred.ID == "" && strings.ToUpper(cred.Host) == strings.ToUpper(urlTarget.Host) {
					this.Remote[i].TargetCredentials = cred
					this.Remote[i].TargetCredentials.Host = urlTarget.Host
				}
			}
		}

		if this.Remote[i].SourceCredentialsID != "" {
			this.Remote[i].SourceCredentials = GGCredentials{}
			for _, cred := range this.Credentials {
				if cred.ID == this.Remote[i].SourceCredentialsID {
					this.Remote[i].SourceCredentials = cred
					this.Remote[i].SourceCredentials.Host = urlSource.Host
				}
			}
		}
		if this.Remote[i].TargetCredentialsID != "" {
			this.Remote[i].TargetCredentials = GGCredentials{}
			for _, cred := range this.Credentials {
				if cred.ID == this.Remote[i].TargetCredentialsID {
					this.Remote[i].TargetCredentials = cred
					this.Remote[i].TargetCredentials.Host = urlTarget.Host
				}
			}
		}
	}
}

func (this GGMirror) GetTargetFolder() string {
	var buffer bytes.Buffer

	url, _ := url.Parse(this.Target)

	buffer.WriteString(NormalizeStringToFilePath(url.Host))

	for _, shatterling := range strings.Split(strings.Trim(url.Path, "/"), "/") {
		buffer.WriteRune('_')
		buffer.WriteString(NormalizeStringToFilePath(shatterling))
	}

	return filepath.Join(ExpandPath(this.TempBaseFolder), TEMPFOLDERNAME, buffer.String())
}

func (this GGMirror) Update(config GGMConfig) {
	folder := this.GetTargetFolder()

	if !PathIsValid(folder) {
		EXIT_ERROR("The temporary ggm path is not a valid path '"+folder+"'", EXIT_FILESYSTEM_ACCESS_ERROR)
	}

	err := os.MkdirAll(folder, 0777)

	if err != nil {
		EXIT_ERROR("Cannot create tmp folder '"+folder+"'", EXIT_FILESYSTEM_ACCESS_ERROR)
	}

	repo := GitController{Folder: folder}

	if repo.ExistsLocal() {
		repo.GarbageCollect()
	}

	if this.AutoBranchDiscovery {
		repo.CloneOrPull(this.PrimaryBranch, this.Source, this.SourceCredentials, config.CredentialMode, config.AlwaysCleanNetRC)
		this.Branches = repo.ListLocalBranches()

		for _, branch := range this.Branches {
			LOG_OUT("Found branch " + branch + " in source-remote")
		}

		LOG_OUT("")
	}

	for _, branch := range this.Branches {
		LOG_OUT("Getting branch " + branch + " from source-remote")
		repo.CloneOrPull(branch, this.Source, this.SourceCredentials, config.CredentialMode, config.AlwaysCleanNetRC)

		LOG_OUT("Pushing branch " + branch + " to target-remote")
		repo.PushBack(branch, this.Target, this.TargetCredentials, config.CredentialMode, this.Force, config.AutoForceFallback, config.AlwaysCleanNetRC)
	}
}

func (this GGMirror) CleanFolder() {
	folder := this.GetTargetFolder()

	if !PathIsValid(folder) {
		EXIT_ERROR("The temporary ggm path is not a valid path '"+folder+"'", EXIT_FILESYSTEM_ACCESS_ERROR)
	}

	if PathExists(folder) {
		err := CleanFolder(folder)

		if err != nil {
			EXIT_ERROR("Cannot clean tmp folder '"+folder+"'", EXIT_FILESYSTEM_ACCESS_ERROR)
		}
	}

	err := os.MkdirAll(folder, 0777)

	if err != nil {
		EXIT_ERROR("Cannot create tmp folder '"+folder+"'", EXIT_FILESYSTEM_ACCESS_ERROR)
	}

}

func (this GGMirror) OutputStatus(config GGMConfig) {
	folderLocal := this.GetTargetFolder()

	valName := forceStrLen(this.GetShortName(), STAT_COL_NAME)

	if this.AutoBranchDiscovery {

		repo := GitController{Folder: folderLocal}
		repo.SetSilent()

		if repo.ExistsLocal() {
			this.Branches = repo.ListLocalBranches()

			if len(this.Branches) > 0 {
				for _, branch := range this.Branches {
					valBranch := forceStrLen(branch, STAT_COL_BRANCH)
					valSource := forceStrLen(this.GetStatusSource(config, branch, 8), STAT_COL_SOURCE)
					valLocal := forceStrLen(this.GetStatusLocal(branch, 8), STAT_COL_LOCAL)
					valRemote := forceStrLen(this.GetStatusRemote(config, branch, 8), STAT_COL_TARGET)
					LOG_OUT(diff(valSource, valLocal, valRemote, "X", " ") + "| " + valName + "| " + valBranch + "| " + valSource + " | " + valLocal + " | " + valRemote)
				}
			} else {
				valBranch := forceStrLen("NO BRANCHES", STAT_COL_BRANCH)
				valSource := forceStrLen("ERROR", STAT_COL_SOURCE)
				valLocal := forceStrLen("ERROR", STAT_COL_LOCAL)
				valRemote := forceStrLen("ERROR", STAT_COL_TARGET)
				LOG_OUT("X| " + valName + "| " + valBranch + "| " + valSource + " | " + valLocal + " | " + valRemote)
			}
		} else {
			valBranch := forceStrLen("NO REPO", STAT_COL_BRANCH)
			valSource := forceStrLen("ERROR", STAT_COL_SOURCE)
			valLocal := forceStrLen("N/A", STAT_COL_LOCAL)
			valRemote := forceStrLen("ERROR", STAT_COL_TARGET)
			LOG_OUT("X| " + valName + "| " + valBranch + "| " + valSource + " | " + valLocal + " | " + valRemote)
		}
	} else {
		for _, branch := range this.Branches {
			valBranch := forceStrLen(branch, STAT_COL_BRANCH)
			valSource := forceStrLen(this.GetStatusSource(config, branch, 8), STAT_COL_SOURCE)
			valLocal := forceStrLen(this.GetStatusLocal(branch, 8), STAT_COL_LOCAL)
			valRemote := forceStrLen(this.GetStatusRemote(config, branch, 8), STAT_COL_TARGET)
			LOG_OUT(diff(valSource, valLocal, valRemote, "X", " ") + "| " + valName + "| " + valBranch + "| " + valSource + " | " + valLocal + " | " + valRemote)
		}
	}
}

func (this GGMirror) GetShortName() string {
	shatterlings := strings.Split(strings.Trim(this.Source, "/"), "/")
	sn := shatterlings[len(shatterlings)-1]
	if strings.HasSuffix(strings.ToLower(sn), ".git") {
		sn = sn[:len(sn)-4]
	}
	return sn
}

func (this GGMirror) GetStatusLocal(branch string, hashlen int) string {
	folder := this.GetTargetFolder()

	if !PathIsValid(folder) {
		return "N/A"
	}

	repo := GitController{Folder: folder}
	repo.SetSilent()

	if !repo.ExistsLocal() {
		return "N/A"
	}

	exitcode, stdout, _, err := CmdRun(folder, true, "git", "show-ref", branch)

	if err != nil {
		return "ERROR"
	}

	if exitcode != 0 {
		return "ERROR"
	}

	return stdout[:hashlen]
}

func (this GGMirror) GetStatusSource(config GGMConfig, branch string, hashlen int) string {
	return this.GetStatus(config, this.Source, this.SourceCredentials, branch, hashlen)
}

func (this GGMirror) GetStatusRemote(config GGMConfig, branch string, hashlen int) string {
	return this.GetStatus(config, this.Target, this.TargetCredentials, branch, hashlen)
}

func (this GGMirror) GetStatus(config GGMConfig, url string, cred GGCredentials, branch string, hashlen int) string {
	folder := this.GetTargetFolder()

	if !PathIsValid(folder) {
		folder = this.TempBaseFolder
	}

	if !PathIsValid(folder) {
		return "ERROR"
	}

	repo := GitController{Folder: folder}
	repo.SetSilent()

	exitcode, stdout, _ := repo.ExecCredGitCommandSafe(cred, config.CredentialMode, config.AlwaysCleanNetRC, cred.NoSSLVerify, "ls-remote", url, branch)

	if exitcode != 0 {
		return "ERROR"
	}

	if hashlen > len(stdout) {
		hashlen = len(stdout)
	}

	return stdout[:hashlen]
}
