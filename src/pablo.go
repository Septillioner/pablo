package main

import (
	"fmt"
	"os"
	"pablo/pkg/config"
	"pablo/pkg/deploy"
	"pablo/pkg/filter"
	"pablo/pkg/health"
	"pablo/pkg/hooks"
	"pablo/pkg/template"
	"pablo/pkg/ui"
	"time"

	"github.com/spf13/cobra"
)

var (
	envName  string
	manifest string
	Version  = "1.1.0"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "pablo",
		Short: "Pablo is a visionary DevOps assistant",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ui.Header()
		},
	}

	// 1. RUN Command (formerly up)
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Executes the deployment pipeline",
		RunE:  runPipeline,
	}
	runCmd.Flags().StringVarP(&envName, "env", "e", "production", "Target environment")
	runCmd.Flags().StringVarP(&manifest, "file", "f", "pablo.yaml", "Path to manifest")

	// 2. INIT Command
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initializes a new pablo.yaml sample",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Log(">", "Initializing sample pablo.yaml...")
			sample := `name: my-app
version: 1.0.0
type: static
source:
  path: ./src
  build_command: "npm run build"
artifacts:
  base_path: ./dist
  include: ["*"]
  exclude: ["*.log"]
environments:
  production:
    target_path: /var/www/html
    strategy: backup
    variables:
      DB_HOST: "prod-db"
pipeline:
  health_check: "http://localhost:8080/health"
`
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
			cfg, err := config.LoadConfig(manifest)
			if err != nil {
				ui.Log("-", fmt.Sprintf("Validation failed: %v", err))
				os.Exit(1)
			}
			ui.Log("+", fmt.Sprintf("Manifest is valid! (Project: %s, Version: %s)", cfg.Name, cfg.Version))
		},
	}
	checkCmd.Flags().StringVarP(&manifest, "file", "f", "pablo.yaml", "Path to manifest")

	// 4. VERSION Command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Displays Pablo version information",
		Run: func(cmd *cobra.Command, args []string) {
			ui.Log("*", fmt.Sprintf("Pablo Version: %s", Version))
			ui.Log("*", "Build: Development")
		},
	}

	rootCmd.AddCommand(runCmd, initCmd, checkCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runPipeline(cmd *cobra.Command, args []string) error {
	start := time.Now()

	ui.Log("*", fmt.Sprintf("Loading manifest: %s", manifest))
	cfg, err := config.LoadConfig(manifest)
	if err != nil {
		ui.Log("-", "Failed to load manifest")
		ui.Result(false, time.Since(start))
		return err
	}

	env, ok := cfg.Environments[envName]
	if !ok {
		ui.Log("-", fmt.Sprintf("Environment '%s' not found", envName))
		ui.Result(false, time.Since(start))
		return fmt.Errorf("environment not found")
	}

	ui.Section("Deployment Info")
	ui.Log("*", fmt.Sprintf("Project: %s", cfg.Name))
	ui.Log("*", fmt.Sprintf("Version: %s", cfg.Version))
	ui.Log("*", fmt.Sprintf("Target:  %s", envName))

	// 1. Pre-deploy hooks
	ui.Section("Phase 1: Pre-Deploy")
	if cfg.Hooks.Pre != "" {
		ui.Log(">", fmt.Sprintf("Executing: %s", cfg.Hooks.Pre))
		if err := hooks.Execute(cfg.Hooks.Pre); err != nil {
			ui.Log("-", "Pre-deploy hook failed")
			ui.Result(false, time.Since(start))
			return err
		}
		ui.Log("+", "Pre-deploy hook completed")
	} else {
		ui.Log("*", "No pre-deploy hooks defined")
	}

	// 2. Build
	ui.Section("Phase 2: Build")
	if cfg.Source.BuildCommand != "" {
		ui.Log(">", fmt.Sprintf("Running build: %s", cfg.Source.BuildCommand))
		ui.ProgressBar(50, "Building")
		if err := hooks.Execute(cfg.Source.BuildCommand); err != nil {
			ui.Log("-", "Build failed")
			ui.Result(false, time.Since(start))
			return err
		}
		ui.ProgressBar(100, "Building")
		ui.Log("+", "Build completed")
	} else {
		ui.Log("*", "No build command defined")
	}

	// 3. Filtering & Packaging
	ui.Section("Phase 3: Artifacts")
	ui.Log("*", "Filtering files...")
	files, err := filter.GetFiles(cfg.Artifacts.BasePath, cfg.Artifacts.Include, cfg.Artifacts.Exclude)
	if err != nil {
		ui.Log("-", "Filtering failed")
		ui.Result(false, time.Since(start))
		return err
	}
	ui.Log("+", fmt.Sprintf("Found %d artifact(s) to deploy", len(files)))

	// 4. Deployment
	ui.Section("Phase 4: Deployment")
	ui.Log(">", fmt.Sprintf("Deploying to %s (Strategy: %s)", env.TargetPath, env.Strategy))
	if err := deploy.Deploy(files, cfg.Artifacts.BasePath, env.TargetPath, env.Strategy); err != nil {
		ui.Log("-", "Deployment failed")
		ui.Result(false, time.Since(start))
		return err
	}
	ui.Log("+", "Deployment successful")

	// 5. Template Engine
	ui.Section("Phase 5: Post-Processing")
	if len(env.Variables) > 0 {
		ui.Log("*", "Applying template variables...")
		if err := template.ProcessFiles(env.TargetPath, env.Variables); err != nil {
			ui.Log("-", "Template processing failed")
			ui.Result(false, time.Since(start))
			return err
		}
		ui.Log("+", "Template processing completed")
	} else {
		ui.Log("*", "No variables to process")
	}

	// 6. Post-deploy hooks
	if cfg.Hooks.Post != "" {
		ui.Log(">", fmt.Sprintf("Executing: %s", cfg.Hooks.Post))
		if err := hooks.Execute(cfg.Hooks.Post); err != nil {
			if cfg.Pipeline.OnFailure != "" {
				hooks.Execute(cfg.Pipeline.OnFailure)
			}
			ui.Log("-", "Post-deploy hook failed")
			ui.Result(false, time.Since(start))
			return err
		}
		ui.Log("+", "Post-deploy hook completed")
	}

	// 7. Health Check
	if cfg.Pipeline.HealthCheck != "" {
		ui.Section("Phase 6: Health Check")
		ui.Log(">", fmt.Sprintf("Verifying: %s", cfg.Pipeline.HealthCheck))
		if err := health.Check(cfg.Pipeline.HealthCheck, 30*time.Second); err != nil {
			if cfg.Pipeline.OnFailure != "" {
				hooks.Execute(cfg.Pipeline.OnFailure)
			}
			ui.Log("-", "Health check failed")
			ui.Result(false, time.Since(start))
			return err
		}
		ui.Log("+", "Health check passed")
	}

	// 8. Success
	if cfg.Pipeline.OnSuccess != "" {
		hooks.Execute(cfg.Pipeline.OnSuccess)
	}

	ui.Result(true, time.Since(start))
	return nil
}
