package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"pablo/internal/adapters/docker"
	"pablo/internal/adapters/system"
	"pablo/internal/config"
	"pablo/internal/domain"
	"pablo/internal/services/builder"
	"pablo/internal/services/deployer"
	"pablo/internal/services/filter"
	"pablo/internal/services/health"
	"pablo/internal/services/hooks"
	"pablo/internal/services/scm"
	"pablo/internal/services/template"
	"pablo/pkg/ui"

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

func (s *Service) Run(manifestPath, profileName, envName string, allowProtected bool) error {
	start := time.Now()

	ui.Log("*", fmt.Sprintf("Loading manifest: %s", manifestPath))
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

	// 1. Resolve Variables
	vars := s.resolveVariables(profile, env)

	// 2. Pre-Deploy Hooks (Global)
	if profile.Hooks.Pre != "" {
		ui.Section("Phase 1: Pre-Deploy Hooks")
		ui.Log(">", fmt.Sprintf("Executing: %s", profile.Hooks.Pre))
		if err := hooks.Execute(profile.Hooks.Pre, "", vars); err != nil {
			ui.Log("-", "Pre-deploy hook failed")
			ui.Result(false, time.Since(start))
			return err
		}
	}

	// 3. Build Phase
	if err := s.handleBuild(profile, env, cfg.BaseDir, vars, start); err != nil {
		return err
	}

	// 4. Deployment Phase
	if err := s.handleDeployment(profile, env, cfg, vars, allowProtected, start); err != nil {
		return err
	}

	// 5. Post-Deploy Hooks (Global)
	if profile.Hooks.Post != "" {
		ui.Section("Phase 6: Post-Deploy Hooks")
		ui.Log(">", fmt.Sprintf("Executing: %s", profile.Hooks.Post))
		if err := hooks.Execute(profile.Hooks.Post, "", vars); err != nil {
			if profile.Pipeline.OnFailure != "" {
				hooks.Execute(profile.Pipeline.OnFailure, "", vars)
			}
			ui.Log("-", "Post-deploy hook failed")
			ui.Result(false, time.Since(start))
			return err
		}
	}

	// 6. Health Check
	if profile.Pipeline.HealthCheck != "" {
		ui.Section("Health Check")
		ui.Log(">", fmt.Sprintf("Verifying: %s", profile.Pipeline.HealthCheck))
		if err := health.Check(profile.Pipeline.HealthCheck, 30*time.Second); err != nil {
			if profile.Pipeline.OnFailure != "" {
				hooks.Execute(profile.Pipeline.OnFailure, "", vars)
			}
			ui.Log("-", "Health check failed")
			ui.Result(false, time.Since(start))
			return err
		}
		ui.Log("+", "Health check passed")
	}

	if profile.Pipeline.OnSuccess != "" {
		hooks.Execute(profile.Pipeline.OnSuccess, "", vars)
	}

	ui.Result(true, time.Since(start))
	return nil
}

