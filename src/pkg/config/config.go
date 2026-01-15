package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name         string                 `yaml:"name"`
	Version      string                 `yaml:"version"`
	Type         string                 `yaml:"type"`
	Source       SourceConfig           `yaml:"source"`
	Artifacts    ArtifactsConfig        `yaml:"artifacts"`
	Environments map[string]Environment `yaml:"environments"`
	Pipeline     PipelineConfig         `yaml:"pipeline"`
	Hooks        LifecycleHooks         `yaml:"hooks"` // Added for the first example structure
}

type SourceConfig struct {
	Path         string `yaml:"path"`
	BuildCommand string `yaml:"build_command"`
}

type ArtifactsConfig struct {
	BasePath string   `yaml:"base_path"`
	Exclude  []string `yaml:"exclude"`
	Include  []string `yaml:"include"`
}

type Environment struct {
	TargetPath string            `yaml:"target_path"`
	Variables  map[string]string `yaml:"variables"`
	Strategy   string            `yaml:"strategy"` // blue-green, backup, overwrite
}

type PipelineConfig struct {
	OnSuccess   string `yaml:"on_success"`
	OnFailure   string `yaml:"on_failure"`
	HealthCheck string `yaml:"health_check"`
}

type LifecycleHooks struct {
	Pre  string `yaml:"pre"`
	Post string `yaml:"post"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
