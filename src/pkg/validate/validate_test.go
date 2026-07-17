package validate

import (
	"strings"
	"testing"
)

func TestValidateSequences(t *testing.T) {
	t.Run("valid sequence preserves step order", func(t *testing.T) {
		content := []byte(`
name: seq-app
version: 1.0.0
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
          source:
            dir: ./dist
          target_path: ./dist/api-staging
      production:
        deploy:
          source:
            dir: ./dist
          target_path: ./dist/api-prod
  web:
    type: static
    environments:
      production:
        deploy:
          source:
            dir: ./dist
          target_path: ./dist/web
`)
		diags, cfg, err := ValidateYAML(content, ".")
		if err != nil {
			t.Fatalf("ValidateYAML() error = %v", err)
		}
		if HasErrors(diags) {
			t.Fatalf("unexpected errors: %+v", diags)
		}
		steps := cfg.Sequences["release"]
		want := []string{"api/staging", "api/production", "web/production"}
		if len(steps) != len(want) {
			t.Fatalf("steps len = %d, want %d", len(steps), len(want))
		}
		for i := range want {
			if steps[i] != want[i] {
				t.Fatalf("steps[%d] = %q, want %q", i, steps[i], want[i])
			}
		}
	})

	t.Run("empty sequence is error", func(t *testing.T) {
		content := []byte(`
name: seq-app
version: 1.0.0
sequences:
  empty: []
profiles:
  default:
    type: static
    environments:
      prod:
        deploy:
          source:
            dir: ./dist
          target_path: ./dist
`)
		diags, _, err := ValidateYAML(content, ".")
		if err != nil {
			t.Fatalf("ValidateYAML() error = %v", err)
		}
		if !HasErrors(diags) {
			t.Fatal("expected error for empty sequence")
		}
		if !containsMessage(diags, "at least one step") {
			t.Fatalf("expected empty-sequence message, got %+v", diags)
		}
	})

	t.Run("invalid step format is error", func(t *testing.T) {
		content := []byte(`
name: seq-app
version: 1.0.0
sequences:
  bad:
    - only-profile
profiles:
  default:
    type: static
    environments:
      prod:
        deploy:
          source:
            dir: ./dist
          target_path: ./dist
`)
		diags, _, err := ValidateYAML(content, ".")
		if err != nil {
			t.Fatalf("ValidateYAML() error = %v", err)
		}
		if !HasErrors(diags) {
			t.Fatal("expected error for invalid step format")
		}
	})

	t.Run("dot separator is error", func(t *testing.T) {
		content := []byte(`
name: seq-app
version: 1.0.0
sequences:
  bad:
    - default.prod
profiles:
  default:
    type: static
    environments:
      prod:
        deploy:
          source:
            dir: ./dist
          target_path: ./dist
`)
		diags, _, err := ValidateYAML(content, ".")
		if err != nil {
			t.Fatalf("ValidateYAML() error = %v", err)
		}
		if !HasErrors(diags) {
			t.Fatal("expected error for profile.env form")
		}
	})

	t.Run("missing profile is error", func(t *testing.T) {
		content := []byte(`
name: seq-app
version: 1.0.0
sequences:
  bad:
    - missing/prod
profiles:
  default:
    type: static
    environments:
      prod:
        deploy:
          source:
            dir: ./dist
          target_path: ./dist
`)
		diags, _, err := ValidateYAML(content, ".")
		if err != nil {
			t.Fatalf("ValidateYAML() error = %v", err)
		}
		if !HasErrors(diags) {
			t.Fatal("expected error for missing profile")
		}
		if !containsMessage(diags, `profile "missing" not found`) {
			t.Fatalf("expected missing profile message, got %+v", diags)
		}
	})

	t.Run("missing environment is error", func(t *testing.T) {
		content := []byte(`
name: seq-app
version: 1.0.0
sequences:
  bad:
    - default/staging
profiles:
  default:
    type: static
    environments:
      prod:
        deploy:
          source:
            dir: ./dist
          target_path: ./dist
`)
		diags, _, err := ValidateYAML(content, ".")
		if err != nil {
			t.Fatalf("ValidateYAML() error = %v", err)
		}
		if !HasErrors(diags) {
			t.Fatal("expected error for missing environment")
		}
		if !containsMessage(diags, `environment "staging" not found`) {
			t.Fatalf("expected missing environment message, got %+v", diags)
		}
	})
}

func TestValidateTypeGates(t *testing.T) {
	t.Run("static requires deploy.source", func(t *testing.T) {
		content := []byte(`
name: app
profiles:
  default:
    type: static
    environments:
      prod:
        deploy:
          target_path: ./dist
`)
		diags, _, err := ValidateYAML(content, ".")
		if err != nil {
			t.Fatalf("ValidateYAML() error = %v", err)
		}
		if !HasErrors(diags) || !containsMessage(diags, "deploy.source.dir is required") {
			t.Fatalf("expected source required error, got %+v", diags)
		}
	})

	t.Run("unknown field is error", func(t *testing.T) {
		content := []byte(`
name: app
profiles:
  default:
    type: static
    output_dir: ./dist
    environments:
      prod:
        deploy:
          source:
            dir: ./dist
          target_path: ./out
`)
		diags, _, err := ValidateYAML(content, ".")
		if err != nil {
			t.Fatalf("ValidateYAML() error = %v", err)
		}
		if !HasErrors(diags) || !containsMessage(diags, `unknown field "output_dir"`) {
			t.Fatalf("expected unknown field error, got %+v", diags)
		}
	})

	t.Run("remote credential required", func(t *testing.T) {
		content := []byte(`
name: app
profiles:
  default:
    type: static
    environments:
      prod:
        remote:
          host: example.com
        deploy:
          source:
            dir: ./dist
          target_path: ./out
`)
		diags, _, err := ValidateYAML(content, ".")
		if err != nil {
			t.Fatalf("ValidateYAML() error = %v", err)
		}
		if !HasErrors(diags) || !containsMessage(diags, "remote.credential is required") {
			t.Fatalf("expected credential required error, got %+v", diags)
		}
	})
}

func containsMessage(diags []Diagnostic, substr string) bool {
	for _, d := range diags {
		if strings.Contains(d.Message, substr) {
			return true
		}
	}
	return false
}
