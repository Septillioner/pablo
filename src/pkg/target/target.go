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

	sep := strings.Index(s, "/")
	if sep < 0 {
		return "", "", fmt.Errorf("target must be profile/environment (e.g. default/windows-local)")
	}

	profile = strings.TrimSpace(s[:sep])
	env = strings.TrimSpace(s[sep+1:])
	if profile == "" || env == "" {
		return "", "", fmt.Errorf("target must include both profile and environment")
	}
	if strings.Contains(env, "/") {
		return "", "", fmt.Errorf("environment name cannot contain '/'")
	}

	return profile, env, nil
}

func Format(profile, env string) string {
	return profile + "/" + env
}
