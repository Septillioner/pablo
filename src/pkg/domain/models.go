package domain

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name        string                      `yaml:"name"`
	Version     string                      `yaml:"version"`
	Credentials map[string]CredentialConfig `yaml:"credentials,omitempty"`
	Sequences   map[string][]string         `yaml:"sequences,omitempty"`
	Profiles    map[string]Profile          `yaml:"profiles"`

	// BaseDir is the directory where the manifest file is located.
	BaseDir string `yaml:"-"`
}

type CredentialConfig struct {
	Type       string `yaml:"type"`
	Username   string `yaml:"username,omitempty"`
	Password   string `yaml:"password,omitempty"`
	Key        string `yaml:"key,omitempty"`
	Passphrase string `yaml:"passphrase,omitempty"`
	Value      string `yaml:"value,omitempty"`
}

type EnvConfig struct {
	Variables map[string]string `yaml:"variables,omitempty"`
	EnvFile   string            `yaml:"env_file,omitempty"`
}

type Profile struct {
	Type         string `yaml:"type"`
	EnvConfig    `yaml:",inline"`
	Build        *BuildConfig           `yaml:"build,omitempty"`
	Git          *GitConfig             `yaml:"git,omitempty"`
	Environments map[string]Environment `yaml:"environments"`
}

type GitConfig struct {
	Repo       string `yaml:"repo"`
	Branch     string `yaml:"branch,omitempty"`
	Credential string `yaml:"credential,omitempty"`
}

type ArtifactsConfig struct {
	Dir     string   `yaml:"dir,omitempty"`
	Include []string `yaml:"include,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
}

type BuildConfig struct {
	Command   string `yaml:"command"`
	Path      string `yaml:"path,omitempty"`
	EnvConfig `yaml:",inline"`
}

type Environment struct {
	Deploy       DeployConfig        `yaml:"deploy"`
	Remote       *RemoteConfig       `yaml:"remote,omitempty"`
	Build        *BuildConfig        `yaml:"build,omitempty"`
	EnvConfig    `yaml:",inline"`
	RegisterPath *RegisterPathConfig `yaml:"register_path,omitempty"`
}

type RemoteConfig struct {
	Host                string `yaml:"host"`
	Credential          string `yaml:"credential"`
	HostKeyVerification string `yaml:"host_key_verification,omitempty"`
	TrustOnFirstUse     string `yaml:"trust_on_first_use,omitempty"`
}

type DeployConfig struct {
	Source         *ArtifactsConfig `yaml:"source,omitempty"`
	TargetPath     string           `yaml:"target_path"`
	Strategy       string           `yaml:"strategy,omitempty"`
	Docker         *DockerConfig    `yaml:"docker,omitempty"`
	BlueGreen      *BlueGreenConfig `yaml:"blue_green,omitempty"`
	PreCommands    []DeployCommand  `yaml:"pre_commands,omitempty"`
	PostCommands   []DeployCommand  `yaml:"post_commands,omitempty"`
	Transfer       string           `yaml:"transfer,omitempty"`
	VerifyChecksum bool             `yaml:"verify_checksum,omitempty"`
}

const (
	DeployCwdProject = "project"
	DeployCwdTarget  = "target"
)

// DeployCommand is a pre/post shell step. YAML may be a plain string or
// { command, cwd } where cwd is "project" (manifest dir) or "target"
// (target_path / blue-green idle slot).
type DeployCommand struct {
	Command string `yaml:"command,omitempty"`
	Cwd     string `yaml:"cwd,omitempty"`
}

func (c *DeployCommand) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		return nil
	}
	switch value.Kind {
	case yaml.ScalarNode:
		c.Command = value.Value
		return nil
	case yaml.MappingNode:
		var raw struct {
			Command string `yaml:"command"`
			Cwd     string `yaml:"cwd"`
		}
		if err := value.Decode(&raw); err != nil {
			return err
		}
		c.Command = raw.Command
		c.Cwd = raw.Cwd
		return nil
	default:
		return fmt.Errorf("deploy command must be a string or an object with command")
	}
}

type BlueGreenConfig struct {
	Slots         []SlotConfig `yaml:"slots"`
	DetectCommand string       `yaml:"detect_command"`
	SwitchCommand string       `yaml:"switch_command,omitempty"`
}

type SlotConfig struct {
	Path          string `yaml:"path"`
	Key           string `yaml:"key,omitempty"`
	SwitchCommand string `yaml:"switch_command,omitempty"`
}

type DockerConfig struct {
	ComposeFile    string `yaml:"compose_file"`
	Build          bool   `yaml:"build,omitempty"`
	StopBeforeSync *bool  `yaml:"stop_before_sync,omitempty"`
}

// StopBeforeSyncEnabled is true when stop_before_sync is omitted or explicitly true.
func (d *DockerConfig) StopBeforeSyncEnabled() bool {
	if d == nil || d.StopBeforeSync == nil {
		return true
	}
	return *d.StopBeforeSync
}

type RegisterPathConfig struct {
	Scope string `yaml:"scope"`
}

func (c *Config) GetProfile(name string) (*Profile, error) {
	profile, ok := c.Profiles[name]
	if !ok {
		return nil, nil
	}
	return &profile, nil
}

func (c *Config) GetCredential(name string) (*CredentialConfig, error) {
	if name == "" {
		return nil, nil
	}
	cred, ok := c.Credentials[name]
	if !ok {
		return nil, nil
	}
	return &cred, nil
}
