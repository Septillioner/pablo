package deployer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	sshAdapter "pablo/internal/adapters/ssh"
	"pablo/pkg/domain"

	"golang.org/x/crypto/ssh"
)

func TestIsProtectedPath(t *testing.T) {
	svc := New()

	t.Run("unix protected roots", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("unix path semantics differ on Windows")
		}
		tests := []struct {
			path string
			want bool
		}{
			{"/", true},
			{"/etc", true},
			{"/usr", true},
			{"/bin", true},
			{"/var", true},
		}
		for _, tt := range tests {
			t.Run(tt.path, func(t *testing.T) {
				if got := svc.isProtectedPath(tt.path); got != tt.want {
					t.Fatalf("isProtectedPath(%q) = %v, want %v", tt.path, got, tt.want)
				}
			})
		}
	})

	t.Run("windows protected roots", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("windows path checks only on Windows")
		}
		tests := []struct {
			path string
			want bool
		}{
			{`C:\`, true},
			{`C:\Windows`, true},
			{`C:\Program Files`, true},
			{`C:\apps\myapp`, false},
		}
		for _, tt := range tests {
			t.Run(tt.path, func(t *testing.T) {
				if got := svc.isProtectedPath(tt.path); got != tt.want {
					t.Fatalf("isProtectedPath(%q) = %v, want %v", tt.path, got, tt.want)
				}
			})
		}
	})

	t.Run("safe subdirectory", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("unix path semantics differ on Windows")
		}
		if got := svc.isProtectedPath("/opt/myapp"); got {
			t.Fatal("expected /opt/myapp to be safe")
		}
	})
}

func TestDeployRenameReplace(t *testing.T) {
	svc := New()

	t.Run("missing target copies without rename", func(t *testing.T) {
		source := t.TempDir()
		target := filepath.Join(t.TempDir(), "deploy")
		mustWrite(t, filepath.Join(source, "app.txt"), "v1")

		if err := svc.Deploy([]string{filepath.Join(source, "app.txt")}, source, target, "rename-replace", false); err != nil {
			t.Fatalf("Deploy() error = %v", err)
		}
		assertFileContent(t, filepath.Join(target, "app.txt"), "v1")
	})

	t.Run("existing file replaced and backup removed", func(t *testing.T) {
		source := t.TempDir()
		target := t.TempDir()
		mustWrite(t, filepath.Join(target, "app.txt"), "old")
		mustWrite(t, filepath.Join(source, "app.txt"), "new")

		if err := svc.Deploy([]string{filepath.Join(source, "app.txt")}, source, target, "rename-replace", false); err != nil {
			t.Fatalf("Deploy() error = %v", err)
		}

		assertFileContent(t, filepath.Join(target, "app.txt"), "new")

		entries, err := os.ReadDir(target)
		if err != nil {
			t.Fatalf("ReadDir() error = %v", err)
		}
		for _, e := range entries {
			if strings.Contains(e.Name(), ".") && strings.Count(e.Name(), "_") >= 2 {
				t.Fatalf("expected no timestamp backup file, found %q", e.Name())
			}
		}
	})

	t.Run("failure rolls back renames and written files", func(t *testing.T) {
		source := t.TempDir()
		target := t.TempDir()
		mustWrite(t, filepath.Join(target, "first.txt"), "old-first")
		mustWrite(t, filepath.Join(source, "first.txt"), "new-first")

		files := []string{
			filepath.Join(source, "first.txt"),
			filepath.Join(source, "missing.txt"),
		}

		err := svc.Deploy(files, source, target, "rename-replace", false)
		if err == nil {
			t.Fatal("expected Deploy() to fail on missing source file")
		}

		assertFileContent(t, filepath.Join(target, "first.txt"), "old-first")
	})
}

func TestDeployRenameReplaceRemote(t *testing.T) {
	source := t.TempDir()
	mustWrite(t, filepath.Join(source, "app.txt"), "remote")
	files := []string{filepath.Join(source, "app.txt")}
	dummyClient := &ssh.Client{}

	t.Run("stages rename before tar transfer", func(t *testing.T) {
		mock := &mockSSH{remoteFileExists: true}
		svc := &Service{ssh: mock}

		if err := svc.DeployRemote(files, source, dummyClient, "/tmp/pablo-rename", "rename-replace", false, ""); err != nil {
			t.Fatalf("DeployRemote() error = %v", err)
		}
		if !mock.usedTransferPipe {
			t.Fatal("expected TransferPipeline for rename-replace tar mode")
		}

		foundRename := false
		foundCleanup := false
		for _, cmd := range mock.executeCalls {
			if strings.Contains(cmd, "mv /tmp/pablo-rename/app.txt") {
				foundRename = true
			}
			if strings.Contains(cmd, "rm -f /tmp/pablo-rename/app.txt.") {
				foundCleanup = true
			}
		}
		if !foundRename {
			t.Fatalf("expected remote rename command, got: %v", mock.executeCalls)
		}
		if !foundCleanup {
			t.Fatalf("expected backup cleanup command, got: %v", mock.executeCalls)
		}
	})

	t.Run("transfer failure rolls back", func(t *testing.T) {
		mock := &mockSSH{remoteFileExists: true, transferPipeErr: fmt.Errorf("transfer failed")}
		svc := &Service{ssh: mock}

		err := svc.DeployRemote(files, source, dummyClient, "/tmp/pablo-rename", "rename-replace", false, "")
		if err == nil {
			t.Fatal("expected transfer error")
		}

		foundRollbackRemove := false
		foundRollbackRestore := false
		for _, cmd := range mock.executeCalls {
			if strings.Contains(cmd, "rm -f /tmp/pablo-rename/app.txt") && !strings.Contains(cmd, "app.txt.") {
				foundRollbackRemove = true
			}
			if strings.Contains(cmd, "mv /tmp/pablo-rename/app.txt.") && strings.Contains(cmd, "/tmp/pablo-rename/app.txt") {
				foundRollbackRestore = true
			}
		}
		if !foundRollbackRemove {
			t.Fatalf("expected rollback remove, got: %v", mock.executeCalls)
		}
		if !foundRollbackRestore {
			t.Fatalf("expected rollback restore, got: %v", mock.executeCalls)
		}
	})
}

func TestDeployOverwrite(t *testing.T) {
	svc := New()
	source := t.TempDir()
	target := t.TempDir()

	mustWrite(t, filepath.Join(source, "a.txt"), "alpha")
	mustWrite(t, filepath.Join(source, "sub", "b.txt"), "beta")

	files := []string{
		filepath.Join(source, "a.txt"),
		filepath.Join(source, "sub", "b.txt"),
	}

	if err := svc.Deploy(files, source, target, "overwrite", false); err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	assertFileContent(t, filepath.Join(target, "a.txt"), "alpha")
	assertFileContent(t, filepath.Join(target, "sub", "b.txt"), "beta")
}

func TestDeployBackup(t *testing.T) {
	svc := New()

	t.Run("missing target is no-op", func(t *testing.T) {
		source := t.TempDir()
		target := filepath.Join(t.TempDir(), "deploy")
		mustWrite(t, filepath.Join(source, "app.txt"), "v1")

		if err := svc.Deploy([]string{filepath.Join(source, "app.txt")}, source, target, "backup", false); err != nil {
			t.Fatalf("Deploy() error = %v", err)
		}
		assertFileContent(t, filepath.Join(target, "app.txt"), "v1")
	})

	t.Run("existing target is renamed", func(t *testing.T) {
		source := t.TempDir()
		parent := t.TempDir()
		target := filepath.Join(parent, "site")
		mustWrite(t, filepath.Join(target, "old.txt"), "old")
		mustWrite(t, filepath.Join(source, "new.txt"), "new")

		if err := svc.Deploy([]string{filepath.Join(source, "new.txt")}, source, target, "backup", false); err != nil {
			t.Fatalf("Deploy() error = %v", err)
		}

		assertFileContent(t, filepath.Join(target, "new.txt"), "new")

		entries, err := os.ReadDir(parent)
		if err != nil {
			t.Fatalf("ReadDir() error = %v", err)
		}
		backupFound := false
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "site_backup_") {
				backupFound = true
				assertFileContent(t, filepath.Join(parent, e.Name(), "old.txt"), "old")
			}
		}
		if !backupFound {
			t.Fatal("expected backup directory with site_backup_ prefix")
		}
	})
}

func TestDeployRecreate(t *testing.T) {
	svc := New()
	source := t.TempDir()
	target := t.TempDir()

	mustWrite(t, filepath.Join(target, "stale.txt"), "stale")
	mustWrite(t, filepath.Join(source, "fresh.txt"), "fresh")

	if err := svc.Deploy([]string{filepath.Join(source, "fresh.txt")}, source, target, "recreate", false); err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(target, "stale.txt")); !os.IsNotExist(err) {
		t.Fatal("expected stale file to be removed")
	}
	assertFileContent(t, filepath.Join(target, "fresh.txt"), "fresh")
}

func TestDeployProtectedPathSafety(t *testing.T) {
	svc := New()
	source := t.TempDir()
	mustWrite(t, filepath.Join(source, "app.txt"), "data")

	var protected string
	if runtime.GOOS == "windows" {
		protected = `C:\Windows`
	} else {
		protected = "/etc"
	}

	strategies := []string{"backup", "recreate"}
	for _, strategy := range strategies {
		t.Run(strategy, func(t *testing.T) {
			err := svc.Deploy([]string{filepath.Join(source, "app.txt")}, source, protected, strategy, false)
			if err == nil {
				t.Fatalf("expected safety error for %s on %s", strategy, protected)
			}
			if !strings.Contains(err.Error(), "safety break") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestCopyFilePreservesMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file mode semantics differ on Windows")
	}

	svc := New()
	dir := t.TempDir()
	src := filepath.Join(dir, "bin")
	dst := filepath.Join(dir, "out", "bin")

	const wantMode os.FileMode = 0750
	if err := os.WriteFile(src, []byte("exec"), wantMode); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := svc.copyFile(src, dst); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode().Perm() != wantMode.Perm() {
		t.Fatalf("mode = %o, want %o", info.Mode().Perm(), wantMode.Perm())
	}
}

type mockSSH struct {
	executeCalls      []string
	usedTransferPipe  bool
	transferFileCount int
	transferPipeErr   error
	remoteFileExists  bool
}

func (m *mockSSH) Connect(host string, cred *domain.CredentialConfig, opts sshAdapter.HostKeyOptions) (*ssh.Client, error) {
	return nil, nil
}

func (m *mockSSH) ExecuteCommand(_ *ssh.Client, command string) (string, error) {
	m.executeCalls = append(m.executeCalls, command)
	if m.remoteFileExists && strings.Contains(command, "__renamed__") {
		return "__renamed__", nil
	}
	return "", nil
}

func (m *mockSSH) ExecuteCommandWithStdin(_ *ssh.Client, command string, _ io.Reader) (string, error) {
	m.executeCalls = append(m.executeCalls, command)
	return "", nil
}

func (m *mockSSH) CreateBackup(_ *ssh.Client, targetPath string) error {
	return nil
}

func (m *mockSSH) TransferPipeline(_ *ssh.Client, files []string, sourceBase, remotePath string) error {
	m.usedTransferPipe = true
	if m.transferPipeErr != nil {
		return m.transferPipeErr
	}
	return nil
}

func (m *mockSSH) TransferFile(_ *ssh.Client, localPath, remotePath string) error {
	m.transferFileCount++
	return nil
}

func TestDeployRemote(t *testing.T) {
	source := t.TempDir()
	mustWrite(t, filepath.Join(source, "app.txt"), "remote")

	files := []string{filepath.Join(source, "app.txt")}
	dummyClient := &ssh.Client{}

	t.Run("default uses tar transfer", func(t *testing.T) {
		mock := &mockSSH{}
		svc := &Service{ssh: mock}

		if err := svc.DeployRemote(files, source, dummyClient, "/var/www/app", "overwrite", false, ""); err != nil {
			t.Fatalf("DeployRemote() error = %v", err)
		}
		if !mock.usedTransferPipe {
			t.Fatal("expected TransferPipeline for default remoteTransfer")
		}
		if mock.transferFileCount != 0 {
			t.Fatalf("expected no TransferFile calls, got %d", mock.transferFileCount)
		}
	})

	t.Run("scp transfer branch", func(t *testing.T) {
		mock := &mockSSH{}
		svc := &Service{ssh: mock}

		if err := svc.DeployRemote(files, source, dummyClient, "/var/www/app", "overwrite", false, "scp"); err != nil {
			t.Fatalf("DeployRemote() error = %v", err)
		}
		if mock.usedTransferPipe {
			t.Fatal("expected TransferFile path, not TransferPipeline")
		}
		if mock.transferFileCount != 1 {
			t.Fatalf("expected 1 TransferFile call, got %d", mock.transferFileCount)
		}
	})

	t.Run("protected remote path", func(t *testing.T) {
		mock := &mockSSH{}
		svc := &Service{ssh: mock}

		protected := "/usr"
		if runtime.GOOS == "windows" {
			protected = `C:\Windows`
		}

		err := svc.DeployRemote(files, source, dummyClient, protected, "recreate", false, "tar")
		if err == nil {
			t.Fatal("expected safety error for protected remote path")
		}
		if !strings.Contains(err.Error(), "safety break") {
			t.Fatalf("unexpected error: %v", err)
		}
		if mock.usedTransferPipe {
			t.Fatal("expected no transfer before safety check failure")
		}
	})
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	if string(data) != want {
		t.Fatalf("file %s = %q, want %q", path, string(data), want)
	}
}
