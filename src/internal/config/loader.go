package config

import (
	"os"
	"pablo/internal/domain"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Loader handles configuration loading
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

	var cfg domain.Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	// Set BaseDir to the directory where the manifest is located
	cfg.BaseDir = filepath.Dir(absPath)

	// Handle inheritance: Propagate Profile-level settings to Environments
	for name, profile := range cfg.Profiles {

		// 1. Inherit from Profile to Environments
		for envName, env := range profile.Environments {

			// A. Propagate Global Variables to Environment level
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

			// B. Propagate Build Config (default if not set in env)
			if profile.Build != nil {
				if env.Build == nil {
					b := *profile.Build
					env.Build = &b
				} else {
					// Merge profile build into env build
					if env.Build.Path == "" {
						env.Build.Path = profile.Build.Path
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

			// C. Propagate OutputDir (Artifacts) to Deploy.Source
			if profile.OutputDir.Dir != "" || len(profile.OutputDir.Include) > 0 {
				if env.Deploy.Source == nil {
					src := profile.OutputDir
					env.Deploy.Source = &src
				}
			}

			// D. Propagate Variables from Environment to Deploy level
			if env.EnvConfig.Variables != nil {
				if env.Deploy.EnvConfig.Variables == nil {
					env.Deploy.EnvConfig.Variables = make(map[string]string)
				}
				for k, v := range env.EnvConfig.Variables {
					if _, exists := env.Deploy.EnvConfig.Variables[k]; !exists {
						env.Deploy.EnvConfig.Variables[k] = v
					}
				}
			}
			if env.Deploy.EnvConfig.EnvFile == "" && env.EnvConfig.EnvFile != "" {
				env.Deploy.EnvConfig.EnvFile = env.EnvConfig.EnvFile
			}

			profile.Environments[envName] = env
		}

		cfg.Profiles[name] = profile
	}

	// Backward compatibility transformation
	if len(cfg.Profiles) == 0 && cfg.Type != "" {
		legacyEnvs := make(map[string]domain.Environment)
		for envName, env := range cfg.Environments {
			newEnv := domain.Environment{
				Deploy: domain.DeployConfig{
					TargetPath: env.TargetPath,
					Strategy:   env.Strategy,
				},
				EnvConfig: domain.EnvConfig{
					Variables: env.Variables,
				},
				RegisterPath: env.RegisterPath,
			}
			legacyEnvs[envName] = newEnv
		}

		cfg.Profiles = map[string]domain.Profile{
			"default": {
				Type: cfg.Type,
				OutputDir: domain.ArtifactsConfig{
					Include: cfg.Source.Include,
					Exclude: cfg.Source.Exclude,
				},
				Environments: legacyEnvs,
				Pipeline:     cfg.Pipeline,
				Hooks:        cfg.Hooks,
			},
		}
	}

	return &cfg, nil
}
