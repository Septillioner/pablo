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
			if cmd.Name() == "lsp" || cmd.Name() == "inspect" || cmd.Name() == "update" {
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
  pablo run --profile api --env production
  pablo check --file my-pipeline.yaml
  pablo init
  pablo lsp
  pablo update`

	var runCmd = &cobra.Command{
		Use:   "run [profile/env]",
		Short: "Executes the deployment pipeline",
		Args:  cobra.MaximumNArgs(1),
		Example: `  pablo run default/windows-local
  pablo run default.windows-local -f pablo.yaml
  pablo run -p api -e staging`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile, env, err := resolveRunTarget(cmd, args)
			if err != nil {
				return err
			}
			return pipelineSvc.Run(manifest, profile, env, allowProtected)
		},
	}
	runCmd.Flags().StringVarP(&envName, "env", "e", "production", "Target environment")
	runCmd.Flags().StringVarP(&profileName, "profile", "p", "default", "Target profile")
	runCmd.Flags().StringVarP(&manifest, "file", "f", "pablo.yaml", "Path to manifest")
	runCmd.Flags().BoolVar(&allowProtected, "force", false, "Allow deployment to protected system directories")

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initializes a new pablo.yaml sample",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Log(">", "Initializing sample pablo.yaml...")
			sample := fmt.Sprintf(`name: my-app
version: %s
profiles:
  default:
    type: static
    build:
      command: npm run build
      output_dir: ./dist
    artifacts:
      include: ["**/*"]
    environments:
      production:
        deploy:
          target_path: /var/www/check
          strategy: backup
`, Version)
			err := os.WriteFile("pablo_sample.yaml", []byte(sample), 0644)
			if err != nil {
				ui.Log("-", fmt.Sprintf("Failed to create sample: %v", err))
				return
			}
			ui.Log("+", "Sample pablo_sample.yaml created successfully.")
		},
	}

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

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Displays Pablo version information",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Log("*", fmt.Sprintf("Pablo Version: %s", Version))
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

	var lspCmd = &cobra.Command{
		Use:   "lsp",
		Short: "Start Pablo language server (stdio)",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return lsp.RunStdio(Version)
		},
	}

	var updateCheck bool
	var updateVersion string
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the Pablo CLI binary from GitHub Releases",
		Example: `  pablo update
  pablo update --check
  pablo update --version v1.5.0`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(updateCheck, updateVersion)
		},
	}
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "Check for updates without downloading")
	updateCmd.Flags().StringVar(&updateVersion, "version", "", "Pin a release tag (e.g. v1.5.0); also reads PABLO_VERSION")

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
	return nil
}

func runUpdate(checkOnly bool, pinnedVersion string) error {
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
			ui.Log("-", fmt.Sprintf("Update check failed: %v", err))
			return err
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
