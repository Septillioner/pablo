package ssh

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"pablo/internal/domain"

	"golang.org/x/crypto/ssh"
)

type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

// Connect establishes an SSH connection using the provided credentials
func (a *Adapter) Connect(host string, cred *domain.CredentialConfig) (*ssh.Client, error) {
	var authMethod ssh.AuthMethod

	switch cred.Type {
	case "ssh":
		if cred.Key != "" {
			// SSH key authentication
			keyPath := expandPath(cred.Key)
			key, err := os.ReadFile(keyPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read SSH key: %w", err)
			}

			var signer ssh.Signer
			if cred.Passphrase != "" {
				signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(cred.Passphrase))
			} else {
				signer, err = ssh.ParsePrivateKey(key)
			}
			if err != nil {
				return nil, fmt.Errorf("failed to parse SSH key: %w", err)
			}
			authMethod = ssh.PublicKeys(signer)
		} else if cred.Password != "" {
			// Password authentication
			authMethod = ssh.Password(cred.Password)
		} else {
			return nil, fmt.Errorf("SSH credential must have either key or password")
		}
	default:
		return nil, fmt.Errorf("unsupported credential type: %s", cred.Type)
	}

	config := &ssh.ClientConfig{
		User:            cred.Username,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Add proper host key verification
	}

	// Add default port if not specified
	if !strings.Contains(host, ":") {
		host = host + ":22"
	}

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", host, err)
	}

	return client, nil
}

// TransferFile transfers a single file to the remote server using SCP
func (a *Adapter) TransferFile(client *ssh.Client, localPath, remotePath string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	stat, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Start SCP command on remote
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C%04o %d %s\n", stat.Mode().Perm(), stat.Size(), filepath.Base(remotePath))
		io.Copy(w, localFile)
		fmt.Fprint(w, "\x00")
	}()

	remoteDir := filepath.Dir(remotePath)
	if err := session.Run(fmt.Sprintf("mkdir -p %s && scp -t %s", remoteDir, remotePath)); err != nil {
		return fmt.Errorf("failed to transfer file: %w", err)
	}

	return nil
}

// TransferPipeline transfers multiple files using a single tar stream over SSH
func (a *Adapter) TransferPipeline(client *ssh.Client, files []string, sourceBase, remotePath string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	w, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Start remote tar command
	if err := session.Start(fmt.Sprintf("mkdir -p %s && tar -xf - -C %s", remotePath, remotePath)); err != nil {
		w.Close()
		return fmt.Errorf("failed to start remote tar: %w", err)
	}

	tw := tar.NewWriter(w)
	for _, file := range files {
		if err := a.addToTar(tw, file, sourceBase); err != nil {
			tw.Close()
			w.Close()
			return err
		}
	}

	if err := tw.Close(); err != nil {
		w.Close()
		return fmt.Errorf("failed to close tar writer: %w", err)
	}
	w.Close()

	if err := session.Wait(); err != nil {
		return fmt.Errorf("remote tar failed: %w", err)
	}

	return nil
}

func (a *Adapter) addToTar(tw *tar.Writer, filePath, sourceBase string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(sourceBase, filePath)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(stat, "")
	if err != nil {
		return err
	}
	header.Name = relPath

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if !stat.IsDir() {
		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
	}

	return nil
}

// ExecuteCommand runs a command on the remote server
func (a *Adapter) ExecuteCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// CreateBackup creates a backup of the target directory on the remote server
func (a *Adapter) CreateBackup(client *ssh.Client, targetPath string) error {
	timestamp := "$(date +%Y%m%d_%H%M%S)"
	backupPath := fmt.Sprintf("%s_backup_%s", targetPath, timestamp)
	command := fmt.Sprintf("if [ -d %s ]; then cp -r %s %s; fi", targetPath, targetPath, backupPath)

	_, err := a.ExecuteCommand(client, command)
	return err
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
