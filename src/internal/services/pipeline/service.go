package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"pablo/internal/adapters/docker"
	sshAdapter "pablo/internal/adapters/ssh"
	"pablo/internal/adapters/system"
	"pablo/internal/services/builder"
	"pablo/internal/services/deployer"
	"pablo/internal/services/filter"
	"pablo/internal/services/hooks"
	"pablo/internal/services/scm"
	"pablo/internal/services/template"
	"pablo/pkg/config"
	"pablo/pkg/domain"
	"pablo/pkg/pathutil"
	"pablo/pkg/target"
	"pablo/pkg/ui"
	"pablo/pkg/validate"

	"golang.org/x/crypto/ssh"
)

type Service struct {
	loader   *config.Loader
	deployer *deployer.Service
	builder  *builder.Service
	scm      *scm.Service
	docker   *docker.Adapter
}

func New(loader *config.Loader, d *deployer.Service, b *builder.Service, s *scm.Service, doc *docker.Adapter) *Service {
	return &Service{
		loader:   loader,
		deployer: d,
		builder:  b,
		scm:      s,
		docker:   doc,
	}
}

func (s *Service) resolvePath(baseDir, path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(baseDir, path)
}

func (s *Service) Run(manifestPath, profileName, envName string, allowProtected, verbose bool) error {
	start := time.Now()

	ui.Log("*", fmt.Sprintf("Loading manifest: %s", manifestPath))
	absManifest, err := filepath.Abs(manifestPath)
	if err != nil {
		ui.Log("-", "Failed to resolve manifest path")
		ui.Result(false, time.Since(start))
		return err
	}

	manifestData, err := os.ReadFile(absManifest)
	if err != nil {
		ui.Log("-", "Failed to load manifest")
		ui.Result(false, time.Since(start))
		return err
	}

	diags, _, err := validate.ValidateYAML(manifestData, filepath.Dir(absManifest))
	if err != nil {
		ui.Log("-", "Failed to validate manifest")
		ui.Result(false, time.Since(start))
		return err
	}
	for _, d := range diags {
		if d.Severity == validate.SeverityError {
			ui.Log("-", validate.FormatDiagnostic(absManifest, d))
		}
	}
	if validate.HasErrors(diags) {
		ui.Result(false, time.Since(start))
		return fmt.Errorf("manifest validation failed")
	}

	cfg, err := s.loader.Load(manifestPath)
	if err != nil {
		ui.Log("-", "Failed to load manifest")
		ui.Result(false, time.Since(start))
		return err
	}

	profile, err := cfg.GetProfile(profileName)
	if err != nil || profile == nil {
		ui.Log("-", fmt.Sprintf("Profile '%s' not found", profileName))
		ui.Result(false, time.Since(start))
		return fmt.Errorf("profile not found")
	}

	env, ok := profile.Environments[envName]
	if !ok {
		ui.Log("-", fmt.Sprintf("Environment '%s' not found in profile '%s'", envName, profileName))
		ui.Result(false, time.Since(start))
		return fmt.Errorf("environment not found")
	}

	ui.Section("Deployment Info")
	ui.Log("*", fmt.Sprintf("Project: %s", cfg.Name))
	ui.Log("*", fmt.Sprintf("Version: %s", cfg.Version))
	ui.Log("*", fmt.Sprintf("Profile: %s (%s)", profileName, profile.Type))
	ui.Log("*", fmt.Sprintf("Target:  %s", envName))

	vars := s.resolveVariables(env)

	if err := s.handleBuild(profile, env, cfg.BaseDir, vars, start); err != nil {
		return err
	}

	if err := s.handleDeployment(profile, env, cfg, vars, allowProtected, verbose, start); err != nil {
		return err
	}

	ui.Result(true, time.Since(start))
	return nil
}

