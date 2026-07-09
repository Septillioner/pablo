package system

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func setTempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", home)
	} else {
		t.Setenv("HOME", home)
	}
	return home
}

func TestAddPathScopeRouting(t *testing.T) {
	switch runtime.GOOS {
	case "windows", "linux", "darwin":
		// supported platforms
	default:
		t.Skipf("unsupported platform: %s", runtime.GOOS)
	}

	t.Run("user scope on unix", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("unix user scope test")
		}
		home := setTempHome(t)
		cfg := filepath.Join(home, ".bashrc")
		if err := os.WriteFile(cfg, []byte("# shell config\n"), 0644); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		newPath := filepath.Join(home, "bin")
		if err := AddPath(newPath, "user", "test-project"); err != nil {
			t.Fatalf("AddPath() error = %v", err)
		}

		content, err := os.ReadFile(cfg)
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		tag := "# Added by pablo for test-project"
		if !strings.Contains(string(content), tag) {
			t.Fatalf("expected comment tag in %s", cfg)
		}
		if !strings.Contains(string(content), newPath) {
			t.Fatalf("expected path %q in shell config", newPath)
		}
	})

	t.Run("system scope requires elevated paths", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("windows system scope uses PowerShell; covered by removePathWindows test")
		}
		err := AddPath("/opt/pablo-test/bin", "system", "test-project")
		if err == nil {
			t.Skip("system scope succeeded; environment has write access to /etc")
		}
		if !strings.Contains(err.Error(), "permission denied") && !strings.Contains(err.Error(), "path registration failed") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAddPathUnixUserIdempotency(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix user scope only")
	}

	home := setTempHome(t)
	cfg := filepath.Join(home, ".bashrc")
	if err := os.WriteFile(cfg, []byte("# shell\n"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	newPath := filepath.Join(home, "tools", "bin")
	project := "idempotent-project"

	if err := AddPath(newPath, "user", project); err != nil {
		t.Fatalf("first AddPath() error = %v", err)
	}
	first, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if err := AddPath(newPath, "user", project); err != nil {
		t.Fatalf("second AddPath() error = %v", err)
	}
	second, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	tag := "# Added by pablo for idempotent-project"
	if strings.Count(string(second), tag) != 1 {
		t.Fatalf("expected exactly one project tag, got content:\n%s", string(second))
	}
	if len(second) != len(first) {
		t.Fatalf("idempotent AddPath changed file size: before=%d after=%d", len(first), len(second))
	}
}

func TestRemovePathUnixUser(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix user scope only")
	}

	home := setTempHome(t)
	cfg := filepath.Join(home, ".bashrc")
	project := "remove-me"
	newPath := filepath.Join(home, "app", "bin")
	tag := "# Added by pablo for remove-me"
	initial := "# shell config\n" + tag + "\nexport PATH=\"$PATH:" + newPath + "\"\n"
	if err := os.WriteFile(cfg, []byte(initial), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := RemovePath(project, "", "user"); err != nil {
		t.Fatalf("RemovePath() error = %v", err)
	}

	content, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	body := string(content)
	if strings.Contains(body, tag) {
		t.Fatalf("comment tag still present:\n%s", body)
	}
	if strings.Contains(body, newPath) {
		t.Fatalf("export line still present:\n%s", body)
	}
	if !strings.Contains(body, "# shell config") {
		t.Fatal("unrelated shell config line was removed")
	}
}

func TestRemovePathUnixSystem(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix system scope only")
	}

	dropIn := filepath.Join(t.TempDir(), "pablo.sh")
	initial := "export PATH=\"$PATH:/opt/pablo-app/bin\"\nexport PATH=\"$PATH:/opt/other/bin\"\n"
	if err := os.WriteFile(dropIn, []byte(initial), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	prev := systemDropInPathOverride
	systemDropInPathOverride = dropIn
	t.Cleanup(func() { systemDropInPathOverride = prev })

	if err := removePathUnixSystem("/opt/pablo-app/bin"); err != nil {
		t.Fatalf("removePathUnixSystem() error = %v", err)
	}

	content, err := os.ReadFile(dropIn)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	body := string(content)
	if strings.Contains(body, "/opt/pablo-app/bin") {
		t.Fatalf("target path line still present:\n%s", body)
	}
	if !strings.Contains(body, "/opt/other/bin") {
		t.Fatalf("unrelated path line was removed:\n%s", body)
	}
}

func TestRemovePathWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows-only behavior lock-in")
	}

	target := filepath.Join(t.TempDir(), "pablo-bin")
	err := removePathWindows(target, "user")
	if err != nil {
		t.Fatalf("removePathWindows() error = %v", err)
	}
}

func TestRemovePathUnsupportedPlatform(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		t.Skip("platform is supported")
	}

	err := RemovePath("proj", "/tmp/bin", "user")
	if err == nil {
		t.Fatal("expected unsupported platform error")
	}
	if !strings.Contains(err.Error(), "unsupported platform") {
		t.Fatalf("unexpected error: %v", err)
	}
}
