package inspect

import (
	"fmt"
	"pablo/pkg/config"
	"pablo/pkg/domain"
	"pablo/pkg/validate"
	"sort"
)

type ProfileInfo struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Environments []string `json:"environments"`
}

type Result struct {
	Name     string        `json:"name"`
	Version  string        `json:"version"`
	Profiles []ProfileInfo `json:"profiles"`
}

func FromConfig(cfg *domain.Config) Result {
	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)

	profiles := make([]ProfileInfo, 0, len(names))
	for _, name := range names {
		profile := cfg.Profiles[name]
		envNames := make([]string, 0, len(profile.Environments))
		for envName := range profile.Environments {
			envNames = append(envNames, envName)
		}
		sort.Strings(envNames)

		profiles = append(profiles, ProfileInfo{
			Name:         name,
			Type:         profile.Type,
			Environments: envNames,
		})
	}

	return Result{
		Name:     cfg.Name,
		Version:  cfg.Version,
		Profiles: profiles,
	}
}

func FromYAML(data []byte, baseDir string) (Result, error) {
	diags, cfg, err := validate.ValidateYAML(data, baseDir)
	if err != nil {
		return Result{}, err
	}
	if validate.HasErrors(diags) {
		return Result{}, fmt.Errorf("manifest validation failed")
	}
	if cfg == nil {
		loader := config.NewLoader()
		cfg, err = loader.LoadFromBytes(data, baseDir)
		if err != nil {
			return Result{}, err
		}
	}
	return FromConfig(cfg), nil
}
