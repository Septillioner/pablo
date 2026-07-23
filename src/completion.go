package main

import (
	"pablo/pkg/config"
	"pablo/pkg/domain"
	"pablo/pkg/target"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

const (
	defaultManifestFile = "pablo.yaml"
	sequenceArgKeyword  = "sequence"
	fileFlagName        = "file"
	profileFlagName     = "profile"
	envFlagName         = "env"
)

func registerManifestCompletions(cmd *cobra.Command) {
	_ = cmd.RegisterFlagCompletionFunc(fileFlagName, completeManifestFile)
	if cmd.Flags().Lookup(profileFlagName) != nil {
		_ = cmd.RegisterFlagCompletionFunc(profileFlagName, completeProfileFlag)
	}
	if cmd.Flags().Lookup(envFlagName) != nil {
		_ = cmd.RegisterFlagCompletionFunc(envFlagName, completeEnvFlag)
	}
}

func completeManifestFile(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveDefault
}

func completeProfileFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg, err := loadCompletionConfig(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return profileNames(cfg), cobra.ShellCompDirectiveNoFileComp
}

func completeEnvFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg, err := loadCompletionConfig(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if cmd.Flags().Changed(profileFlagName) {
		profile, _ := cmd.Flags().GetString(profileFlagName)
		if p, ok := cfg.Profiles[profile]; ok {
			return sortedKeys(p.Environments), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return allEnvironmentNames(cfg), cobra.ShellCompDirectiveNoFileComp
}

func completeRunArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	cfg, err := loadCompletionConfig(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	if len(args) == 1 {
		if args[0] != sequenceArgKeyword {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return sequenceNames(cfg), cobra.ShellCompDirectiveNoFileComp
	}

	suggestions := make([]string, 0, 1+len(cfg.Profiles)*4)
	suggestions = append(suggestions, sequenceArgKeyword)
	suggestions = append(suggestions, profileEnvTargets(cfg, toComplete)...)
	return suggestions, cobra.ShellCompDirectiveNoFileComp
}

func loadCompletionConfig(cmd *cobra.Command) (*domain.Config, error) {
	path := defaultManifestFile
	if cmd != nil {
		if v, err := cmd.Flags().GetString(fileFlagName); err == nil && v != "" {
			path = v
		}
	}
	return config.NewLoader().Load(path)
}

func profileNames(cfg *domain.Config) []string {
	return sortedKeys(cfg.Profiles)
}

func sequenceNames(cfg *domain.Config) []string {
	return sortedKeys(cfg.Sequences)
}

func allEnvironmentNames(cfg *domain.Config) []string {
	seen := make(map[string]struct{})
	for _, profile := range cfg.Profiles {
		for envName := range profile.Environments {
			seen[envName] = struct{}{}
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func profileEnvTargets(cfg *domain.Config, toComplete string) []string {
	prefixProfile := ""
	if slash := strings.Index(toComplete, "/"); slash >= 0 {
		prefixProfile = toComplete[:slash]
	}

	var out []string
	for _, profileName := range profileNames(cfg) {
		if prefixProfile != "" && profileName != prefixProfile {
			continue
		}
		profile := cfg.Profiles[profileName]
		for _, envName := range sortedKeys(profile.Environments) {
			out = append(out, target.Format(profileName, envName))
		}
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
