package inspect

import (
	"testing"

	"pablo/pkg/domain"
)

func TestFromConfigSequences(t *testing.T) {
	cfg := &domain.Config{
		Name:    "seq-app",
		Version: "1.0.0",
		Profiles: map[string]domain.Profile{
			"api": {
				Type: "static",
				Environments: map[string]domain.Environment{
					"staging": {},
					"prod":    {},
				},
			},
		},
		Sequences: map[string][]string{
			"release": {"api/staging", "api/prod"},
			"alpha":   {"api/prod"},
		},
	}

	result := FromConfig(cfg)
	if len(result.Sequences) != 2 {
		t.Fatalf("sequences len = %d, want 2", len(result.Sequences))
	}

	// Sequence names are sorted for stable inspect output; steps keep manifest order.
	if result.Sequences[0].Name != "alpha" {
		t.Fatalf("sequences[0].Name = %q, want alpha", result.Sequences[0].Name)
	}
	if result.Sequences[1].Name != "release" {
		t.Fatalf("sequences[1].Name = %q, want release", result.Sequences[1].Name)
	}

	steps := result.Sequences[1].Steps
	if len(steps) != 2 || steps[0] != "api/staging" || steps[1] != "api/prod" {
		t.Fatalf("release steps = %#v, want [api/staging api/prod]", steps)
	}
}

func TestFromYAMLSequences(t *testing.T) {
	content := []byte(`
name: seq-app
version: 1.0.0
sequences:
  release:
    - api/staging
    - api/prod
profiles:
  api:
    type: static
    environments:
      staging:
        deploy:
          source:
            dir: ./dist
          target_path: ./a
      prod:
        deploy:
          source:
            dir: ./dist
          target_path: ./b
`)
	result, err := FromYAML(content, ".")
	if err != nil {
		t.Fatalf("FromYAML() error = %v", err)
	}
	if len(result.Sequences) != 1 {
		t.Fatalf("sequences len = %d, want 1", len(result.Sequences))
	}
	if result.Sequences[0].Name != "release" {
		t.Fatalf("Name = %q, want release", result.Sequences[0].Name)
	}
	steps := result.Sequences[0].Steps
	if len(steps) != 2 || steps[0] != "api/staging" || steps[1] != "api/prod" {
		t.Fatalf("steps = %#v", steps)
	}
}
