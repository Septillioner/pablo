package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsConfigExt(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"app.yaml", true},
		{"app.YML", true},
		{"settings.json", true},
		{"web.config", true},
		{"data.xml", true},
		{"notes.txt", true},
		{"app.ini", true},
		{"binary.exe", false},
		{"image.png", false},
		{"archive.tar.gz", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isConfigExt(tt.path); got != tt.want {
				t.Fatalf("isConfigExt(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestReplaceVariables(t *testing.T) {
	t.Run("happy path replaces placeholder", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "app.yaml")
		original := []byte("url: {{API_URL}}\n")
		if err := os.WriteFile(path, original, 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}

		err := replaceVariables(path, map[string]string{"API_URL": "https://example.com"})
		if err != nil {
			t.Fatalf("replaceVariables() error = %v", err)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		want := "url: https://example.com\n"
		if string(got) != want {
			t.Fatalf("content = %q, want %q", got, want)
		}
	})

	t.Run("no write when no placeholders", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "app.yaml")
		original := []byte("url: https://static.example.com\n")
		if err := os.WriteFile(path, original, 0o600); err != nil {
			t.Fatalf("write: %v", err)
		}
		infoBefore, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat: %v", err)
		}

		err = replaceVariables(path, map[string]string{"API_URL": "https://example.com"})
		if err != nil {
			t.Fatalf("replaceVariables() error = %v", err)
		}
		infoAfter, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat after: %v", err)
		}
		if infoBefore.ModTime() != infoAfter.ModTime() {
			t.Fatal("file should not be rewritten when no placeholders match")
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if string(got) != string(original) {
			t.Fatalf("content changed unexpectedly: %q", got)
		}
	})

	t.Run("multiple placeholders and repeated key", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "app.yaml")
		content := "host: {{HOST}}\nbackup: {{HOST}}\nport: {{PORT}}\n"
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}

		err := replaceVariables(path, map[string]string{"HOST": "localhost", "PORT": "8080"})
		if err != nil {
			t.Fatalf("replaceVariables() error = %v", err)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		want := "host: localhost\nbackup: localhost\nport: 8080\n"
		if string(got) != want {
			t.Fatalf("content = %q, want %q", got, want)
		}
	})
}

func TestProcessFiles(t *testing.T) {
	t.Run("processes only config extensions", func(t *testing.T) {
		dir := t.TempDir()
		yamlPath := filepath.Join(dir, "app.yaml")
		exePath := filepath.Join(dir, "app.exe")
		pngPath := filepath.Join(dir, "logo.png")
		for path, body := range map[string]string{
			yamlPath: "key: {{KEY}}\n",
			exePath:  "{{KEY}}",
			pngPath:  "{{KEY}}",
		} {
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				t.Fatalf("write %s: %v", path, err)
			}
		}

		err := ProcessFiles(dir, map[string]string{"KEY": "value"})
		if err != nil {
			t.Fatalf("ProcessFiles() error = %v", err)
		}

		yamlGot, _ := os.ReadFile(yamlPath)
		if string(yamlGot) != "key: value\n" {
			t.Fatalf("yaml content = %q", yamlGot)
		}
		exeGot, _ := os.ReadFile(exePath)
		if string(exeGot) != "{{KEY}}" {
			t.Fatalf("exe should be untouched, got %q", exeGot)
		}
		pngGot, _ := os.ReadFile(pngPath)
		if string(pngGot) != "{{KEY}}" {
			t.Fatalf("png should be untouched, got %q", pngGot)
		}
	})

	t.Run("empty variables map is no-op", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "app.yaml")
		if err := os.WriteFile(path, []byte("key: {{KEY}}\n"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}

		err := ProcessFiles(dir, map[string]string{})
		if err != nil {
			t.Fatalf("ProcessFiles() error = %v", err)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if string(got) != "key: {{KEY}}\n" {
			t.Fatalf("content should be unchanged, got %q", got)
		}
	})
}
