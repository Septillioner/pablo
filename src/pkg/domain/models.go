package domain

import "gopkg.in/yaml.v3"

type Config struct {
	Name        string                      `yaml:"name"`
	Version     string                      `yaml:"version"`
	Credentials map[string]CredentialConfig `yaml:"credentials,omitempty"`
	Profiles    map[string]Profile          `yaml:"profiles"`

	// BaseDir is the directory where the manifest file is located.
	BaseDir string `yaml:"-"`

	// Legacy fields for backward compatibility
	Type         string                 `yaml:"type,omitempty"`
	Source       SourceConfig           `yaml:"source,omitempty"`
	Environments map[string]Environment `yaml:"environments,omitempty"`
	Pipeline     PipelineConfig         `yaml:"pipeline,omitempty"`
	Hooks        LifecycleHooks         `yaml:"hooks,omitempty"`
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
	OutputDir    ArtifactsConfig        `yaml:"output_dir,omitempty"`
	Environments map[string]Environment `yaml:"environments"`
	Hooks        LifecycleHooks         `yaml:"hooks,omitempty"`
	Pipeline     PipelineConfig         `yaml:"pipeline,omitempty"`
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

func (a *ArtifactsConfig) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		a.Dir = value.Value
		return nil
	}

	type alias ArtifactsConfig
	var aux alias
	if err := value.Decode(&aux); err != nil {
		return err
	}
	*a = ArtifactsConfig(aux)
	return nil
}

type BuildConfig struct {
	Command   string `yaml:"command"`
	Path      string `yaml:"path,omitempty"`
	EnvConfig `yaml:",inline"`
}

type Environment struct {
	Deploy       DeployConfig  `yaml:"deploy"`
	Remote       *RemoteConfig `yaml:"remote,omitempty"`
	Build        *BuildConfig  `yaml:"build,omitempty"`
	EnvConfig    `yaml:",inline"`
	RegisterPath *RegisterPathConfig `yaml:"register_path,omitempty"`

	TargetPath string `yaml:"target_path,omitempty"`
	Strategy   string `yaml:"strategy,omitempty"`
}

type RemoteConfig struct {
	Method     string `yaml:"method"`
	Host       string `yaml:"host"`
	Credential string `yaml:"credential"`
}

type DeployConfig struct {
	Method       string           `yaml:"method,omitempty"`
	Source       *ArtifactsConfig `yaml:"source,omitempty"`
	SSH          *SSHConfig       `yaml:"ssh,omitempty"`
	Credential   string           `yaml:"credential,omitempty"`
	Host         string           `yaml:"host,omitempty"`
	TargetPath   string           `yaml:"target_path"`
	Strategy     string           `yaml:"strategy,omitempty"`
	Docker       *DockerConfig    `yaml:"docker,omitempty"`
	Service      *ServiceConfig   `yaml:"service,omitempty"`
	PreCommands  []string         `yaml:"pre_commands,omitempty"`
	PostCommands []string         `yaml:"post_commands,omitempty"`
	Remote       string           `yaml:"remote,omitempty"`
	EnvConfig    `yaml:",inline"`
}

type SSHConfig struct {
	Host       string `yaml:"host"`
	Credential string `yaml:"credential"`
}

type DockerConfig struct {
	ComposeFile string `yaml:"compose_file"`
	Build       bool   `yaml:"build,omitempty"`
	Command     string `yaml:"command,omitempty"`
}

type ServiceConfig struct {
	Type    string `yaml:"type"`
	Name    string `yaml:"name"`
	Restart bool   `yaml:"restart,omitempty"`
}

type RegisterPathConfig struct {
	Scope string `yaml:"scope"`
}

type PipelineConfig struct {
	OnSuccess   string `yaml:"on_success,omitempty"`
	OnFailure   string `yaml:"on_failure,omitempty"`
	HealthCheck string `yaml:"health_check,omitempty"`
}

type LifecycleHooks struct {
	Pre  string `yaml:"pre,omitempty"`
	Post string `yaml:"post,omitempty"`
}

type SourceConfig struct {
	Path    string   `yaml:"path,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
	Include []string `yaml:"include,omitempty"`
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
