package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsWailsRunningAppToGenerateBindings(t *testing.T) {
	type test struct {
		name  string
		input []string
		want  bool
	}

	tests := []test{
		{
			name:  "Empty args should qualify as false",
			input: []string{},
			want:  false,
		},
		{
			name:  "No wailsbindings arg should qualify as false",
			input: []string{"arg1", "arg2"},
			want:  false,
		},
		{
			name:  "wailsbindings arg should qualify as true",
			input: []string{"/wailsbindings", "arg1", "arg2"},
			want:  true,
		},
		{
			name:  "wailsbindings arg should qualify as true even if it is not a path",
			input: []string{"wailsbindings", "arg1", "arg2"},
			want:  true,
		},
		{
			name:  "wailsbindings arg should qualify as false if it is not the first arg",
			input: []string{"arg1", "wailsbindings", "arg2"},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsWailsRunningAppToGenerateBindings(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetAppDataDir(t *testing.T) {
	type test struct {
		name         string
		pwd          string
		isDebugBuild bool
		want         string
	}

	homeDirectory, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []test{
		{
			name:         "Debug build should return relative path",
			pwd:          "/usr/opt/projects/build/bin/swervo_v1.32.1",
			isDebugBuild: true,
			want:         "/usr/opt/projects/build/bin/swervo_data",
		},
		{
			name:         "Debug build should return relative path when app is running from a macOS .app bundle",
			pwd:          "/usr/opt/projects/build/bin/swervo.app/Contents/MacOS/swervo_v1.32.1",
			isDebugBuild: true,
			want:         "/usr/opt/projects/build/bin/swervo_data",
		},
		{
			name:         "Release build should return the user's home directory",
			pwd:          "/home/testuser/swervo_v1.32.1",
			isDebugBuild: false,
			want:         filepath.Join(homeDirectory, ".swervo_data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetAppDataDir(tt.pwd, tt.isDebugBuild)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInitializeAppDataDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	type test struct {
		name         string
		appDataDir   string
		wantDirExist bool
	}

	tests := []test{
		{
			name:         "Should create the directory if it does not exist",
			appDataDir:   filepath.Join(tmpDir, "swervo_data"),
			wantDirExist: true,
		},
		{
			name:         "Should create the directory and all of its non existent parents if they do not exist",
			appDataDir:   filepath.Join(tmpDir, "my/fun/storage/location/swervo_data"),
			wantDirExist: true,
		},
		{
			name:         "Should not return an error if the directory already exists",
			appDataDir:   filepath.Join(tmpDir, "swervo_data"),
			wantDirExist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitializeAppDataDir(tt.appDataDir)

			if tt.wantDirExist {
				assert.DirExists(t, tt.appDataDir)
			} else {
				assert.NoDirExists(t, tt.appDataDir)
			}
		})
	}
}
