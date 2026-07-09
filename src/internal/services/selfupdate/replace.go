package selfupdate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

func replaceExecutable(targetPath, sourcePath string) error {
	targetPath, err := filepath.Abs(filepath.Clean(targetPath))
	if err != nil {
		return err
	}
	sourcePath, err = filepath.Abs(filepath.Clean(sourcePath))
	if err != nil {
		return err
	}

	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("create install directory: %w", err)
	}

	if runtime.GOOS == "windows" {
		return replaceExecutableWindows(targetPath, sourcePath)
	}
	return replaceExecutableUnix(targetPath, sourcePath)
}

func replaceExecutableWindows(targetPath, sourcePath string) error {
	oldPath := targetPath + ".old"
	_ = os.Remove(oldPath)

	if _, err := os.Stat(targetPath); err == nil {
		if err := os.Rename(targetPath, oldPath); err != nil {
			return fmt.Errorf(
				"cannot replace %s: file may be locked (close terminals or VS using pablo, then retry): %w",
				targetPath,
				err,
			)
		}
	}

	if err := copyFile(sourcePath, targetPath); err != nil {
		_ = os.Rename(oldPath, targetPath)
		return err
	}

	_ = os.Remove(oldPath)
	return nil
}

func replaceExecutableUnix(targetPath, sourcePath string) error {
	targetDir := filepath.Dir(targetPath)
	tempTarget := filepath.Join(targetDir, "."+filepath.Base(targetPath)+".new")

	_ = os.Remove(tempTarget)
	if err := copyFile(sourcePath, tempTarget); err != nil {
		return err
	}
	if err := os.Chmod(tempTarget, 0o755); err != nil {
		os.Remove(tempTarget)
		return err
	}

	oldPath := targetPath + ".old"
	_ = os.Remove(oldPath)

	if _, err := os.Stat(targetPath); err == nil {
		if err := os.Rename(targetPath, oldPath); err != nil {
			os.Remove(tempTarget)
			return fmt.Errorf("cannot replace %s: %w", targetPath, err)
		}
	}

	if err := os.Rename(tempTarget, targetPath); err != nil {
		_ = os.Rename(oldPath, targetPath)
		os.Remove(tempTarget)
		return fmt.Errorf("install updated binary: %w", err)
	}

	_ = os.Remove(oldPath)
	return nil
}

func copyFile(sourcePath, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open source binary: %w", err)
	}
	defer source.Close()

	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("write target binary: %w", err)
	}

	if _, err := io.Copy(target, source); err != nil {
		target.Close()
		os.Remove(targetPath)
		return fmt.Errorf("copy binary: %w", err)
	}
	if err := target.Close(); err != nil {
		os.Remove(targetPath)
		return err
	}
	return nil
}
