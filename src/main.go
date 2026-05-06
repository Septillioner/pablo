package main

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"pablo/internal/adapters/docker"
	"pablo/internal/config"
	"pablo/internal/services/builder"
	"pablo/internal/services/deployer"
	"pablo/internal/services/pipeline"
	"pablo/internal/services/scm"
	"pablo/pkg/ui"

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
	// Dependency Injection Root
	cfgLoader := config.NewLoader()
	deployerSvc := deployer.New()
	builderSvc := builder.New()
	scmSvc := scm.New()
	dockerAdapter := docker.New()

	pipelineSvc := pipeline.New(cfgLoader, deployerSvc, builderSvc, scmSvc, dockerAdapter)

	// CLI Setup
	var rootCmd = &cobra.Command{
		Use:   "pablo",
		Short: "Pablo is a visionary DevOps assistant",
		Long: `Pablo is a production-ready CLI tool designed to simplify deployments. 
It supports multiple profiles, environment-based configurations, and automatic 
artifact filtering, path registration, and health checks.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ui.Header(Version)
		},
	}

	// Custom Usage template for the "Wow" effect
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

	rootCmd.Example = `  pablo run --profile api --env production
  pablo check --file my-pipeline.yaml
  pablo init`

	// 1. RUN Command
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Executes the deployment pipeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			return pipelineSvc.Run(manifest, profileName, envName, allowProtected)
		},
	}
	runCmd.Flags().StringVarP(&envName, "env", "e", "production", "Target environment")
	runCmd.Flags().StringVarP(&profileName, "profile", "p", "default", "Target profile")
	runCmd.Flags().StringVarP(&manifest, "file", "f", "pablo.yaml", "Path to manifest")
	runCmd.Flags().BoolVar(&allowProtected, "force", false, "Allow deployment to protected system directories")

	// 2. INIT Command
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

	// 3. CHECK Command
	var checkCmd = &cobra.Command{
		Use:   "check",
		Short: "Validates the manifest file",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Log("*", fmt.Sprintf("Checking manifest: %s", manifest))
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

	// 4. UNINSTALL Command
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

	// 5. VERSION Command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Displays Pablo version information",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Log("*", fmt.Sprintf("Pablo Version: %s", Version))
			ui.Log("*", "Architecture: Modular Monolith")
		},
	}

	rootCmd.AddCommand(runCmd, initCmd, checkCmd, uninstallCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
