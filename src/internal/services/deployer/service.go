package deployer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	sshAdapter "pablo/internal/adapters/ssh"
	"pablo/internal/domain"

	"golang.org/x/crypto/ssh"
)

type Service struct {
	ssh *sshAdapter.Adapter
}

func New() *Service {
	return &Service{
		ssh: sshAdapter.New(),
	}
}

func (s *Service) Deploy(files []string, sourceBase, targetPath string, strategy string, allowProtected bool) error {
	// Normalize target path
	targetPath = filepath.Clean(targetPath)

	if (strategy == "recreate" || strategy == "backup") && !allowProtected {
		if s.isProtectedPath(targetPath) {
			return fmt.Errorf("safety break: target path '%s' is a protected system directory (use --force to override)", targetPath)
		}
	}

	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return fmt.Errorf("critical: cannot create or access target directory %s: %w", targetPath, err)
	}

	switch strategy {
	case "backup":
		if err := s.backup(targetPath); err != nil {
			return err
		}
	case "recreate":
		// ui is not available here, so we just log to stdout or return error
		// Note: we already checked isProtectedPath above
		if err := os.RemoveAll(targetPath); err != nil {
			return fmt.Errorf("failed to clean target directory: %w", err)
		}
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			return fmt.Errorf("failed to recreate target directory: %w", err)
		}
	case "blue-green":
		return fmt.Errorf("blue-green strategy not yet fully implemented")
	}

	for _, file := range files {
		rel, err := filepath.Rel(sourceBase, file)
		if err != nil {
			return err
		}

		dest := filepath.Join(targetPath, rel)
		if err := s.copyFile(file, dest); err != nil {
			return err
		}
	}

	return nil
}

// DeployRemote deploys files to a remote server via SSH
func (s *Service) DeployRemote(files []string, sourceBase string, sshClient *ssh.Client, targetPath string, strategy string, allowProtected bool, remoteTransfer string) error {
	// Default to tar if not specified or specified as tar
	if remoteTransfer == "" {
		remoteTransfer = "tar"
	}

	// Basic remote path safety check
	if (strategy == "recreate" || strategy == "backup") && !allowProtected {
		if s.ssh != nil && s.isProtectedPath(targetPath) {
			return fmt.Errorf("safety break: remote target path '%s' appears to be a protected system directory (use --force to override)", targetPath)
		}
	}

	// Handle strategies that require preparation
	switch strategy {
	case "backup":
		if err := s.ssh.CreateBackup(sshClient, targetPath); err != nil {
			return fmt.Errorf("failed to create remote backup: %w", err)
		}
	case "recreate":
		// Note: we already checked isProtectedPath above
		// Use rm -rf to clean the directory contents or the directory itself
		if _, err := s.ssh.ExecuteCommand(sshClient, fmt.Sprintf("rm -rf %s && mkdir -p %s", targetPath, targetPath)); err != nil {
			return fmt.Errorf("failed to clean remote directory: %w", err)
		}
	}

	// Ensure remote directory exists
	if _, err := s.ssh.ExecuteCommand(sshClient, fmt.Sprintf("mkdir -p %s", targetPath)); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Transfer strategy
	if remoteTransfer == "tar" {
		// High performance bulk transfer
		if err := s.ssh.TransferPipeline(sshClient, files, sourceBase, targetPath); err != nil {
			return fmt.Errorf("batch transfer failed: %w", err)
		}
	} else {
		// Legacy one-by-one transfer
		for _, file := range files {
			rel, err := filepath.Rel(sourceBase, file)
			if err != nil {
				return err
			}

			remoteDest := filepath.Join(targetPath, rel)
			if err := s.ssh.TransferFile(sshClient, file, remoteDest); err != nil {
				return fmt.Errorf("failed to transfer %s: %w", file, err)
			}

			// Set executable permissions for binary files
			// In tar mode, permissions are preserved by tar itself
			if _, err := s.ssh.ExecuteCommand(sshClient, fmt.Sprintf("chmod +x %s", remoteDest)); err != nil {
				return fmt.Errorf("failed to set permissions: %w", err)
			}
		}
	}

	return nil
}

func (s *Service) isProtectedPath(path string) bool {
	clean := filepath.Clean(path)

	// Block root and shallow paths
	if clean == "/" || clean == "." || clean == ".." || clean == "\\" {
		return true
	}

	// Top-level system directories (Unix)
	protected := []string{
		"/bin", "/boot", "/dev", "/etc", "/home", "/lib", "/lib64",
		"/media", "/mnt", "/opt", "/proc", "/root", "/run", "/sbin",
		"/srv", "/sys", "/tmp", "/usr", "/var",
	}

	for _, p := range protected {
		if clean == p {
			return true
		}
	}

	// Windows critical paths (basic check)
	lower := strings.ToLower(clean)
	if lower == "c:" || lower == "c:\\" || lower == "c:\\windows" || lower == "c:\\program files" {
		return true
	}

	return false
}

func (s *Service) backup(targetPath string) error {
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return nil
	}

	backupName := fmt.Sprintf("%s_backup_%s", targetPath, time.Now().Format("20060102_150405"))
	fmt.Printf("Backing up %s to %s\n", targetPath, backupName)
	return os.Rename(targetPath, backupName)
}

func (s *Service) copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err = io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// ConnectSSH establishes an SSH connection using credentials
func (s *Service) ConnectSSH(host string, cred *domain.CredentialConfig) (*ssh.Client, error) {
	return s.ssh.Connect(host, cred)
}

// ExecuteRemoteCommand runs a command on the remote server
func (s *Service) ExecuteRemoteCommand(client *ssh.Client, command string) (string, error) {
	return s.ssh.ExecuteCommand(client, command)
}
