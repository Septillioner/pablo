package deployer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"pablo/pkg/pathutil"

	"golang.org/x/crypto/ssh"
)

const renameReplaceTimeLayout = "20060102_150405"

type renameReplaceState struct {
	suffix  string
	renames []renamePair
	written []string
}

type renamePair struct {
	dest   string
	backup string
}

func renameReplaceSuffix() string {
	now := time.Now()
	return fmt.Sprintf("%s_%03d", now.Format(renameReplaceTimeLayout), now.Nanosecond()/1e6)
}

func (s *Service) deployRenameReplaceLocal(files []string, sourceBase, targetPath string) error {
	state := &renameReplaceState{suffix: renameReplaceSuffix()}

	for _, file := range files {
		rel, err := filepath.Rel(sourceBase, file)
		if err != nil {
			s.rollbackRenameReplaceLocal(state)
			return err
		}

		dest := filepath.Join(targetPath, rel)
		if err := s.renameReplaceStageLocal(dest, state); err != nil {
			s.rollbackRenameReplaceLocal(state)
			return err
		}

		if err := s.copyFile(file, dest); err != nil {
			s.rollbackRenameReplaceLocal(state)
			return err
		}
		state.written = append(state.written, dest)
	}

	s.cleanupRenameReplaceBackupsLocal(state)
	return nil
}

func (s *Service) renameReplaceStageLocal(dest string, state *renameReplaceState) error {
	if _, err := os.Stat(dest); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	backupPath := dest + "." + state.suffix
	if err := os.Rename(dest, backupPath); err != nil {
		return fmt.Errorf("failed to rename %s to %s: %w", dest, backupPath, err)
	}
	state.renames = append(state.renames, renamePair{dest: dest, backup: backupPath})
	return nil
}

func (s *Service) rollbackRenameReplaceLocal(state *renameReplaceState) {
	for _, dest := range state.written {
		_ = os.Remove(dest)
	}
	for i := len(state.renames) - 1; i >= 0; i-- {
		pair := state.renames[i]
		_ = os.Rename(pair.backup, pair.dest)
	}
}

func (s *Service) cleanupRenameReplaceBackupsLocal(state *renameReplaceState) {
	for _, pair := range state.renames {
		_ = os.Remove(pair.backup)
	}
}

func (s *Service) deployRenameReplaceRemote(
	files []string,
	sourceBase string,
	sshClient *ssh.Client,
	targetPath string,
	remoteTransfer string,
) error {
	state := &renameReplaceState{suffix: renameReplaceSuffix()}

	for _, file := range files {
		rel, err := filepath.Rel(sourceBase, file)
		if err != nil {
			s.rollbackRenameReplaceRemote(sshClient, state)
			return err
		}

		dest := pathutil.JoinRemote(targetPath, rel)
		if err := s.renameReplaceStageRemote(sshClient, dest, state); err != nil {
			s.rollbackRenameReplaceRemote(sshClient, state)
			return err
		}
		state.written = append(state.written, dest)
	}

	if err := s.transferRenameReplaceRemote(files, sourceBase, sshClient, targetPath, remoteTransfer); err != nil {
		s.rollbackRenameReplaceRemote(sshClient, state)
		return err
	}

	s.cleanupRenameReplaceBackupsRemote(sshClient, state)
	return nil
}

func (s *Service) renameReplaceStageRemote(sshClient *ssh.Client, dest string, state *renameReplaceState) error {
	backupPath := dest + "." + state.suffix
	command := fmt.Sprintf("if [ -e %s ]; then mv %s %s && echo __renamed__; fi", dest, dest, backupPath)
	out, err := s.ssh.ExecuteCommand(sshClient, command)
	if err != nil {
		return fmt.Errorf("failed to rename remote %s: %w", dest, err)
	}
	if strings.Contains(out, "__renamed__") {
		state.renames = append(state.renames, renamePair{dest: dest, backup: backupPath})
	}
	return nil
}

func (s *Service) transferRenameReplaceRemote(
	files []string,
	sourceBase string,
	sshClient *ssh.Client,
	targetPath string,
	remoteTransfer string,
) error {
	if remoteTransfer == "tar" {
		if err := s.ssh.TransferPipeline(sshClient, files, sourceBase, targetPath); err != nil {
			return fmt.Errorf("batch transfer failed: %w", err)
		}
		return nil
	}

	for _, file := range files {
		rel, err := filepath.Rel(sourceBase, file)
		if err != nil {
			return err
		}

		remoteDest := pathutil.JoinRemote(targetPath, rel)
		if err := s.ssh.TransferFile(sshClient, file, remoteDest); err != nil {
			return fmt.Errorf("failed to transfer %s: %w", file, err)
		}

		if _, err := s.ssh.ExecuteCommand(sshClient, fmt.Sprintf("chmod +x %s", remoteDest)); err != nil {
			return fmt.Errorf("failed to set permissions: %w", err)
		}
	}
	return nil
}

func (s *Service) rollbackRenameReplaceRemote(sshClient *ssh.Client, state *renameReplaceState) {
	for _, dest := range state.written {
		_, _ = s.ssh.ExecuteCommand(sshClient, fmt.Sprintf("rm -f %s", dest))
	}
	for i := len(state.renames) - 1; i >= 0; i-- {
		pair := state.renames[i]
		_, _ = s.ssh.ExecuteCommand(sshClient, fmt.Sprintf("if [ -e %s ]; then mv %s %s; fi", pair.backup, pair.backup, pair.dest))
	}
}

func (s *Service) cleanupRenameReplaceBackupsRemote(sshClient *ssh.Client, state *renameReplaceState) {
	for _, pair := range state.renames {
		_, _ = s.ssh.ExecuteCommand(sshClient, fmt.Sprintf("rm -f %s", pair.backup))
	}
}
