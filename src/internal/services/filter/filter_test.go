package filter

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMatch(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		tests := []struct {
			name     string
			path     string
			patterns []string
			want     bool
		}{
			{"glob star suffix", "app.js", []string{"*.js"}, true},
			{"glob star suffix no match", "app.go", []string{"*.js"}, false},
			{"exact name", "readme.txt", []string{"readme.txt"}, true},
			{"any of multiple patterns", "main.go", []string{"*.js", "*.go"}, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := Match(tt.path, tt.patterns)
				if err != nil {
					t.Fatalf("Match() error = %v", err)
				}
				if got != tt.want {
					t.Fatalf("Match(%q, %v) = %v, want %v", tt.path, tt.patterns, got, tt.want)
				}
			})
		}
	})

	t.Run("edge cases", func(t *testing.T) {
		tests := []struct {
			name     string
			path     string
			patterns []string
			want     bool
		}{
			{"empty pattern list", "file.txt", nil, false},
			{"directory trailing slash pattern", "logs/app.log", []string{"logs/"}, true},
			{"directory prefix without slash", "logs/app.log", []string{"logs"}, false},
			{"windows backslash path", `src\main.go`, []string{"*.go"}, true},
			{"windows backslash pattern", "src/main.go", []string{`src\*.go`}, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := Match(tt.path, tt.patterns)
				if err != nil {
					t.Fatalf("Match() error = %v", err)
				}
				if got != tt.want {
					t.Fatalf("Match(%q, %v) = %v, want %v", tt.path, tt.patterns, got, tt.want)
				}
			})
		}
	})

	t.Run("invalid glob syntax", func(t *testing.T) {
		_, err := Match("file.txt", []string{"[unclosed"})
		if err == nil {
			t.Fatal("expected error for invalid glob pattern")
		}
	})

	t.Run("depth semantics", func(t *testing.T) {
		tests := []struct {
			name     string
			path     string
			patterns []string
			want     bool
		}{
			{"basename any depth", "sub/app.exe", []string{"*.exe"}, true},
			{"root anchored slash", "app.exe", []string{"/*.exe"}, true},
			{"root anchored dot slash", "app.exe", []string{"./*.exe"}, true},
			{"root anchored excludes nested", "sub/app.exe", []string{"/*.exe"}, false},
			{"globstar exe any depth", "a/b/c.exe", []string{"**/*.exe"}, true},
			{"globstar exe root", "app.exe", []string{"**/*.exe"}, true},
			{"globstar all", "deep/nested/file.txt", []string{"**/*"}, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := Match(tt.path, tt.patterns)
				if err != nil {
					t.Fatalf("Match() error = %v", err)
				}
				if got != tt.want {
					t.Fatalf("Match(%q, %v) = %v, want %v", tt.path, tt.patterns, got, tt.want)
				}
			})
		}
	})
}

func TestGetFiles(t *testing.T) {
	setupTree := func(t *testing.T) string {
		t.Helper()
		root := t.TempDir()
		mustMkdir(t, filepath.Join(root, "src"))
		mustMkdir(t, filepath.Join(root, "logs"))
		mustWrite(t, filepath.Join(root, "src", "main.go"), "package main")
		mustWrite(t, filepath.Join(root, "src", "app.js"), "console.log()")
		mustWrite(t, filepath.Join(root, "readme.txt"), "hello")
		mustWrite(t, filepath.Join(root, "logs", "app.log"), "log")
		return root
	}

	t.Run("include only", func(t *testing.T) {
		root := setupTree(t)
		files, err := GetFiles(root, []string{"*.go", "*.js"}, nil)
		if err != nil {
			t.Fatalf("GetFiles() error = %v", err)
		}
		assertRelPaths(t, root, files, []string{"src/main.go", "src/app.js"})
	})

	t.Run("exclude only", func(t *testing.T) {
		root := setupTree(t)
		files, err := GetFiles(root, nil, []string{"logs/", "*.txt"})
		if err != nil {
			t.Fatalf("GetFiles() error = %v", err)
		}
		assertRelPaths(t, root, files, []string{"src/main.go", "src/app.js"})
	})

	t.Run("include and exclude", func(t *testing.T) {
		root := setupTree(t)
		files, err := GetFiles(root, []string{"*"}, []string{"*.js"})
		if err != nil {
			t.Fatalf("GetFiles() error = %v", err)
		}
		assertRelPaths(t, root, files, []string{"src/main.go", "readme.txt", "logs/app.log"})
	})

	t.Run("empty include includes all non-excluded", func(t *testing.T) {
		root := setupTree(t)
		files, err := GetFiles(root, nil, nil)
		if err != nil {
			t.Fatalf("GetFiles() error = %v", err)
		}
		assertRelPaths(t, root, files, []string{"src/main.go", "src/app.js", "readme.txt", "logs/app.log"})
	})

	t.Run("subdirectories", func(t *testing.T) {
		root := setupTree(t)
		files, err := GetFiles(root, []string{"src/*"}, nil)
		if err != nil {
			t.Fatalf("GetFiles() error = %v", err)
		}
		assertRelPaths(t, root, files, []string{"src/main.go", "src/app.js"})
	})

	t.Run("root only include", func(t *testing.T) {
		root := setupTree(t)
		mustWrite(t, filepath.Join(root, "app.exe"), "bin")
		mustMkdir(t, filepath.Join(root, "bin"))
		mustWrite(t, filepath.Join(root, "bin", "tool.exe"), "bin")

		files, err := GetFiles(root, []string{"./*.exe"}, nil)
		if err != nil {
			t.Fatalf("GetFiles() error = %v", err)
		}
		assertRelPaths(t, root, files, []string{"app.exe"})
	})

	t.Run("empty source directory", func(t *testing.T) {
		root := t.TempDir()
		files, err := GetFiles(root, nil, nil)
		if err != nil {
			t.Fatalf("GetFiles() error = %v", err)
		}
		if len(files) != 0 {
			t.Fatalf("expected no files, got %v", files)
		}
	})

	t.Run("nonexistent base path", func(t *testing.T) {
		_, err := GetFiles(filepath.Join(t.TempDir(), "missing"), nil, nil)
		if err == nil {
			t.Fatal("expected error for nonexistent base path")
		}
	})

	t.Run("walk access error", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("directory permission walk errors are not reliable on Windows")
		}
		root := t.TempDir()
		secret := filepath.Join(root, "secret")
		mustMkdir(t, secret)
		mustWrite(t, filepath.Join(secret, "data.txt"), "secret")
		if err := os.Chmod(secret, 0); err != nil {
			t.Fatalf("chmod: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(secret, 0o755) })

		_, err := GetFiles(root, nil, nil)
		if err == nil {
			t.Fatal("expected walk access error")
		}
	})
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertRelPaths(t *testing.T, base string, files, want []string) {
	t.Helper()
	got := make([]string, 0, len(files))
	for _, f := range files {
		rel, err := filepath.Rel(base, f)
		if err != nil {
			t.Fatalf("rel %s: %v", f, err)
		}
		got = append(got, filepath.ToSlash(rel))
	}
	if len(got) != len(want) {
		t.Fatalf("got %d files %v, want %d %v", len(got), got, len(want), want)
	}
	wantSet := make(map[string]struct{}, len(want))
	for _, w := range want {
		wantSet[w] = struct{}{}
	}
	for _, g := range got {
		if _, ok := wantSet[g]; !ok {
			t.Fatalf("unexpected file %q in %v", g, got)
		}
		delete(wantSet, g)
	}
	if len(wantSet) != 0 {
		t.Fatalf("missing files %v (got %v)", wantSet, got)
	}
}
