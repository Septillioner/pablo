package filter

import (
	"os"
	"path/filepath"
	"strings"
)

func Match(path string, patterns []string) (bool, error) {
	for _, pattern := range patterns {
		// Standardize separators for Windows/Unix compatibility
		pattern = filepath.ToSlash(pattern)
		matchPath := filepath.ToSlash(path)

		matched, err := filepath.Match(pattern, matchPath)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}

		// Handle directory-style patterns (e.g., "logs/")
		if strings.HasSuffix(pattern, "/") && strings.HasPrefix(matchPath, strings.TrimSuffix(pattern, "/")) {
			return true, nil
		}
	}
	return false, nil
}

func GetFiles(basePath string, includes, excludes []string) ([]string, error) {
	var files []string
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}

		// Check excludes
		if len(excludes) > 0 {
			excluded, err := Match(relPath, excludes)
			if err != nil {
				return err
			}
			if excluded {
				return nil
			}
		}

		// Check includes (if empty, include everything not excluded)
		if len(includes) > 0 {
			included, err := Match(relPath, includes)
			if err != nil {
				return err
			}
			if !included {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})

	return files, err
}
