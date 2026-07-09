package docker

import (
	"reflect"
	"testing"
)

func TestComposeUpArgs(t *testing.T) {
	tests := []struct {
		name    string
		file    string
		build   bool
		want    []string
	}{
		{
			name:  "without build",
			file:  "docker-compose.yml",
			build: false,
			want:  []string{"compose", "-f", "docker-compose.yml", "up", "-d"},
		},
		{
			name:  "with build",
			file:  "compose.prod.yml",
			build: true,
			want:  []string{"compose", "-f", "compose.prod.yml", "up", "-d", "--build"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := composeUpArgs(tt.file, tt.build)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("composeUpArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComposeDownArgs(t *testing.T) {
	want := []string{"compose", "-f", "docker-compose.yml", "down"}
	got := composeDownArgs("docker-compose.yml")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("composeDownArgs() = %v, want %v", got, want)
	}
}
