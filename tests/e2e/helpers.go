//go:build integration

package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

const sshHost = "127.0.0.1:2222"

var e2eRoot string

func initE2ERoot(t *testing.T) string {
	t.Helper()
	if e2eRoot != "" {
		return e2eRoot
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	e2eRoot = wd
	return e2eRoot
}

func generateTestKeys() error {
	dir := filepath.Join(e2eRoot, "keys")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("mkdir keys: %w", err)
	}
	priv := filepath.Join(dir, "id_ed25519")
	pub := priv + ".pub"
	if _, err := os.Stat(priv); err == nil {
		return nil
	}
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", priv, "-N", "")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ssh-keygen: %w\n%s", err, out)
	}
	if _, err := os.Stat(pub); err != nil {
		return fmt.Errorf("missing public key: %w", err)
	}
	return nil
}

func setupE2E() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	e2eRoot = wd
	if err := generateTestKeys(); err != nil {
		return err
	}
	cmd := exec.Command("docker", "compose", "up", "-d", "--wait", "--build")
	cmd.Dir = filepath.Join(e2eRoot, "docker")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker compose up: %w\n%s", err, out)
	}
	return waitForSSH(30 * time.Second)
}

func teardownE2E() {
	cmd := exec.Command("docker", "compose", "down", "-v")
	cmd.Dir = filepath.Join(e2eRoot, "docker")
	_, _ = cmd.CombinedOutput()
}

func waitForSSH(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if err := pingSSH(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("ssh not ready after %s: %w", timeout, lastErr)
}

func pingSSH() error {
	client, err := sshDial()
	if err != nil {
		return err
	}
	client.Close()
	return nil
}

func pabloBinaryPath(t *testing.T) string {
	t.Helper()
	name := "pablo-e2e"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(initE2ERoot(t), "..", "..", "build", name)
}

func ensurePabloBinary(t *testing.T) {
	t.Helper()
	bin := pabloBinaryPath(t)
	if st, err := os.Stat(bin); err == nil && st.Size() > 0 {
		return
	}
	if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
		t.Fatalf("mkdir build: %v", err)
	}
	srcDir := filepath.Join(initE2ERoot(t), "..", "..", "src")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = srcDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build pablo: %v\n%s", err, out)
	}
}

func runPablo(t *testing.T, scenarioDir string, args ...string) {
	t.Helper()
	ensurePabloBinary(t)
	cmd := exec.Command(pabloBinaryPath(t), args...)
	cmd.Dir = scenarioDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("pablo %s failed: %v\nstdout:\n%s\nstderr:\n%s",
			strings.Join(args, " "), err, stdout.String(), stderr.String())
	}
}

func sshDial() (*ssh.Client, error) {
	keyPath := filepath.Join(e2eRoot, "keys", "id_ed25519")
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read key: %w", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}
	config := &ssh.ClientConfig{
		User:            "deploy",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return ssh.Dial("tcp", sshHost, config)
}

func sshExec(t *testing.T, command string) string {
	t.Helper()
	client, err := sshDial()
	if err != nil {
		t.Fatalf("ssh dial: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("ssh session: %v", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	if err := session.Run(command); err != nil {
		t.Fatalf("ssh run %q: %v\nstdout: %s\nstderr: %s",
			command, err, stdout.String(), stderr.String())
	}
	return stdout.String()
}

func scenarioDir(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(initE2ERoot(t), "scenarios", name)
}
