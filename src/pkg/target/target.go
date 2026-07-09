package target

import (
	"fmt"
	"strings"
)

func Parse(s string) (profile, env string, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", "", fmt.Errorf("target is required")
	}

	sep := -1
	sepChar := '/'
	if idx := strings.Index(s, "/"); idx >= 0 {
		sep = idx
		sepChar = '/'
	} else if idx := strings.Index(s, "."); idx >= 0 {
		sep = idx
		sepChar = '.'
	}

	if sep < 0 {
		return "", "", fmt.Errorf("target must be profile%cevironment (e.g. default/windows-local)", sepChar)
	}

	profile = strings.TrimSpace(s[:sep])
	env = strings.TrimSpace(s[sep+1:])
	if profile == "" || env == "" {
		return "", "", fmt.Errorf("target must include both profile and environment")
	}
	if strings.ContainsAny(env, "/.") {
		return "", "", fmt.Errorf("environment name cannot contain '/' or '.'")
	}

	return profile, env, nil
}

func Format(profile, env string) string {
	return profile + "/" + env
}
