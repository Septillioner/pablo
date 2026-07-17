package deployer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

// VerifyRemoteChecksums hashes local artifacts and checks them on the remote
// host with `sha256sum -c` over stdin (no remote checksum file on disk).
func (s *Service) VerifyRemoteChecksums(client *ssh.Client, files []string, sourceBase, targetPath string) error {
	var lines strings.Builder
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf("checksum stat %s: %w", file, err)
		}
		if info.IsDir() {
			continue
		}

		rel, err := filepath.Rel(sourceBase, file)
		if err != nil {
			return fmt.Errorf("checksum rel %s: %w", file, err)
		}
		rel = filepath.ToSlash(rel)

		sum, err := hashFileSHA256(file)
		if err != nil {
			return fmt.Errorf("checksum hash %s: %w", file, err)
		}
		lines.WriteString(sum)
		lines.WriteString("  ")
		lines.WriteString(rel)
		lines.WriteString("\n")
	}

	manifest := lines.String()
	if manifest == "" {
		return nil
	}

	targetPath = filepath.ToSlash(targetPath)
	cmd := fmt.Sprintf("cd %s && sha256sum -c -", targetPath)
	if strings.ContainsAny(targetPath, " \t") {
		cmd = fmt.Sprintf("cd %q && sha256sum -c -", targetPath)
	}
	output, err := s.ssh.ExecuteCommandWithStdin(client, cmd, strings.NewReader(manifest))
	if err != nil {
		trimmed := strings.TrimSpace(output)
		if trimmed != "" {
			return fmt.Errorf("remote checksum verification failed: %w\n%s", err, trimmed)
		}
		return fmt.Errorf("remote checksum verification failed: %w", err)
	}
	return nil
}

func hashFileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