func (s *Service) RunSequence(manifestPath, sequenceName string, allowProtected, verbose bool) error {
	start := time.Now()

	ui.Log("*", fmt.Sprintf("Loading manifest: %s", manifestPath))
	absManifest, err := filepath.Abs(manifestPath)
	if err != nil {
		ui.Log("-", "Failed to resolve manifest path")
		ui.Result(false, time.Since(start))
		return err
	}

	manifestData, err := os.ReadFile(absManifest)
	if err != nil {
		ui.Log("-", "Failed to load manifest")
		ui.Result(false, time.Since(start))
		return err
	}

	diags, _, err := validate.ValidateYAML(manifestData, filepath.Dir(absManifest))
	if err != nil {
		ui.Log("-", "Failed to validate manifest")
		ui.Result(false, time.Since(start))
		return err
	}
	for _, d := range diags {
		if d.Severity == validate.SeverityError {
			ui.Log("-", validate.FormatDiagnostic(absManifest, d))
		}
	}
	if validate.HasErrors(diags) {
		ui.Result(false, time.Since(start))
		return fmt.Errorf("manifest validation failed")
	}

	cfg, err := s.loader.Load(manifestPath)
	if err != nil {
		ui.Log("-", "Failed to load manifest")
		ui.Result(false, time.Since(start))
		return err
	}

	steps, ok := cfg.Sequences[sequenceName]
	if !ok {
		ui.Log("-", fmt.Sprintf("Sequence '%s' not found", sequenceName))
		ui.Result(false, time.Since(start))
		return fmt.Errorf("sequence not found")
	}

	total := len(steps)
	ui.Section("Sequence")
	ui.Log("*", fmt.Sprintf("Project: %s", cfg.Name))
	ui.Log("*", fmt.Sprintf("Sequence: %s (%d steps)", sequenceName, total))

	for i, step := range steps {
		profileName, envName, err := target.Parse(step)
		if err != nil {
			ui.Log("-", fmt.Sprintf("Sequence step %d/%d: %v", i+1, total, err))
			ui.Result(false, time.Since(start))
			return err
		}

		ui.Log("*", fmt.Sprintf("Sequence step %d/%d: %s", i+1, total, step))
		if err := s.Run(manifestPath, profileName, envName, allowProtected, verbose); err != nil {
			ui.Log("-", fmt.Sprintf("Sequence aborted at step %d/%d", i+1, total))
			ui.Result(false, time.Since(start))
			return err
		}
	}

	ui.Result(true, time.Since(start))
	return nil
}

func (s *Service) resolveVariables(env domain.Environment) map[string]string {
	vars := make(map[string]string)
	for k, v := range env.EnvConfig.Variables {
		vars[k] = v
	}
	return vars
}

func (s *Service) handleBuild(profile *domain.Profile, env domain.Environment, baseDir string, vars map[string]string, start time.Time) error {
	buildConfig := profile.Build
	if env.Build != nil {
		buildConfig = env.Build
	}

	if buildConfig != nil && buildConfig.Command != "" {
		ui.Section("Phase 2: Build")
		ui.Log(">", fmt.Sprintf("Running build: %s", buildConfig.Command))

		path := buildConfig.Path
		if path == "" {
			path = baseDir
		} else if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, path)
		}

		if buildConfig.EnvFile != "" {
			envFilePath := filepath.Join(path, buildConfig.EnvFile)
			ui.Log("*", fmt.Sprintf("Writing variables to %s", buildConfig.EnvFile))
			if err := s.writeEnvFile(envFilePath, vars); err != nil {
				ui.Log("-", "Failed to write env file")
				ui.Result(false, time.Since(start))
				return err
			}
		}

		cmd := exec.Command("sh", "-c", buildConfig.Command)
		if strings.Contains(os.Getenv("OS"), "Windows") {
			cmd = exec.Command("cmd", "/C", buildConfig.Command)
		}
		cmd.Dir = path
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		for k, v := range vars {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}

		ui.ProgressBar(50, "Building")
		if err := cmd.Run(); err != nil {
			ui.Log("-", "Build failed")
			ui.Result(false, time.Since(start))
			return err
		}
		ui.ProgressBar(100, "Building")
		ui.Log("+", "Build completed")
	}
	return nil
}

