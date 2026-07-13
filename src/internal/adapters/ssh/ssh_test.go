package ssh

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pablo/pkg/domain"
)

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"plain", "/etc/ssh/id_rsa", "/etc/ssh/id_rsa"},
		{"tilde", "~/keys/id_ed25519", filepath.Join(home, "keys", "id_ed25519")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expandPath(tt.in); got != tt.want {
				t.Fatalf("expandPath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestConnectErrors(t *testing.T) {
	a := New()
	opts := HostKeyOptions{Verification: hostKeyVerificationOff}

	t.Run("unsupported credential type", func(t *testing.T) {
		_, err := a.Connect("host", &domain.CredentialConfig{Type: "token", Username: "u"}, opts)
		if err == nil || !strings.Contains(err.Error(), "unsupported credential type") {
			t.Fatalf("expected unsupported type error, got %v", err)
		}
	})

	t.Run("ssh missing key and password", func(t *testing.T) {
		_, err := a.Connect("host", &domain.CredentialConfig{Type: "ssh", Username: "deploy"}, opts)
		if err == nil || !strings.Contains(err.Error(), "must have either key or password") {
			t.Fatalf("expected key/password error, got %v", err)
		}
	})

	t.Run("ssh missing key file", func(t *testing.T) {
		_, err := a.Connect("host", &domain.CredentialConfig{
			Type:     "ssh",
			Username: "deploy",
			Key:      filepath.Join(t.TempDir(), "nonexistent_key"),
		}, opts)
		if err == nil || !strings.Contains(err.Error(), "failed to read SSH key") {
			t.Fatalf("expected read key error, got %v", err)
		}
	})
}

func TestAddToTar(t *testing.T) {
	a := New()
	base := t.TempDir()
	sub := filepath.Join(base, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(sub, "hello.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	if err := a.addToTar(tw, filePath, base); err != nil {
		t.Fatalf("addToTar: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}

	tr := tar.NewReader(&buf)
	hdr, err := tr.Next()
	if err != nil {
		t.Fatalf("read tar header: %v", err)
	}
	rel, err := filepath.Rel(base, filePath)
	if err != nil {
		t.Fatal(err)
	}
	if hdr.Name != rel {
		t.Fatalf("header name %q, want %q", hdr.Name, rel)
	}
	var content bytes.Buffer
	if _, err := content.ReadFrom(tr); err != nil {
		t.Fatal(err)
	}
	if content.String() != "hello" {
		t.Fatalf("content %q, want hello", content.String())
	}
}

func TestAddToTarMissingFile(t *testing.T) {
	a := New()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	err := a.addToTar(tw, filepath.Join(t.TempDir(), "missing.txt"), t.TempDir())
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
