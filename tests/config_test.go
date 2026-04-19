package cmd_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/PranayD1807/Commit-Roaster/internal/config"
)

func TestConfigRoundTrip(t *testing.T) {
	// Set up temporary home directory
	tempDir, err := os.MkdirTemp("", "commit-roaster-test-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Override HOME/USERPROFILE env var
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", tempDir)
	} else {
		t.Setenv("HOME", tempDir)
	}

	// 1. Verify loading default config when file is missing
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load should not fail for missing file: %v", err)
	}
	if cfg.AIEnabled {
		t.Errorf("Expected AIEnabled to be false, got true")
	}

	// 2. Save config
	testCfg := config.Config{
		AIEnabled: true,
		Provider:  "gemini",
		Model:     "gemini-2.5-flash",
		APIKey:    "test-api-key-12345",
	}

	if err := config.Save(testCfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// 3. Verify permissions on Unix-like systems
	if runtime.GOOS != "windows" {
		path, err := config.ConfigPath()
		if err != nil {
			t.Fatalf("Failed to get config path: %v", err)
		}

		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Failed to stat config file: %v", err)
		}

		mode := info.Mode().Perm()
		if mode != 0600 {
			t.Errorf("Expected config file permissions to be 0600, got %o", mode)
		}

		dir, err := config.ConfigDir()
		if err != nil {
			t.Fatalf("Failed to get config dir: %v", err)
		}
		dirInfo, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("Failed to stat config dir: %v", err)
		}
		dirMode := dirInfo.Mode().Perm()
		if dirMode != 0700 {
			t.Errorf("Expected config dir permissions to be 0700, got %o", dirMode)
		}
	}

	// 4. Load it back and verify values
	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loaded.AIEnabled != testCfg.AIEnabled ||
		loaded.Provider != testCfg.Provider ||
		loaded.Model != testCfg.Model ||
		loaded.APIKey != testCfg.APIKey {
		t.Errorf("Loaded config %+v does not match saved config %+v", loaded, testCfg)
	}
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"", "****"},
		{"123", "****"},
		{"12345678", "****"},
		{"123456789", "1234****6789"},
		{"abcdefghijkl", "abcd****ijkl"},
	}

	for _, tc := range tests {
		got := config.MaskKey(tc.key)
		if got != tc.expected {
			t.Errorf("MaskKey(%q) = %q; expected %q", tc.key, got, tc.expected)
		}
	}
}

func TestCorruptConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "commit-roaster-test-home-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", tempDir)
	} else {
		t.Setenv("HOME", tempDir)
	}

	// Write invalid JSON to config file
	dir, _ := config.ConfigDir()
	_ = os.MkdirAll(dir, 0700)
	path, _ := config.ConfigPath()
	_ = os.WriteFile(path, []byte("invalid json"), 0600)

	_, err = config.Load()
	if err == nil {
		t.Errorf("Expected error when loading corrupt config, got nil")
	}
}
