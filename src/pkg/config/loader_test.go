package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeManifest(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}

func TestLoad(t *testing.T) {
	loader := NewLoader()

	t.Run("happy path minimal manifest", func(t *testing.T) {
		dir := t.TempDir()
		path := writeManifest(t, dir, "pablo.yaml", `
name: test-app
profiles:
  default:
    type: static
    environments:
      prod:
        deploy:
          target_path: /var/www
`)
		cfg, err := loader.Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.Name != "test-app" {
			t.Fatalf("Name = %q, want test-app", cfg.Name)
		}
		if cfg.BaseDir != dir {
			t.Fatalf("BaseDir = %q, want %q", cfg.BaseDir, dir)
		}
	})

	t.Run("inherit profile variables and env_file", func(t *testing.T) {
		dir := t.TempDir()
		path := writeManifest(t, dir, "pablo.yaml", `
profiles:
  default:
    type: static
    variables:
      SHARED: from-profile
    env_file: .env.profile
    environments:
      prod:
        variables:
          ENV_ONLY: local
        deploy:
          target_path: /var/www
`)
		cfg, err := loader.Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		env := cfg.Profiles["default"].Environments["prod"]
		if env.EnvConfig.Variables["SHARED"] != "from-profile" {
			t.Fatalf("SHARED = %q, want from-profile", env.EnvConfig.Variables["SHARED"])
		}
		if env.EnvConfig.Variables["ENV_ONLY"] != "local" {
			t.Fatalf("ENV_ONLY = %q, want local", env.EnvConfig.Variables["ENV_ONLY"])
		}
		if env.EnvConfig.EnvFile != ".env.profile" {
			t.Fatalf("EnvFile = %q, want .env.profile", env.EnvConfig.EnvFile)
		}
	})

	t.Run("inherit profile build to environment", func(t *testing.T) {
		dir := t.TempDir()
		path := writeManifest(t, dir, "pablo.yaml", `
profiles:
  default:
    type: binary
    build:
      command: make all
      path: ./src
      variables:
        GOOS: linux
    environments:
      prod:
        deploy:
          target_path: /opt/app
`)
		cfg, err := loader.Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		env := cfg.Profiles["default"].Environments["prod"]
		if env.Build == nil {
			t.Fatal("expected inherited build config")
		}
		if env.Build.Command != "make all" || env.Build.Path != "./src" {
			t.Fatalf("build = %+v, want command=make all path=./src", env.Build)
		}
		if env.Build.Variables["GOOS"] != "linux" {
			t.Fatalf("GOOS = %q, want linux", env.Build.Variables["GOOS"])
		}
	})

	t.Run("merge profile build with environment override", func(t *testing.T) {
		dir := t.TempDir()
		path := writeManifest(t, dir, "pablo.yaml", `
profiles:
  default:
    type: binary
    build:
      command: make all
      path: ./src
      variables:
        GOOS: linux
        CGO: "0"
    environments:
      prod:
        build:
          command: make prod
          variables:
            CGO: "1"
        deploy:
          target_path: /opt/app
`)
		cfg, err := loader.Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		env := cfg.Profiles["default"].Environments["prod"]
		if env.Build.Command != "make prod" {
			t.Fatalf("Command = %q, want make prod", env.Build.Command)
		}
		if env.Build.Path != "./src" {
			t.Fatalf("Path = %q, want ./src", env.Build.Path)
		}
		if env.Build.Variables["GOOS"] != "linux" {
			t.Fatalf("GOOS = %q, want linux", env.Build.Variables["GOOS"])
		}
		if env.Build.Variables["CGO"] != "1" {
			t.Fatalf("CGO = %q, want 1", env.Build.Variables["CGO"])
		}
	})

	t.Run("propagate output_dir to deploy source", func(t *testing.T) {
		dir := t.TempDir()
		path := writeManifest(t, dir, "pablo.yaml", `
profiles:
  default:
    type: static
    output_dir:
      dir: ./dist
      include: ["*.js"]
      exclude: ["*.map"]
    environments:
      prod:
        deploy:
          target_path: /var/www
`)
		cfg, err := loader.Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		src := cfg.Profiles["default"].Environments["prod"].Deploy.Source
		if src == nil {
			t.Fatal("expected deploy source from output_dir")
		}
		if src.Dir != "./dist" || len(src.Include) != 1 || src.Include[0] != "*.js" {
			t.Fatalf("source = %+v", src)
		}
	})

	t.Run("propagate environment variables to deploy level", func(t *testing.T) {
		dir := t.TempDir()
		path := writeManifest(t, dir, "pablo.yaml", `
profiles:
  default:
    type: static
    environments:
      prod:
        variables:
          API_URL: https://api.example.com
        env_file: .env.deploy
        deploy:
          target_path: /var/www
`)
		cfg, err := loader.Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		deploy := cfg.Profiles["default"].Environments["prod"].Deploy
		if deploy.EnvConfig.Variables["API_URL"] != "https://api.example.com" {
			t.Fatalf("API_URL = %q", deploy.EnvConfig.Variables["API_URL"])
		}
		if deploy.EnvConfig.EnvFile != ".env.deploy" {
			t.Fatalf("EnvFile = %q, want .env.deploy", deploy.EnvConfig.EnvFile)
		}
	})

	t.Run("legacy top level type and environments", func(t *testing.T) {
		dir := t.TempDir()
		path := writeManifest(t, dir, "pablo.yaml", `
type: static
source:
  include: ["*.html"]
  exclude: ["*.tmp"]
environments:
  prod:
    target_path: /var/www
    strategy: overwrite
    variables:
      MODE: prod
`)
		cfg, err := loader.Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		profile, ok := cfg.Profiles["default"]
		if !ok {
			t.Fatal("expected default profile from legacy manifest")
		}
		if profile.Type != "static" {
			t.Fatalf("Type = %q, want static", profile.Type)
		}
		if len(profile.OutputDir.Include) != 1 || profile.OutputDir.Include[0] != "*.html" {
			t.Fatalf("OutputDir.Include = %v", profile.OutputDir.Include)
		}
		env := profile.Environments["prod"]
		if env.Deploy.TargetPath != "/var/www" || env.Deploy.Strategy != "overwrite" {
			t.Fatalf("deploy = %+v", env.Deploy)
		}
		if env.EnvConfig.Variables["MODE"] != "prod" {
			t.Fatalf("MODE = %q, want prod", env.EnvConfig.Variables["MODE"])
		}
	})

	t.Run("sequences preserve step order", func(t *testing.T) {
		dir := t.TempDir()
		path := writeManifest(t, dir, "pablo.yaml", `
name: seq-app
sequences:
  release:
    - api/staging
    - api/production
    - web/production
profiles:
  api:
    type: static
    environments:
      staging:
        deploy:
          target_path: ./a
      production:
        deploy:
          target_path: ./b
  web:
    type: static
    environments:
      production:
        deploy:
          target_path: ./c
`)
		cfg, err := loader.Load(path)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		steps := cfg.Sequences["release"]
		want := []string{"api/staging", "api/production", "web/production"}
		if len(steps) != len(want) {
			t.Fatalf("steps len = %d, want %d (%v)", len(steps), len(want), steps)
		}
		for i := range want {
			if steps[i] != want[i] {
				t.Fatalf("steps[%d] = %q, want %q", i, steps[i], want[i])
			}
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := loader.Load(filepath.Join(t.TempDir(), "missing.yaml"))
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		dir := t.TempDir()
		path := writeManifest(t, dir, "pablo.yaml", "name: [unclosed")
		_, err := loader.Load(path)
		if err == nil {
			t.Fatal("expected error for invalid YAML")
		}
	})
}