func (s *Service) handleDeployment(profile *domain.Profile, env domain.Environment, cfg *domain.Config, vars map[string]string, allowProtected, verbose bool, start time.Time) error {
	isRemote := env.Remote != nil

	// 1. Pre-deployment Commands
	if len(env.Deploy.PreCommands) > 0 {
		ui.Section("Phase 3: Pre-Deployment Commands")
		if err := s.runCommands(env.Deploy.PreCommands, env, isRemote, cfg, vars); err != nil {
			ui.Log("-", "Pre-deployment commands failed")
			ui.Result(false, time.Since(start))
			return err
		}
	}

	// 2. Main Deployment
	ui.Section("Phase 4: Deployment")
	switch profile.Type {
	case "static", "binary":
		if isRemote {
			if err := s.deployRemoteSSH(profile, env, cfg, start, allowProtected, verbose, vars); err != nil {
				return err
			}
		} else {
			if err := s.deployLocal(profile, env, cfg.BaseDir, allowProtected, verbose, vars, start); err != nil {
				return err
			}
		}
	case "docker":
		if err := s.deployDocker(profile, env, cfg, cfg.BaseDir, start, vars); err != nil {
			return err
		}
	case "git-sync":
		if err := s.deployGitSync(profile, env, cfg, cfg.BaseDir, start, vars); err != nil {
			return err
		}
	}

	// 3. Post-deployment Commands
	if len(env.Deploy.PostCommands) > 0 {
		ui.Section("Phase 5: Post-Deployment Commands")
		if err := s.runCommands(env.Deploy.PostCommands, env, isRemote, cfg, vars); err != nil {
			ui.Log("-", "Post-deployment commands failed")
			ui.Result(false, time.Since(start))
			return err
		}
	}

	// 4. System Integration (Register PATH)
	if profile.Type == "binary" && env.RegisterPath != nil {
		if err := s.handlePathRegistration(env, cfg, isRemote); err != nil {
			ui.Log("!", fmt.Sprintf("Path registration warning: %v", err))
		}
	}

	return nil
}

