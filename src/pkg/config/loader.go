package config

import (
	"os"
	"pablo/pkg/domain"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) Load(path string) (*domain.Config, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	return l.LoadFromBytes(data, filepath.Dir(absPath))
}

func (l *Loader) LoadFromBytes(data []byte, baseDir string) (*domain.Config, error) {
	var cfg domain.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	cfg.BaseDir = baseDir
	applyInheritance(&cfg)
	return &cfg, nil
}

func (l *Loader) ParseDocument(data []byte) (*yaml.Node, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	if len(root.Content) == 0 {
		return &root, nil
	}
	return root.Content[0], nil
}

// applyInheritance merges profile variables, env_file, and build into each environment.
// Nothing else cascades (deploy.source is always explicit).
func applyInheritance(cfg *domain.Config) {
	for name, profile := range cfg.Profiles {
		for envName, env := range profile.Environments {
			if profile.EnvConfig.Variables != nil {
				if env.EnvConfig.Variables == nil {
					env.EnvConfig.Variables = make(map[string]string)
				}
				for k, v := range profile.EnvConfig.Variables {
					if _, exists := env.EnvConfig.Variables[k]; !exists {
						env.EnvConfig.Variables[k] = v
					}
				}
			}
			if env.EnvConfig.EnvFile == "" && profile.EnvConfig.EnvFile != "" {
				env.EnvConfig.EnvFile = profile.EnvConfig.EnvFile
			}

			if profile.Build != nil {
				if env.Build == nil {
					b := *profile.Build
					env.Build = &b
				} else {
					if env.Build.Path == "" {
						env.Build.Path = profile.Build.Path
					}
					if env.Build.Command == "" {
						env.Build.Command = profile.Build.Command
					}
					if env.Build.EnvFile == "" {
						env.Build.EnvFile = profile.Build.EnvFile
					}
					if profile.Build.Variables != nil {
						if env.Build.Variables == nil {
							env.Build.Variables = make(map[string]string)
						}
						for k, v := range profile.Build.Variables {
							if _, exists := env.Build.Variables[k]; !exists {
								env.Build.Variables[k] = v
							}
						}
					}
				}
			}

			profile.Environments[envName] = env
		}

		cfg.Profiles[name] = profile
	}
}
