package pipeline

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func writeSequenceManifest(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "pablo.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}

func failBuildCommand() string {
	if runtime.GOOS == "windows" {
		return "exit /b 1"
	}
	return "exit 1"
}

func TestRunSequence(t *testing.T) {
	s := newTestPipeline()

	t.Run("sequence not found", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src")
		if err := os.MkdirAll(src, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(src, "index.html"), []byte("<html></html>"), 0o644); err != nil {
			t.Fatalf("write artifact: %v", err)
		}

		path := writeSequenceManifest(t, dir, `
name: seq-app
version: 1.0.0
profiles:
  local:
    type: static
    environments:
      a:
        deploy:
          source:
            dir: ./src
            include: ["*.html"]
          target_path: ./dist-a
          strategy: overwrite
`)
		err := s.RunSequence(path, "missing", false, false)
		if err == nil {
			t.Fatal("expected error for missing sequence")
		}
		if !strings.Contains(err.Error(), "sequence not found") {
			t.Fatalf("error = %v, want sequence not found", err)
		}
	})

	t.Run("runs steps in list order", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src")
		if err := os.MkdirAll(src, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(src, "index.html"), []byte("<html></html>"), 0o644); err != nil {
			t.Fatalf("write artifact: %v", err)
		}

		path := writeSequenceManifest(t, dir, `
name: seq-app
version: 1.0.0
sequences:
  local:
    - app/first
    - app/second
profiles:
  app:
    type: static
    environments:
      first:
        deploy:
          source:
            dir: ./src
            include: ["*.html"]
          target_path: ./dist-first
          strategy: overwrite
      second:
        deploy:
          source:
            dir: ./src
            include: ["*.html"]
          target_path: ./dist-second
          strategy: overwrite
`)
		if err := s.RunSequence(path, "local", false, false); err != nil {
			t.Fatalf("RunSequence() error = %v", err)
		}

		first := filepath.Join(dir, "dist-first", "index.html")
		second := filepath.Join(dir, "dist-second", "index.html")
		if _, err := os.Stat(first); err != nil {
			t.Fatalf("first step artifact missing: %v", err)
		}
		if _, err := os.Stat(second); err != nil {
			t.Fatalf("second step artifact missing: %v", err)
		}
	})

	t.Run("aborts on first failure", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src")
		if err := os.MkdirAll(src, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(src, "index.html"), []byte("<html></html>"), 0o644); err != nil {
			t.Fatalf("write artifact: %v", err)
		}

		path := writeSequenceManifest(t, dir, `
name: seq-app
version: 1.0.0
sequences:
  local:
    - app/fail
    - app/ok
profiles:
  app:
    type: static
    environments:
      fail:
        build:
          command: "`+failBuildCommand()+`"
          path: .
        deploy:
          source:
            dir: ./src
            include: ["*.html"]
          target_path: ./dist-fail
          strategy: overwrite
      ok:
        deploy:
          source:
            dir: ./src
            include: ["*.html"]
          target_path: ./dist-ok
          strategy: overwrite
`)
		err := s.RunSequence(path, "local", false, false)
		if err == nil {
			t.Fatal("expected RunSequence to fail on first step")
		}

		okPath := filepath.Join(dir, "dist-ok", "index.html")
		if _, err := os.Stat(okPath); err == nil {
			t.Fatal("second step should not run after first failure")
		}
	})
}
