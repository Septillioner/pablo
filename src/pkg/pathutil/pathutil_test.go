package pathutil

import "testing"

func TestJoinRemote(t *testing.T) {
	tests := []struct {
		name string
		elem []string
		want string
	}{
		{
			name: "multiple segments",
			elem: []string{"var", "www", "app"},
			want: "var/www/app",
		},
		{
			name: "empty elements skipped",
			elem: []string{"var", "", "www"},
			want: "var/www",
		},
		{
			name: "backslash normalized",
			elem: []string{`var\www`, "app"},
			want: "var/www/app",
		},
		{
			name: "duplicate slashes collapsed",
			elem: []string{"var//www", "app"},
			want: "var/www/app",
		},
		{
			name: "absolute path preserved",
			elem: []string{"/var", "www"},
			want: "/var/www",
		},
		{
			name: "all empty",
			elem: []string{"", ""},
			want: "",
		},
		{
			name: "single absolute segment",
			elem: []string{"/var/www"},
			want: "/var/www",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JoinRemote(tt.elem...)
			if got != tt.want {
				t.Fatalf("JoinRemote(%v) = %q, want %q", tt.elem, got, tt.want)
			}
		})
	}
}

func TestDirRemote(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "nested path",
			path: "/var/www/app/index.html",
			want: "/var/www/app",
		},
		{
			name: "root slash",
			path: "/",
			want: "/",
		},
		{
			name: "single segment",
			path: "app",
			want: ".",
		},
		{
			name: "file without extension",
			path: "/var/www/README",
			want: "/var/www",
		},
		{
			name: "backslash normalized",
			path: `/var\www\app`,
			want: "/var/www",
		},
		{
			name: "absolute two segments",
			path: "/var/www",
			want: "/var",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DirRemote(tt.path)
			if got != tt.want {
				t.Fatalf("DirRemote(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
