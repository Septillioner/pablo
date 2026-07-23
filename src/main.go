package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pablo/internal/adapters/docker"
	"pablo/internal/lsp"
	"pablo/internal/services/builder"
	"pablo/internal/services/deployer"
	"pablo/internal/services/initcmd"
	"pablo/internal/services/pipeline"
	"pablo/internal/services/scm"
	"pablo/internal/services/selfupdate"
	"pablo/pkg/config"
	"pablo/pkg/inspect"
	"pablo/pkg/target"
	"pablo/pkg/ui"
	"pablo/pkg/validate"

	"github.com/spf13/cobra"
)

//go:embed VERSION
var versionFile embed.FS

var (
	envName        string
	profileName    string
	manifest       string
	allowProtected bool
	verbose        bool
	Version        string
)

func init() {
	v, _ := versionFile.ReadFile("VERSION")
	Version = strings.TrimSpace(string(v))
}

func main() {
	cfgLoader := config.NewLoader()
	deployerSvc := deployer.New()
	builderSvc := builder.New()
	scmSvc := scm.New()
	dockerAdapter := docker.New()

	pipelineSvc := pipeline.New(cfgLoader, deployerSvc, builderSvc, scmSvc, dockerAdapter)

	var rootCmd = &cobra.Command{
		Use:   "pablo",
		Short: "Pablo is a visionary DevOps assistant",
		Long: `Pablo is a production-ready CLI tool designed to simplify deployments. 
It supports multiple profiles, environment-based configurations, and automatic 
artifact filtering, path registration, and health checks.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if shouldSkipBrandHeader(cmd) {
				return
			}
			ui.Header(Version)
		},
	}

	rootCmd.SetUsageTemplate(`USAGE:
  {{.UseLine}}

{{if .HasAvailableSubCommands}}COMMANDS:
{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}  {{rpad .Name .NamePadding}} {{.Short}}
{{end}}{{end}}{{end}}
{{if .HasAvailableLocalFlags}}FLAGS:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

{{end}}{{if .HasExample}}EXAMPLES:
{{.Example}}
{{end}}
USE "pablo [command] --help" FOR MORE INFORMATION ABOUT A COMMAND.
`)

	rootCmd.Example = `  pablo run default/windows-local
  pablo run sequence extension
  pablo run --profile api --env production
  pablo check --file my-pipeline.yaml
  pablo init
  pablo lsp
  pablo update
  pablo update check`

	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "List each artifact path during deployment")

	var runCmd = &cobra.Command{
		Use:   "run [profile/env | sequence <name>]",
		Short: "Executes the deployment pipeline",
		Args:  cobra.MaximumNArgs(2),
		Example: `  pablo run default/windows-local
  pablo run default.windows-local -f pablo.yaml
  pablo run sequence extension
  pablo run -p api -e staging`,
		ValidArgsFunction: completeRunArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && args[0] == "sequence" {
				if len(args) < 2 {
					return fmt.Errorf("sequence name is required (usage: pablo run sequence <name>)")
				}
				if cmd.Flags().Changed("profile") || cmd.Flags().Changed("env") {
					return fmt.Errorf("cannot combine sequence with -p/--profile or -e/--env")
				}
				return pipelineSvc.RunSequence(manifest, args[1], allowProtected, verbose)
			}
			if len(args) == 2 {
				return fmt.Errorf("unexpected second argument; use 'pablo run sequence <name>' for sequences")
			}
			profile, env, err := resolveRunTarget(cmd, args)
			if err != nil {
				return err
			}
			return pipelineSvc.Run(manifest, profile, env, allowProtected, verbose)
		},
	}
	runCmd.Flags().StringVarP(&envName, "env", "e", "production", "Target environment")
	runCmd.Flags().StringVarP(&profileName, "profile", "p", "default", "Target profile")
	runCmd.Flags().StringVarP(&manifest, "file", "f", "pablo.yaml", "Path to manifest")
	runCmd.Flags().BoolVar(&allowProtected, "force", false, "Allow deployment to protected system directories")
	registerManifestCompletions(runCmd)

	var useTemplate bool
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initializes a new pablo.yaml sample",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initcmd.Run(initcmd.Options{
				Version:        Version,
				TemplateWizard: useTemplate,
			}); err != nil {
				os.Exit(1)
			}
		},
	}
	initCmd.Flags().BoolVarP(&useTemplate, "template", "t", false, "Interactive template type wizard")

	var checkCmd = &cobra.Command{
		Use:   "check",
		Short: "Validates the manifest file",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Log("*", fmt.Sprintf("Checking manifest: %s", manifest))
			if err := printManifestValidation(manifest); err != nil {
				os.Exit(1)
			}

			cfg, err := cfgLoader.Load(manifest)
			if err != nil {
				ui.Log("-", fmt.Sprintf("Validation failed: %v", err))
				os.Exit(1)
			}
			ui.Log("+", fmt.Sprintf("Manifest is valid! (Project: %s, Version: %s)", cfg.Name, cfg.Version))

			if profileName != "" {
				profile, _ := cfg.GetProfile(profileName)
				if profile == nil {
					ui.Log("-", fmt.Sprintf("Profile '%s' NOT found", profileName))
				} else {
					ui.Log("+", fmt.Sprintf("Profile '%s' (type: %s) found", profileName, profile.Type))
					if envName != "" {
						if _, ok := profile.Environments[envName]; ok {
							ui.Log("+", fmt.Sprintf("Environment '%s' found", envName))
						} else {
							ui.Log("-", fmt.Sprintf("Environment '%s' NOT found", envName))
						}
					}
				}
			}
		},
	}
	checkCmd.Flags().StringVarP(&manifest, "file", "f", "pablo.yaml", "Path to manifest")
	checkCmd.Flags().StringVarP(&profileName, "profile", "p", "", "Validate specific profile")
	checkCmd.Flags().StringVarP(&envName, "env", "e", "", "Validate specific environment")
	registerManifestCompletions(checkCmd)

	var removeBackups bool
	var uninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Removes deployed files and cleans up PATH entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			if envName == "" {
				ui.Log("-", "Environment (-e) is required for uninstall")
				return fmt.Errorf("environment flag is required")
			}
			return pipelineSvc.Uninstall(manifest, profileName, envName, removeBackups)
		},
	}
	uninstallCmd.Flags().StringVarP(&manifest, "file", "f", "pablo.yaml", "Path to manifest")
	uninstallCmd.Flags().StringVarP(&profileName, "profile", "p", "default", "Profile to uninstall")
	uninstallCmd.Flags().StringVarP(&envName, "env", "e", "", "Environment to uninstall (required)")
	uninstallCmd.Flags().BoolVar(&removeBackups, "remove-backups", false, "Also remove backup directories")
	registerManifestCompletions(uninstallCmd)

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Displays Pablo version information",
		Run: func(cmd *cobra.Command, args []string) {
			// Brand + version come from PersistentPreRun Header.
		},
	}

	var inspectJson bool
	var inspectCmd = &cobra.Command{
		Use:   "inspect",
		Short: "Lists profiles and environments from a manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInspect(manifest, inspectJson)
		},
	}
	inspectCmd.Flags().StringVarP(&manifest, "file", "f", "pablo.yaml", "Path to manifest")
	inspectCmd.Flags().BoolVar(&inspectJson, "json", false, "Output JSON")
	registerManifestCompletions(inspectCmd)

	var lspCmd = &cobra.Command{
		Use:   "lsp",
		Short: "Start Pablo language server (stdio)",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return lsp.RunStdio(Version)
		},
	}

	var updateCheckFlag bool
	var updateVersion string
	var updateJSON bool
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the Pablo CLI binary from GitHub Releases",
		Example: `  pablo update
  pablo update check
  pablo update check --json
  pablo update --version v1.5.0`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(updateCheckFlag, updateVersion, updateJSON)
		},
	}
	updateCmd.Flags().BoolVar(&updateCheckFlag, "check", false, "Deprecated: use 'pablo update check'")
	updateCmd.PersistentFlags().StringVar(&updateVersion, "version", "", "Pin a release tag (e.g. v1.5.0); also reads PABLO_VERSION")
	updateCmd.PersistentFlags().BoolVar(&updateJSON, "json", false, "Machine-readable JSON output (check only)")

	var updateCheckCmd = &cobra.Command{
		Use:   "check",
		Short: "Check whether a newer CLI release is available",
		Example: `  pablo update check
  pablo update check --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(true, updateVersion, updateJSON)
		},
	}
	updateCmd.AddCommand(updateCheckCmd)

	rootCmd.AddCommand(runCmd, initCmd, checkCmd, uninstallCmd, versionCmd, inspectCmd, lspCmd, updateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printManifestValidation(manifestPath string) error {
	absPath, err := filepath.Abs(manifestPath)
	if err != nil {
		ui.Log("-", fmt.Sprintf("Validation failed: %v", err))
		return err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		ui.Log("-", fmt.Sprintf("Validation failed: %v", err))
		return err
	}

	diags, _, err := validate.ValidateYAML(data, filepath.Dir(absPath))
	if err != nil {
		ui.Log("-", fmt.Sprintf("Validation failed: %v", err))
		return err
	}

	for _, d := range diags {
		ui.Log("-", validate.FormatDiagnostic(absPath, d))
	}

	if validate.HasErrors(diags) {
		return fmt.Errorf("manifest validation failed")
	}

	return nil
}

func resolveRunTarget(cmd *cobra.Command, args []string) (string, string, error) {
	if len(args) == 0 {
		return profileName, envName, nil
	}

	if cmd.Flags().Changed("profile") || cmd.Flags().Changed("env") {
		return "", "", fmt.Errorf("cannot combine target argument with -p/--profile or -e/--env")
	}

	profile, env, err := target.Parse(args[0])
	if err != nil {
		return "", "", err
	}
	return profile, env, nil
}

func runInspect(manifestPath string, asJSON bool) error {
	absPath, err := filepath.Abs(manifestPath)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	result, err := inspect.FromYAML(data, filepath.Dir(absPath))
	if err != nil {
		return err
	}

	if asJSON {
		encoded, err := json.Marshal(result)
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
		return nil
	}

	ui.Log("*", fmt.Sprintf("Project: %s (%s)", result.Name, result.Version))
	if len(result.Profiles) == 0 {
		ui.Log("-", "No profiles found")
		return nil
	}

	for _, profile := range result.Profiles {
		ui.Log(">", fmt.Sprintf("%s (%s)", profile.Name, profile.Type))
		for _, env := range profile.Environments {
			ui.Log(" ", fmt.Sprintf("- %s", env))
		}
	}

	if len(result.Sequences) > 0 {
		ui.Log("*", "Sequences:")
		for _, seq := range result.Sequences {
			ui.Log(">", seq.Name)
			for _, step := range seq.Steps {
				ui.Log(" ", fmt.Sprintf("- %s", step))
			}
		}
	}
	return nil
}

func shouldSkipBrandHeader(cmd *cobra.Command) bool {
	if isShellCompletionInvocation() {
		return true
	}
	for current := cmd; current != nil; current = current.Parent() {
		switch current.Name() {
		case "lsp", "inspect", "update", "completion", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
			return true
		}
	}
	return false
}

func isShellCompletionInvocation() bool {
	if len(os.Args) < 2 {
		return false
	}
	switch os.Args[1] {
	case cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd, "completion":
		return true
	default:
		return false
	}
}

type updateCheckJSON struct {
	CurrentVersion  string `json:"current_version"`
	LatestVersion   string `json:"latest_version"`
	ReleaseTag      string `json:"release_tag"`
	UpdateAvailable bool   `json:"update_available"`
}

func runUpdate(checkOnly bool, pinnedVersion string, asJSON bool) error {
	if pinnedVersion == "" {
		pinnedVersion = selfupdate.PinnedVersionFromEnv()
	}

	svc := selfupdate.New()
	opts := selfupdate.Options{
		CurrentVersion: Version,
		PinnedVersion:  pinnedVersion,
	}

	if checkOnly {
		result, err := svc.Check(opts)
		if err != nil {
			if !asJSON {
				ui.Log("-", fmt.Sprintf("Update check failed: %v", err))
			}
			return err
		}

		if asJSON {
			payload := updateCheckJSON{
				CurrentVersion:  result.CurrentVersion,
				LatestVersion:   result.LatestVersion,
				ReleaseTag:      result.ReleaseTag,
				UpdateAvailable: !result.UpToDate,
			}
			encoded, err := json.Marshal(payload)
			if err != nil {
				return err
			}
			fmt.Println(string(encoded))
			if !result.UpToDate {
				os.Exit(1)
			}
			return nil
		}

		if result.UpToDate {
			ui.Log("+", fmt.Sprintf("Pablo is up to date (%s)", result.CurrentVersion))
			return nil
		}
		ui.Log("!", fmt.Sprintf("Update available: %s -> %s (%s)", result.CurrentVersion, result.LatestVersion, result.ReleaseTag))
		os.Exit(1)
	}

	check, err := svc.Check(opts)
	if err != nil {
		ui.Log("-", fmt.Sprintf("Update failed: %v", err))
		return err
	}
	if check.UpToDate {
		ui.Log("+", fmt.Sprintf("Pablo is already up to date (%s)", check.CurrentVersion))
		return nil
	}

	result, err := svc.Update(opts)
	if err != nil {
		ui.Log("-", fmt.Sprintf("Update failed: %v", err))
		return err
	}

	ui.Log("+", fmt.Sprintf("Updated Pablo to %s (%s)", result.LatestVersion, result.ReleaseTag))
	ui.Log("*", "Open a new terminal or run pablo version to use the updated binary")
	return nil
}
