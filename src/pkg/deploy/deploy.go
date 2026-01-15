package deploy

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func Deploy(files []string, sourceBase, targetPath string, strategy string) error {
	switch strategy {
	case "backup":
		if err := Backup(targetPath); err != nil {
			return err
		}
	case "blue-green":
		return fmt.Errorf("blue-green strategy not yet fully implemented for general shares")
	}

	for _, file := range files {
		rel, err := filepath.Rel(sourceBase, file)
		if err != nil {
			return err
		}

		dest := filepath.Join(targetPath, rel)
		if err := copyFile(file, dest); err != nil {
			return err
		}
	}

	return nil
}

func Backup(targetPath string) error {
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return nil
	}

	backupName := fmt.Sprintf("%s_backup_%s", targetPath, time.Now().Format("20060102_150405"))
	fmt.Printf("Backing up %s to %s\n", targetPath, backupName)
	return os.Rename(targetPath, backupName)
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
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

	_, err = io.Copy(destFile, sourceFile)
	return err
}
