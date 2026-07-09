package hooks

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExecuteEmptyCommand(t *testing.T) {
	if err := Execute("", t.TempDir(), map[string]string{"FOO": "bar"}); err != nil {
		t.Fatalf("Execute(\"\") error = %v, want nil", err)
	}
}

func TestExecuteEnvAndWorkingDir(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, "env.out")

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = `$env:PABLO_TEST_VAR | Out-File -FilePath env.out -Encoding ascii`
	} else {
		cmd = `echo "$PABLO_TEST_VAR" > env.out`
	}

	env := map[string]string{"PABLO_TEST_VAR": "injected-value"}
	if err := Execute(cmd, dir, env); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	data, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := strings.TrimSpace(string(data))
	if !strings.Contains(got, "injected-value") {
		t.Fatalf("env output = %q, want injected-value", got)
	}
}

func TestExecuteMissingWorkingDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "exit 0"
	} else {
		cmd = "true"
	}

	err := Execute(cmd, missing, nil)
	if err == nil {
		t.Fatal("expected error for missing working directory")
	}
}

func TestExecuteFailingCommand(t *testing.T) {
	dir := t.TempDir()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "exit 1"
	} else {
		cmd = "false"
	}

	err := Execute(cmd, dir, nil)
	if err == nil {
		t.Fatal("expected error for failing command")
	}
}
