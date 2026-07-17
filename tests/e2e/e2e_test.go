//go:build integration

package e2e

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("PABLO_E2E_SKIP_DOCKER") == "1" {
		os.Exit(m.Run())
	}

	if err := setupE2E(); err != nil {
		fmt.Fprintf(os.Stderr, "e2e setup failed: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	teardownE2E()
	os.Exit(code)
}

func TestSSH_StaticDeploy(t *testing.T) {
	dir := scenarioDir(t, "ssh-static")
	manifest := "pablo.yaml"

	runPablo(t, dir, "check", "-f", manifest, "-p", "default", "-e", "e2e")
	runPablo(t, dir, "run", "-f", manifest, "-p", "default", "-e", "e2e", "--force")

	sshExec(t, "test -f /tmp/pablo-e2e-static/index.html")
}

func TestSSH_RenameReplace(t *testing.T) {
	dir := scenarioDir(t, "ssh-rename-replace")
	manifest := "pablo.yaml"
	targetDir := "/tmp/pablo-e2e-rename-replace"

	sshExec(t, fmt.Sprintf("rm -rf %s && mkdir -p %s", targetDir, targetDir))
	sshExec(t, fmt.Sprintf("printf 'old-content' > %s/index.html", targetDir))

	runPablo(t, dir, "check", "-f", manifest, "-p", "default", "-e", "e2e")
	runPablo(t, dir, "run", "-f", manifest, "-p", "default", "-e", "e2e", "--force")

	content := strings.TrimSpace(sshExec(t, fmt.Sprintf("cat %s/index.html", targetDir)))
	if !strings.Contains(content, "Rename Replace") {
		t.Fatalf("expected deployed HTML content, got: %q", content)
	}
	if strings.Contains(content, "old-content") {
		t.Fatal("expected old seed content to be replaced")
	}

	listing := strings.TrimSpace(sshExec(t, fmt.Sprintf("ls -1 %s", targetDir)))
	if listing != "index.html" {
		t.Fatalf("expected only index.html in target, got:\n%s", listing)
	}
}

func TestSSH_DockerRemoteDeploy(t *testing.T) {
	dir := scenarioDir(t, "ssh-docker-remote")
	manifest := "pablo.yaml"

	sshExec(t, "docker rm -f pablo-e2e-sample 2>/dev/null || true")

	runPablo(t, dir, "check", "-f", manifest, "-p", "default", "-e", "e2e")
	runPablo(t, dir, "run", "-f", manifest, "-p", "default", "-e", "e2e", "--force")

	assertSampleContainerRunning(t)

	// Redeploy while the stack is already up (stop_before_sync default).
	runPablo(t, dir, "run", "-f", manifest, "-p", "default", "-e", "e2e", "--force")
	assertSampleContainerRunning(t)
}

func assertSampleContainerRunning(t *testing.T) {
	t.Helper()
	out := sshExec(t, "docker ps --format '{{.Names}}'")
	if !strings.Contains(out, "pablo-e2e-sample") {
		t.Fatalf("expected container pablo-e2e-sample in docker ps, got: %q", out)
	}
}