func (s *Service) resolveVariables(profile *domain.Profile, env domain.Environment) map[string]string {
	vars := make(map[string]string)
	// Merging logic is already partially handled by loader, but we ensure consistency here
	if env.Deploy.Variables != nil {
		for k, v := range env.Deploy.Variables {
			vars[k] = v
		}
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

func (s *Service) handleDeployment(profile *domain.Profile, env domain.Environment, cfg *domain.Config, vars map[string]string, allowProtected bool, start time.Time) error {
	isRemote := env.Remote != nil && env.Remote.Method == "ssh"

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
			if err := s.deployRemoteSSH(profile, env, cfg, start, allowProtected, vars); err != nil {
				return err
			}
		} else {
			if err := s.deployLocal(profile, env, cfg.BaseDir, allowProtected, vars, start); err != nil {
				return err
			}
		}
	case "docker":
		if err := s.deployDocker(profile, env, cfg.BaseDir, start, vars); err != nil {
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
	var sshHost, sshCredential string
	if env.Remote != nil {
		sshHost = env.Remote.Host
		sshCredential = env.Remote.Credential
	} else if env.Deploy.SSH != nil { // Fallback for older configs
		sshHost = env.Deploy.SSH.Host
		sshCredential = env.Deploy.SSH.Credential
	} else { // Another fallback
		sshHost = env.Deploy.Host
		sshCredential = env.Deploy.Credential
	}

	var cred *domain.CredentialConfig
	if sshCredential != "" {
		c, _ := cfg.GetCredential(sshCredential)
		cred = c
	}

	if cred == nil {
		ui.Log("*", "No credential specified, using default (root@~/.ssh/id_rsa)")
		cred = &domain.CredentialConfig{Type: "ssh", Username: "root", Key: "~/.ssh/id_rsa"}
	}

	ui.Log("*", fmt.Sprintf("Connecting to %s as %s", sshHost, cred.Username))
	return s.deployer.ConnectSSH(sshHost, cred)
}

func (s *Service) deployLocal(profile *domain.Profile, env domain.Environment, baseDir string, allowProtected bool, vars map[string]string, start time.Time) error {
	ui.Log("*", "Local deployment initiated.")
	artifactBase, include, exclude := s.resolveArtifacts(profile, env, baseDir)
	files, err := filter.GetFiles(artifactBase, include, exclude)
	if err != nil {
		ui.Log("-", "Filtering failed")
		ui.Result(false, time.Since(start))
		return err
	}
	ui.Log("+", fmt.Sprintf("Found %d artifact(s) to deploy", len(files)))

	targetPath := s.resolvePath(baseDir, env.Deploy.TargetPath)
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

	if env.Deploy.EnvFile != "" && len(vars) > 0 {
		ui.Log("*", fmt.Sprintf("Generating local env file: %s", env.Deploy.EnvFile))
		envFilePath := filepath.Join(targetPath, env.Deploy.EnvFile)
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

func (s *Service) resolveArtifacts(profile *domain.Profile, env domain.Environment, baseDir string) (string, []string, []string) {
	artifactBase := baseDir
	include := profile.OutputDir.Include
	exclude := profile.OutputDir.Exclude

	if profile.OutputDir.Dir != "" {
		artifactBase = s.resolvePath(baseDir, profile.OutputDir.Dir)
	}

	if env.Deploy.Source != nil {
		if env.Deploy.Source.Dir != "" {
			artifactBase = s.resolvePath(baseDir, env.Deploy.Source.Dir)
		}
		if len(env.Deploy.Source.Include) > 0 {
			include = env.Deploy.Source.Include
		}
		if len(env.Deploy.Source.Exclude) > 0 {
			exclude = env.Deploy.Source.Exclude
		}
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

func (s *Service) deployRemoteSSH(profile *domain.Profile, env domain.Environment, cfg *domain.Config, start time.Time, allowProtected bool, vars map[string]string) error {
	ui.Log("*", "Remote SSH deployment initiated.")
	sshClient, err := s.getSSHClient(env, cfg)
	if err != nil {
		ui.Log("-", fmt.Sprintf("SSH connection failed: %v", err))
		ui.Result(false, time.Since(start))
		return err
	}
	defer sshClient.Close()
	ui.Log("+", "SSH connection established")

	artifactBase, include, exclude := s.resolveArtifacts(profile, env, cfg.BaseDir)
	files, err := filter.GetFiles(artifactBase, include, exclude)
	if err != nil {
		ui.Log("-", "Filtering failed")
		ui.Result(false, time.Since(start))
		return err
	}
	ui.Log("+", fmt.Sprintf("Found %d artifact(s) to deploy", len(files)))

	targetPath := env.Deploy.TargetPath
	strategy := env.Deploy.Strategy
	if strategy == "" {
		strategy = "overwrite"
	}

	ui.Log(">", fmt.Sprintf("Deploying to %s:%s (Strategy: %s)", env.Remote.Host, targetPath, strategy))
	if err := s.deployer.DeployRemote(files, artifactBase, sshClient, targetPath, strategy, allowProtected, env.Deploy.Remote); err != nil {
		ui.Log("-", "Remote deployment failed")
		ui.Result(false, time.Since(start))
		return err
	}
	ui.Log("+", "Remote deployment successful")

	if env.Deploy.EnvFile != "" && len(vars) > 0 {
		ui.Log("*", fmt.Sprintf("Generating remote env file: %s", env.Deploy.EnvFile))
		remoteEnvPath := filepath.Join(targetPath, env.Deploy.EnvFile)
		var sb strings.Builder
		for k, v := range vars {
			sb.WriteString(fmt.Sprintf("%s=%s\n", k, v))
		}
		// Ensure remote directory exists before writing
		cmd := fmt.Sprintf("mkdir -p %s && cat << 'EOF' > %s\n%sEOF", filepath.Dir(remoteEnvPath), remoteEnvPath, sb.String())
		if _, err := s.deployer.ExecuteRemoteCommand(sshClient, cmd); err != nil {
			ui.Log("!", fmt.Sprintf("Failed to write remote env file: %v", err))
			return fmt.Errorf("failed to write remote env file: %w", err)
		}
		ui.Log("+", "Remote env file generated")
	}

	return nil
}

func (s *Service) deployDocker(profile *domain.Profile, env domain.Environment, baseDir string, start time.Time, vars map[string]string) error {
	ui.Log("*", "Docker deployment initiated.")
	if profile.Git == nil {
		return fmt.Errorf("git configuration required for docker deployment")
	}

	targetPath := env.Deploy.TargetPath
	ui.Log(">", fmt.Sprintf("Target: %s", targetPath))

	if err := s.scm.CloneOrPull(profile.Git, targetPath); err != nil {
		ui.Log("-", "SCM operation failed")
		ui.Result(false, time.Since(start))
		return err
	}
	ui.Log("+", "SCM operation completed")

	if env.Deploy.EnvFile != "" && len(vars) > 0 {
		ui.Log("*", fmt.Sprintf("Generating env file: %s", env.Deploy.EnvFile))
		envFile := filepath.Join(targetPath, env.Deploy.EnvFile)
		if err := s.writeEnvFile(envFile, vars); err != nil {
			ui.Log("-", "Failed to write env file")
			ui.Result(false, time.Since(start))
			return err
		}
		ui.Log("+", "Env file generated")
	}

	if env.Deploy.Docker != nil {
		ui.Log(">", "Running Docker Compose...")
		composeFile := s.resolvePath(baseDir, env.Deploy.Docker.ComposeFile)
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

func (s *Service) deployGitSync(profile *domain.Profile, env domain.Environment, cfg *domain.Config, baseDir string, start time.Time, vars map[string]string) error {
	ui.Log("*", "Git Sync deployment initiated.")
	isRemote := false
	if env.Remote != nil && env.Remote.Method == "ssh" {
		isRemote = true
	}

	if isRemote {
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

	if env.Deploy.EnvFile != "" && len(vars) > 0 {
		ui.Log("*", fmt.Sprintf("Generating local env file: %s", env.Deploy.EnvFile))
		envFilePath := filepath.Join(targetPath, env.Deploy.EnvFile)
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

	// Mark remote directory as safe (Fixes dubious ownership on remote)
	ui.Log("*", "Marking remote directory as safe in Git configuration")
	safeCmd := fmt.Sprintf("git config --global --add safe.directory %s", targetPath)
	s.deployer.ExecuteRemoteCommand(sshClient, safeCmd) // Ignore error, might not exist

	// Perform Git Sync on Remote
	// Check if directory exists and is a git repo
	checkCmd := fmt.Sprintf("test -d %s/.git && echo 'exists' || echo 'not found'", targetPath)
	output, _ := s.deployer.ExecuteRemoteCommand(sshClient, checkCmd)

	if strings.Contains(output, "exists") {
		ui.Log(">", "Git Pulling...")
		pullCmd := fmt.Sprintf("cd %s && git pull origin %s", targetPath, profile.Git.Branch)
		if _, err := s.deployer.ExecuteRemoteCommand(sshClient, pullCmd); err != nil {
			ui.Log("-", fmt.Sprintf("Remote git pull failed: %v", err))
			ui.Result(false, time.Since(start))
			return fmt.Errorf("remote git pull failed: %w", err)
		}
	} else {
		ui.Log(">", "Git Cloning...")
		parentDir := filepath.Dir(targetPath)
		ensureDirCmd := fmt.Sprintf("mkdir -p %s", parentDir)
		s.deployer.ExecuteRemoteCommand(sshClient, ensureDirCmd) // Ensure parent dir exists

		cloneCmd := fmt.Sprintf("git clone -b %s %s %s", profile.Git.Branch, profile.Git.Repo, targetPath)
		if _, err := s.deployer.ExecuteRemoteCommand(sshClient, cloneCmd); err != nil {
			ui.Log("-", fmt.Sprintf("Remote git clone failed: %v", err))
			ui.Result(false, time.Since(start))
			return fmt.Errorf("remote git clone failed: %w", err)
		}
	}
	ui.Log("+", "Remote code synced")

	if env.Deploy.EnvFile != "" && len(vars) > 0 {
		ui.Log("*", fmt.Sprintf("Generating remote env file: %s", env.Deploy.EnvFile))

		var envContent strings.Builder
		for k, v := range vars {
			envContent.WriteString(fmt.Sprintf("%s=%s\n", k, v))
		}

		remoteEnvPath := filepath.Join(targetPath, env.Deploy.EnvFile)

		ensureDirCmd := fmt.Sprintf("mkdir -p %s", filepath.Dir(remoteEnvPath))
		if _, err := s.deployer.ExecuteRemoteCommand(sshClient, ensureDirCmd); err != nil {
			ui.Log("!", fmt.Sprintf("Failed to create remote directory for env file: %v", err))
			return fmt.Errorf("failed to create remote directory for env file: %w", err)
		}

		cmd := fmt.Sprintf("cat << 'EOF' > %s\n%sEOF", remoteEnvPath, envContent.String())
		if _, err := s.deployer.ExecuteRemoteCommand(sshClient, cmd); err != nil {
			ui.Log("!", fmt.Sprintf("Failed to write remote env file: %v", err))
			return fmt.Errorf("failed to write remote env file: %w", err)
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
		system.RemovePath(cfg.Name)
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
