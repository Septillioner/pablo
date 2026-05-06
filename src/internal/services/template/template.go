package template

import (
	"os"
	"path/filepath"
	"strings"
)

func ProcessFiles(targetPath string, variables map[string]string) error {
	if len(variables) == 0 {
		return nil
	}

	return filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Only process text-based config files
		ext := strings.ToLower(os.ExpandEnv(path))
		if !isConfigExt(ext) {
			return nil
		}

		return replaceVariables(path, variables)
	})
}

func isConfigExt(path string) bool {
	exts := []string{".config", ".json", ".yaml", ".yml", ".xml", ".txt", ".ini"}
	for _, e := range exts {
		if strings.HasSuffix(strings.ToLower(path), e) {
			return true
		}
	}
	return false
}

func replaceVariables(path string, variables map[string]string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(data)
	modified := false

	for k, v := range variables {
		placeholder := "{{" + k + "}}"
		if strings.Contains(content, placeholder) {
			content = strings.ReplaceAll(content, placeholder, v)
			modified = true
		}
	}

	if modified {
		return os.WriteFile(path, []byte(content), 0644)
	}

	return nil
}
