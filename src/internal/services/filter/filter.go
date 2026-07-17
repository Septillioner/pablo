package filter

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Match(relPath string, patterns []string) (bool, error) {
	for _, pattern := range patterns {
		matched, err := matchPattern(pattern, relPath)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func matchPattern(pattern, relPath string) (bool, error) {
	// Normalize separators for cross-platform YAML globs. filepath.ToSlash alone
	// leaves '\' intact on Unix, so Windows-style patterns would never match.
	pattern = toMatchSlash(pattern)
	matchPath := toMatchSlash(relPath)

	if strings.HasSuffix(pattern, "/") {
		dir := strings.TrimSuffix(pattern, "/")
		return matchPath == dir || strings.HasPrefix(matchPath, dir+"/"), nil
	}

	if strings.Contains(pattern, "**") {
		return matchGlobstar(pattern, matchPath)
	}

	if !strings.Contains(pattern, "/") {
		return path.Match(pattern, path.Base(matchPath))
	}

	anchored := strings.TrimPrefix(pattern, "./")
	anchored = strings.TrimPrefix(anchored, "/")
	return path.Match(anchored, matchPath)
}

func matchGlobstar(pattern, matchPath string) (bool, error) {
	if strings.HasPrefix(pattern, "**/") {
		suffix := strings.TrimPrefix(pattern, "**/")
		if suffix == "" || suffix == "*" {
			return true, nil
		}
		if matched, err := path.Match(suffix, matchPath); err != nil {
			return false, err
		} else if matched {
			return true, nil
		}
		if matched, err := path.Match(suffix, path.Base(matchPath)); err != nil {
			return false, err
		} else if matched {
			return true, nil
		}
		for i := 0; i < len(matchPath); i++ {
			if matchPath[i] != '/' {
				continue
			}
			rest := matchPath[i+1:]
			if matched, err := path.Match(suffix, rest); err != nil {
				return false, err
			} else if matched {
				return true, nil
			}
		}
		return false, nil
	}

	if pattern == "**" {
		return true, nil
	}

	starIdx := strings.Index(pattern, "**")
	if starIdx < 0 {
		return path.Match(pattern, matchPath)
	}

	before := strings.TrimSuffix(pattern[:starIdx], "/")
	after := pattern[starIdx+2:]
	if strings.HasPrefix(after, "/") {
		after = after[1:]
	}

	if before == "" {
		if after == "" {
			return true, nil
		}
		return matchGlobstar("**/"+after, matchPath)
	}

	if matchPath != before && !strings.HasPrefix(matchPath, before+"/") {
		return false, nil
	}

	rest := strings.TrimPrefix(matchPath, before)
	rest = strings.TrimPrefix(rest, "/")
	if after == "" {
		return true, nil
	}
	return matchGlobstar("**/"+after, rest)
}

func toMatchSlash(p string) string {
	return strings.ReplaceAll(filepath.ToSlash(p), `\`, "/")
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

		if len(excludes) > 0 {
			excluded, err := Match(relPath, excludes)
			if err != nil {
				return err
			}
			if excluded {
				return nil
			}
		}

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