func (s *Service) runCommands(commands []string, env domain.Environment, isRemote bool, cfg *domain.Config, vars map[string]string) error {
	if isRemote {
		sshClient, err := s.getSSHClient(env, cfg)
		if err != nil {
			return err
		}
		defer sshClient.Close()

		for _, cmd := range commands {
			ui.Log(">", cmd)
			var envPrefix strings.Builder
			for k, v := range vars {
				envPrefix.WriteString(fmt.Sprintf("%s='%s' ", k, v))
			}
			fullCmd := fmt.Sprintf("cd %s && %s%s", env.Deploy.TargetPath, envPrefix.String(), cmd)
			if _, err := s.deployer.ExecuteRemoteCommand(sshClient, fullCmd); err != nil {
				return err
			}
		}
	} else {
		for _, cmdStr := range commands {
			ui.Log(">", cmdStr)
			if err := hooks.Execute(cmdStr, "", vars); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) getSSHClient(env domain.Environment, cfg *domain.Config) (*ssh.Client, error) {
	if env.Remote == nil {
		return nil, fmt.Errorf("remote block is required for SSH operations")
	}

	sshHost := env.Remote.Host
	sshCredential := env.Remote.Credential
	hostKeyOpts := sshAdapter.HostKeyOptions{
		Verification:    env.Remote.HostKeyVerification,
		TrustOnFirstUse: env.Remote.TrustOnFirstUse,
	}

	cred, _ := cfg.GetCredential(sshCredential)
	if cred == nil {
		return nil, fmt.Errorf("credential %q not found", sshCredential)
	}

	if hostKeyOpts.VerificationDisabled() {
		ui.Log("!", "SSH host key verification is disabled for this environment (remote.host_key_verification: off)")
	}

	ui.Log("*", fmt.Sprintf("Connecting to %s as %s", sshHost, cred.Username))
	return s.deployer.ConnectSSH(sshHost, cred, hostKeyOpts)
}

func logArtifacts(files []string, artifactBase string, verbose bool) {
	ui.Log("+", fmt.Sprintf("Found %d artifact(s) to deploy", len(files)))
	if !verbose {
		return
	}
	for _, file := range files {
		rel, err := filepath.Rel(artifactBase, file)
		if err != nil {
			rel = file
		}
		ui.Log(" ", filepath.ToSlash(rel))
	}
}

func (s *Service) deployLocal(profile *domain.Profile, env domain.Environment, baseDir string, allowProtected, verbose bool, vars map[string]string, start time.Time) error {
	ui.Log("*", "Local deployment initiated.")
	artifactBase, include, exclude := s.resolveArtifacts(env, baseDir)
	if err := os.MkdirAll(artifactBase, 0o755); err != nil {
		ui.Log("-", "Failed to prepare artifact directory")
		ui.Result(false, time.Since(start))
		return fmt.Errorf("failed to prepare artifact directory: %w", err)
	}
	files, err := filter.GetFiles(artifactBase, include, exclude)
	if err != nil {
		ui.Log("-", "Filtering failed")
		ui.Result(false, time.Since(start))
		return err
	}
	logArtifacts(files, artifactBase, verbose)

	targetPath := s.resolvePath(baseDir, env.Deploy.TargetPath)
	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		ui.Log("-", "Failed to prepare deploy target directory")
		ui.Result(false, time.Since(start))
		return fmt.Errorf("failed to prepare deploy target directory: %w", err)
	}
	strategy := env.Deploy.Strategy
	if strategy == "" {
		strategy = "overwrite"
	}

	ui.Log(">", fmt.Sprintf("Deploying to %s (Strategy: %s)", targetPath, strategy))
	if err := s.deployer.Deploy(files, artifactBase, targetPath, strategy, allowProtected); err != nil {
		ui.Log("-", "Deployment failed")
		ui.Result(false, time.Since(start))
		return err
	}
	ui.Log("+", "Deployment successful")

	if env.EnvConfig.EnvFile != "" && len(vars) > 0 {
		ui.Log("*", fmt.Sprintf("Generating local env file: %s", env.EnvConfig.EnvFile))
		envFilePath := filepath.Join(targetPath, env.EnvConfig.EnvFile)
		if err := s.writeEnvFile(envFilePath, vars); err != nil {
			ui.Log("!", fmt.Sprintf("Failed to write local env file: %v", err))
			return fmt.Errorf("failed to write local env file: %w", err)
		}
		ui.Log("+", "Local env file generated")
	}

	if profile.Type != "docker" && len(vars) > 0 {
		ui.Log("*", "Applying template variables...")
		if err := template.ProcessFiles(targetPath, vars); err != nil {
			ui.Log("-", "Template processing failed")
			ui.Result(false, time.Since(start))
			return err
		}
		ui.Log("+", "Template processing completed")
	}

	return nil
}

func (s *Service) resolveArtifacts(env domain.Environment, baseDir string) (string, []string, []string) {
	artifactBase := baseDir
	var include, exclude []string

	if env.Deploy.Source != nil {
		if env.Deploy.Source.Dir != "" {
			artifactBase = s.resolvePath(baseDir, env.Deploy.Source.Dir)
		}
		include = env.Deploy.Source.Include
		exclude = env.Deploy.Source.Exclude
	}
	ui.Log("*", fmt.Sprintf("Artifact base: %s", artifactBase))
	ui.Log("*", "Filtering files...")
	return artifactBase, include, exclude
}

func (s *Service) writeEnvFile(path string, vars map[string]string) error {
	var sb strings.Builder
	for k, v := range vars {
		sb.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory for env file: %w", err)
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func (s *Service) handlePathRegistration(env domain.Environment, cfg *domain.Config, isRemote bool) error {
	targetPath := env.Deploy.TargetPath
	scope := env.RegisterPath.Scope
	if scope == "" {
		scope = "user"
	}

	ui.Log("*", fmt.Sprintf("Registering path in %s PATH: %s", strings.Title(scope), targetPath))

	if isRemote {
		// Simplified remote path registration (Bash specific)
		// This logic was previously in deployRemoteSSH, now centralized.
		commentTag := fmt.Sprintf("# Added by pablo for %s", cfg.Name)
		exportLine := fmt.Sprintf("export PATH=\"$PATH:%s\"", targetPath)

		shellFile := ".bashrc"
		if scope == "system" {
			shellFile = "/etc/profile.d/pablo.sh"
		}

		var targetFile string
		if strings.HasPrefix(shellFile, "/") {
			targetFile = shellFile
		} else {
			targetFile = fmt.Sprintf("~/%s", shellFile)
		}

		command := fmt.Sprintf("grep -q '%s' %s 2>/dev/null || echo -e '%s\\n%s' >> %s",
			commentTag, targetFile, commentTag, exportLine, targetFile)

		sshClient, err := s.getSSHClient(env, cfg)
		if err != nil {
			return fmt.Errorf("failed to connect for remote path registration: %w", err)
		}
		defer sshClient.Close()

		if output, err := s.deployer.ExecuteRemoteCommand(sshClient, command); err != nil {
			return fmt.Errorf("failed to register PATH remotely: %v (output: %s)", err, output)
		} else {
			ui.Log("+", "PATH registered on remote server")
		}
	} else {
		if err := system.AddPath(targetPath, scope, cfg.Name); err != nil {
			return fmt.Errorf("failed to register path locally: %w", err)
		} else {
			ui.Log("+", fmt.Sprintf("Path registered in %s scope successfully", scope))
		}
	}
	return nil
}

func (s *Service) deployRemoteSSH(profile *domain.Profile, env domain.Environment, cfg *domain.Config, start time.Time, allowProtected, verbose bool, vars map[string]string) error {
	ui.Log("*", "Remote SSH deployment initiated.")
	sshClient, err := s.getSSHClient(env, cfg)
	if err != nil {
		ui.Log("-", fmt.Sprintf("SSH connection failed: %v", err))
		ui.Result(false, time.Since(start))
		return err
	}
	defer sshClient.Close()
	ui.Log("+", "SSH connection established")

	artifactBase, include, exclude := s.resolveArtifacts(env, cfg.BaseDir)
	files, err := filter.GetFiles(artifactBase, include, exclude)
	if err != nil {
		ui.Log("-", "Filtering failed")
		ui.Result(false, time.Since(start))
		return err
	}
	logArtifacts(files, artifactBase, verbose)

	targetPath := env.Deploy.TargetPath
	strategy := env.Deploy.Strategy
	if strategy == "" {
		strategy = "overwrite"
	}

	ui.Log(">", fmt.Sprintf("Deploying to %s:%s (Strategy: %s)", env.Remote.Host, targetPath, strategy))
	if err := s.deployer.DeployRemote(files, artifactBase, sshClient, targetPath, strategy, allowProtected, env.Deploy.Transfer); err != nil {
		ui.Log("-", "Remote deployment failed")
		ui.Result(false, time.Since(start))
		return err
	}
	ui.Log("+", "Remote deployment successful")

	if env.Deploy.VerifyChecksum {
		ui.Log(">", "Verifying remote artifact checksums...")
		if err := s.deployer.VerifyRemoteChecksums(sshClient, files, artifactBase, targetPath); err != nil {
			ui.Log("-", "Checksum verification failed")
			ui.Result(false, time.Since(start))
			return err
		}
		ui.Log("+", "Checksum verification passed")
	}

	if env.EnvConfig.EnvFile != "" && len(vars) > 0 {
		ui.Log("*", fmt.Sprintf("Generating remote env file: %s", env.EnvConfig.EnvFile))
		if err := s.writeRemoteEnvFile(sshClient, targetPath, env.EnvConfig.EnvFile, vars); err != nil {
			ui.Log("!", fmt.Sprintf("Failed to write remote env file: %v", err))
			return err
		}
		ui.Log("+", "Remote env file generated")
	}

	return nil
}

func (s *Service) deployDocker(profile *domain.Profile, env domain.Environment, cfg *domain.Config, baseDir string, start time.Time, vars map[string]string) error {
	if env.Remote != nil {
		return s.deployDockerRemote(profile, env, cfg, start, vars)
	}

	ui.Log("*", "Docker deployment initiated.")
	if profile.Git == nil {
		return fmt.Errorf("git configuration required for docker deployment")
	}

	targetPath := env.Deploy.TargetPath
	ui.Log(">", fmt.Sprintf("Target: %s", targetPath))

	var composeFile string
	if env.Deploy.Docker != nil {
		composeFile = s.resolvePath(baseDir, env.Deploy.Docker.ComposeFile)
		if err := s.stopComposeBeforeSyncLocal(env.Deploy.Docker, composeFile, targetPath); err != nil {
			ui.Log("-", "Compose precheck failed")
			ui.Result(false, time.Since(start))
			return err
		}
	}

	if err := s.scm.CloneOrPull(profile.Git, targetPath); err != nil {
		ui.Log("-", "SCM operation failed")
		ui.Result(false, time.Since(start))
		return err
	}
	ui.Log("+", "SCM operation completed")

	if env.EnvConfig.EnvFile != "" && len(vars) > 0 {
		ui.Log("*", fmt.Sprintf("Generating env file: %s", env.EnvConfig.EnvFile))
		envFile := filepath.Join(targetPath, env.EnvConfig.EnvFile)
		if err := s.writeEnvFile(envFile, vars); err != nil {
			ui.Log("-", "Failed to write env file")
			ui.Result(false, time.Since(start))
			return err
		}
		ui.Log("+", "Env file generated")
	}

	if env.Deploy.Docker != nil {
		ui.Log(">", "Running Docker Compose...")
		build := env.Deploy.Docker.Build

		if err := s.docker.ComposeUp(composeFile, build, targetPath); err != nil {
			ui.Log("-", "Docker compose failed")
			ui.Result(false, time.Since(start))
			return err
		}
		ui.Log("+", "Docker compose started")
	}

	return nil
}

func (s *Service) deployDockerRemote(profile *domain.Profile, env domain.Environment, cfg *domain.Config, start time.Time, vars map[string]string) error {
	ui.Log("*", "Remote Docker deployment initiated.")
	if profile.Git == nil {
		return fmt.Errorf("git configuration required for docker deployment")
	}

	sshClient, err := s.getSSHClient(env, cfg)
	if err != nil {
		ui.Log("-", fmt.Sprintf("SSH connection failed: %v", err))
		ui.Result(false, time.Since(start))
		return err
	}
	defer sshClient.Close()
	ui.Log("+", "SSH connection established")

	targetPath := env.Deploy.TargetPath
	ui.Log(">", fmt.Sprintf("Target: %s:%s", env.Remote.Host, targetPath))

	var composeFile string
	if env.Deploy.Docker != nil {
		composeFile = env.Deploy.Docker.ComposeFile
		if !strings.HasPrefix(composeFile, "/") {
			composeFile = pathutil.JoinRemote(targetPath, composeFile)
		}
		if err := s.stopComposeBeforeSyncRemote(sshClient, env.Deploy.Docker, composeFile, targetPath); err != nil {
			ui.Log("-", "Compose precheck failed")
			ui.Result(false, time.Since(start))
			return err
		}
	}

	if err := s.syncRepoRemote(sshClient, profile.Git, targetPath); err != nil {
		ui.Log("-", err.Error())
		ui.Result(false, time.Since(start))
		return err
	}

	if env.EnvConfig.EnvFile != "" && len(vars) > 0 {
		if err := s.writeRemoteEnvFile(sshClient, targetPath, env.EnvConfig.EnvFile, vars); err != nil {
			ui.Log("!", fmt.Sprintf("Failed to write remote env file: %v", err))
			return err
		}
		ui.Log("+", "Remote env file generated")
	}

	if env.Deploy.Docker != nil {
		ui.Log(">", "Running Docker Compose on remote...")

		buildFlag := ""
		if env.Deploy.Docker.Build {
			buildFlag = " --build"
		}

		composeCmd := fmt.Sprintf("cd %s && docker compose -f %s up -d%s", targetPath, composeFile, buildFlag)
		out, err := s.deployer.ExecuteRemoteCommand(sshClient, composeCmd)
		if err != nil {
			ui.Log("-", "Remote docker compose failed")
			logRemoteCommandOutput(out)
			ui.Result(false, time.Since(start))
			return fmt.Errorf("remote docker compose failed: %w%s", err, formatRemoteCommandOutput(out))
		}
		ui.Log("+", "Docker compose started on remote")
	}

	return nil
}

// stopComposeBeforeSyncLocal stops a running local Compose stack before git sync when enabled.
func (s *Service) stopComposeBeforeSyncLocal(dockerCfg *domain.DockerConfig, composeFile, targetPath string) error {
	if dockerCfg == nil || !dockerCfg.StopBeforeSyncEnabled() {
		return nil
	}
	if _, err := os.Stat(targetPath); err != nil {
		return nil
	}
	effectiveComposeFile := resolveExistingComposeFile(composeFile, targetPath, dockerCfg.ComposeFile)
	if effectiveComposeFile == "" {
		return nil
	}

	running, err := s.docker.ComposePsRunning(effectiveComposeFile, targetPath)
	if err != nil {
		ui.Log("!", fmt.Sprintf("Compose precheck skipped: %v", err))
		return nil
	}
	if !running {
		return nil
	}

	ui.Log(">", "Compose stack is running; stopping before git sync...")
	if err := s.docker.ComposeDown(effectiveComposeFile, targetPath); err != nil {
		return fmt.Errorf("compose down before sync failed: %w", err)
	}
	ui.Log("+", "Compose stack stopped")
	return nil
}

// stopComposeBeforeSyncRemote stops a running remote Compose stack before git sync when enabled.
func (s *Service) stopComposeBeforeSyncRemote(sshClient *ssh.Client, dockerCfg *domain.DockerConfig, composeFile, targetPath string) error {
	if dockerCfg == nil || !dockerCfg.StopBeforeSyncEnabled() {
		return nil
	}

	checkDirCmd := fmt.Sprintf("test -d %s && echo exists || echo missing", targetPath)
	dirOut, _ := s.deployer.ExecuteRemoteCommand(sshClient, checkDirCmd)
	if !strings.Contains(dirOut, "exists") {
		return nil
	}

	checkFileCmd := fmt.Sprintf("test -f %s && echo exists || echo missing", composeFile)
	fileOut, _ := s.deployer.ExecuteRemoteCommand(sshClient, checkFileCmd)
	if !strings.Contains(fileOut, "exists") {
		return nil
	}

	psCmd := fmt.Sprintf("cd %s && docker compose -f %s ps -q", targetPath, composeFile)
	psOut, err := s.deployer.ExecuteRemoteCommand(sshClient, psCmd)
	if err != nil {
		ui.Log("!", fmt.Sprintf("Compose precheck skipped: %v", err))
		return nil
	}
	if strings.TrimSpace(psOut) == "" {
		return nil
	}

	ui.Log(">", "Compose stack is running on remote; stopping before git sync...")
	downCmd := fmt.Sprintf("cd %s && docker compose -f %s down", targetPath, composeFile)
	out, err := s.deployer.ExecuteRemoteCommand(sshClient, downCmd)
	if err != nil {
		logRemoteCommandOutput(out)
		return fmt.Errorf("remote compose down before sync failed: %w%s", err, formatRemoteCommandOutput(out))
	}
	ui.Log("+", "Remote compose stack stopped")
	return nil
}

func logRemoteCommandOutput(out string) {
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return
	}
	for _, line := range strings.Split(trimmed, "\n") {
		ui.Log("!", line)
	}
}

func formatRemoteCommandOutput(out string) string {
	trimmed := strings.TrimSpace(out)
	if trimmed == "" {
		return ""
	}
	return "\n" + trimmed
}

func resolveExistingComposeFile(resolvedComposeFile, targetPath, configuredComposeFile string) string {
	if _, err := os.Stat(resolvedComposeFile); err == nil {
		return resolvedComposeFile
	}
	if configuredComposeFile == "" {
		return ""
	}
	candidate := configuredComposeFile
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(targetPath, configuredComposeFile)
	}
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return ""
}

// syncRepoRemote clones or pulls the configured git repository into targetPath
// on the remote host over an existing SSH connection.
func (s *Service) syncRepoRemote(sshClient *ssh.Client, git *domain.GitConfig, targetPath string) error {
	ui.Log("*", "Marking remote directory as safe in Git configuration")
	safeCmd := fmt.Sprintf("git config --global --add safe.directory %s", targetPath)
	s.deployer.ExecuteRemoteCommand(sshClient, safeCmd) // Ignore error, might not exist yet

	checkCmd := fmt.Sprintf("test -d %s/.git && echo 'exists' || echo 'not found'", targetPath)
	output, _ := s.deployer.ExecuteRemoteCommand(sshClient, checkCmd)

	if strings.Contains(output, "exists") {
		ui.Log(">", "Git Pulling...")
		pullCmd := fmt.Sprintf("cd %s && git pull origin %s", targetPath, git.Branch)
		if _, err := s.deployer.ExecuteRemoteCommand(sshClient, pullCmd); err != nil {
			return fmt.Errorf("remote git pull failed: %w", err)
		}
	} else {
		ui.Log(">", "Git Cloning...")
		parentDir := pathutil.DirRemote(targetPath)
		ensureDirCmd := fmt.Sprintf("mkdir -p %s", parentDir)
		s.deployer.ExecuteRemoteCommand(sshClient, ensureDirCmd) // Ensure parent dir exists

		cloneCmd := fmt.Sprintf("git clone -b %s %s %s", git.Branch, git.Repo, targetPath)
		if _, err := s.deployer.ExecuteRemoteCommand(sshClient, cloneCmd); err != nil {
			return fmt.Errorf("remote git clone failed: %w", err)
		}
	}
	ui.Log("+", "Remote code synced")
	return nil
}

// writeRemoteEnvFile writes key/value variables to an env file on the remote
// host, creating the parent directory first.
func (s *Service) writeRemoteEnvFile(sshClient *ssh.Client, targetPath, envFile string, vars map[string]string) error {
	var content strings.Builder
	for k, v := range vars {
		content.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}

	remoteEnvPath := pathutil.JoinRemote(targetPath, envFile)
	ensureDirCmd := fmt.Sprintf("mkdir -p %s", pathutil.DirRemote(remoteEnvPath))
	if _, err := s.deployer.ExecuteRemoteCommand(sshClient, ensureDirCmd); err != nil {
		return fmt.Errorf("failed to create remote directory for env file: %w", err)
	}

	cmd := fmt.Sprintf("cat << 'EOF' > %s\n%sEOF", remoteEnvPath, content.String())
	if _, err := s.deployer.ExecuteRemoteCommand(sshClient, cmd); err != nil {
		return fmt.Errorf("failed to write remote env file: %w", err)
	}
	return nil
}

func (s *Service) deployGitSync(profile *domain.Profile, env domain.Environment, cfg *domain.Config, baseDir string, start time.Time, vars map[string]string) error {
	ui.Log("*", "Git Sync deployment initiated.")
	if env.Remote != nil {
		return s.deployGitSyncRemote(profile, env, cfg, start, vars)
	}

	if profile.Git == nil {
		return fmt.Errorf("git configuration required")
	}

	targetPath := env.Deploy.TargetPath
	ui.Log(">", fmt.Sprintf("Syncing to: %s", targetPath))

	if err := s.scm.CloneOrPull(profile.Git, targetPath); err != nil {
		ui.Log("-", "Git sync failed")
		ui.Result(false, time.Since(start))
		return err
	}
	ui.Log("+", "Git sync completed")

	if env.EnvConfig.EnvFile != "" && len(vars) > 0 {
		ui.Log("*", fmt.Sprintf("Generating local env file: %s", env.EnvConfig.EnvFile))
		envFilePath := filepath.Join(targetPath, env.EnvConfig.EnvFile)
		if err := s.writeEnvFile(envFilePath, vars); err != nil {
			ui.Log("!", fmt.Sprintf("Failed to write env file: %v", err))
			return fmt.Errorf("failed to write env file: %w", err)
		}
		ui.Log("+", "Env file generated")
	}

	return nil
}

func (s *Service) deployGitSyncRemote(profile *domain.Profile, env domain.Environment, cfg *domain.Config, start time.Time, vars map[string]string) error {
	ui.Log("*", "Remote Git Sync deployment initiated.")
	if profile.Git == nil {
		return fmt.Errorf("git configuration required")
	}

	sshClient, err := s.getSSHClient(env, cfg)
	if err != nil {
		ui.Log("-", fmt.Sprintf("SSH connection failed: %v", err))
		ui.Result(false, time.Since(start))
		return err
	}
	defer sshClient.Close()
	ui.Log("+", "SSH connection established")

	targetPath := env.Deploy.TargetPath

	if err := s.syncRepoRemote(sshClient, profile.Git, targetPath); err != nil {
		ui.Log("-", err.Error())
		ui.Result(false, time.Since(start))
		return err
	}

	if env.EnvConfig.EnvFile != "" && len(vars) > 0 {
		ui.Log("*", fmt.Sprintf("Generating remote env file: %s", env.EnvConfig.EnvFile))
		if err := s.writeRemoteEnvFile(sshClient, targetPath, env.EnvConfig.EnvFile, vars); err != nil {
			ui.Log("!", fmt.Sprintf("Failed to write remote env file: %v", err))
			return err
		}
		ui.Log("+", "Remote env file generated")
	}

	return nil
}

func (s *Service) Uninstall(manifestPath, profileName, envName string, removeBackups bool) error {
	start := time.Now()
	cfg, err := s.loader.Load(manifestPath)
	if err != nil {
		return err
	}

	profile, _ := cfg.GetProfile(profileName)
	env, ok := profile.Environments[envName]
	if !ok {
		return fmt.Errorf("environment not found")
	}

	targetPath := s.resolvePath(cfg.BaseDir, env.Deploy.TargetPath)
	ui.Section("Uninstall Info")
	ui.Log("*", fmt.Sprintf("Project: %s", cfg.Name))
	ui.Log("*", fmt.Sprintf("Target Path: %s", targetPath))

	if _, err := os.Stat(targetPath); err == nil {
		ui.Log(">", "Removing files...")
		os.RemoveAll(targetPath)
	}

	if profile.Type == "binary" && env.RegisterPath != nil {
		scope := env.RegisterPath.Scope
		if scope == "" {
			scope = "user"
		}
		system.RemovePath(cfg.Name, env.Deploy.TargetPath, scope)
	}

	if removeBackups {
		ui.Section("Remove Backups")
		pattern := filepath.Join(filepath.Dir(targetPath), filepath.Base(targetPath)+"_backup_*")
		matches, _ := filepath.Glob(pattern)
		for _, m := range matches {
			ui.Log(">", "Removing backup: "+filepath.Base(m))
			os.RemoveAll(m)
		}
	}

	ui.Result(true, time.Since(start))
	return nil
}
