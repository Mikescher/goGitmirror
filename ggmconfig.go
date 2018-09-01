package main

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type GGMConfig struct {
	TemporaryPath       string
	AutoCleanTempFolder bool
	AutoForceFallback   bool
	AlwaysCleanNetRC    bool

	Credentials []GGCredentials

	Remote []GGMirror
}

type GGMirror struct {
	Source string
	Target string

	Force bool

	Branches            []string // If not set AutoBranchDiscovery becomes true
	AutoBranchDiscovery bool     // normally not set via TOML, but auto assigned based on host

	SourceCredentials GGCredentials // normally not set via TOML, but auto assigned based on host
	TargetCredentials GGCredentials // normally not set via TOML, but auto assigned based on host
	TempBaseFolder    string        // normally not set via TOML, but auto assigned from root config
}

type GGCredentials struct {
	Host string // Host is empty string for anonymous login

	Username string
	Password string
}

func (this *GGMConfig) LoadFromFile(path string) {

	if _, err := toml.DecodeFile(path, &this); err != nil {
		EXIT_ERROR("ERROR: Cannot load config from "+path+"\n\n"+err.Error(), EXIT_CONFIG_READ_ERROR)
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
	}

	for i := 0; i < len(this.Remote); i++ {
		if this.Remote[i].Source == "" {
			EXIT_ERROR("ERROR: Every remote must have the property 'Source' set", EXIT_CONFIG_READ_ERROR)
		}

		if this.Remote[i].Target == "" {
			EXIT_ERROR("ERROR: Every remote must have the property 'Target' set", EXIT_CONFIG_READ_ERROR)
		}

		this.Remote[i].TempBaseFolder = this.TemporaryPath

		if this.Remote[i].Branches == nil {
			this.Remote[i].Branches = []string{} // Default value
			this.Remote[i].AutoBranchDiscovery = true
		} else {
			this.Remote[i].AutoBranchDiscovery = false
		}

		urlSource, err := url.Parse(this.Remote[i].Source)
		if err != nil {
			EXIT_ERROR("ERROR: The Source '"+this.Remote[i].Source+"' is not a valid URL", EXIT_CONFIG_READ_ERROR)
		}

		urlTarget, err := url.Parse(this.Remote[i].Target)
		if err != nil {
			EXIT_ERROR("ERROR: The Target '"+this.Remote[i].Target+"' is not a valid URL", EXIT_CONFIG_READ_ERROR)
		}

		for _, cred := range this.Credentials {
			if strings.ToUpper(cred.Host) == strings.ToUpper(urlSource.Host) && this.Remote[i].SourceCredentials.Host == "" {
				this.Remote[i].SourceCredentials = cred
				this.Remote[i].SourceCredentials.Host = urlSource.Host
			}

			if strings.ToUpper(cred.Host) == strings.ToUpper(urlTarget.Host) && this.Remote[i].TargetCredentials.Host == "" {
				this.Remote[i].TargetCredentials = cred
				this.Remote[i].TargetCredentials.Host = urlTarget.Host
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

func (this GGMirror) Update(config GGMConfig) error {
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
		repo.CloneOrPull("master", this.Source, this.SourceCredentials, config.AlwaysCleanNetRC)
		this.Branches = repo.ListLocalBranches()

		for _, branch := range this.Branches {
			LOG_OUT("Found branch " + branch + " in source-remote")
		}

		LOG_OUT("")
	}

	for _, branch := range this.Branches {
		LOG_OUT("Getting branch " + branch + " from source-remote")
		repo.CloneOrPull(branch, this.Source, this.SourceCredentials, config.AlwaysCleanNetRC)

		LOG_OUT("Pushing branch " + branch + " to target-remote")
		repo.PushBack(branch, this.Target, this.TargetCredentials, this.Force, config.AutoForceFallback, config.AlwaysCleanNetRC)
	}

	return nil
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
