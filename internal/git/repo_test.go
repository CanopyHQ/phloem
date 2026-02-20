package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRemoteURL_HTTPS(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "HTTPS with .git",
			url:       "https://github.com/CanopyHQ/canopy.git",
			wantOwner: "CanopyHQ",
			wantRepo:  "canopy",
			wantErr:   false,
		},
		{
			name:      "HTTPS without .git",
			url:       "https://github.com/CanopyHQ/canopy",
			wantOwner: "CanopyHQ",
			wantRepo:  "canopy",
			wantErr:   false,
		},
		{
			name:      "HTTP with .git",
			url:       "http://github.com/user/repo.git",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseRemoteURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantOwner, owner)
				assert.Equal(t, tt.wantRepo, repo)
			}
		})
	}
}

func TestParseRemoteURL_SSH(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "SSH with .git",
			url:       "git@github.com:CanopyHQ/canopy.git",
			wantOwner: "CanopyHQ",
			wantRepo:  "canopy",
			wantErr:   false,
		},
		{
			name:      "SSH without .git",
			url:       "git@github.com:user/repo",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseRemoteURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantOwner, owner)
				assert.Equal(t, tt.wantRepo, repo)
			}
		})
	}
}

func TestParseRemoteURL_Invalid(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"empty", ""},
		{"invalid SSH", "git@github.com/invalid"},
		{"invalid HTTPS", "https://github.com/invalid"},
		{"unsupported protocol", "ftp://github.com/user/repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := parseRemoteURL(tt.url)
			assert.Error(t, err)
		})
	}
}

func TestRepository_Scope(t *testing.T) {
	repo := &Repository{
		Owner: "CanopyHQ",
		Name:  "canopy",
	}

	assert.Equal(t, "github.com/CanopyHQ/canopy", repo.Scope())
}
