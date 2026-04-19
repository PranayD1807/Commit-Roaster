package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInitDoesNotModifyGitHooks acts as a regression test to guarantee
// that the application never tries to modify the filesystem or .git/hooks ever again.
func TestInitDoesNotModifyGitHooks(t *testing.T) {
	// 1. Setup mock file system
	tmpDir := t.TempDir()

	// Create a fake .git/hooks directory
	hooksDir := filepath.Join(tmpDir, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("Failed to create mock .git/hooks: %v", err)
	}

	// Create a fake existing hook to ensure it never gets touched
	fakeHookPath := filepath.Join(hooksDir, "commit-msg")
	originalHookContent := "#!/bin/sh\n# Existing Husky Hook\nexit 0"
	if err := os.WriteFile(fakeHookPath, []byte(originalHookContent), 0755); err != nil {
		t.Fatalf("Failed to write mock hook: %v", err)
	}

	// 2. Change working directory to our isolated temp space
	originalWD, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWD)

	// 3. Hijack stdout to capture the printed shell script
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// 4. Run the init command
	cmdInit()

	// 5. Restore stdout
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// 6. Verify everything is perfectly safe and isolated

	// A) Ensure it actually outputs the shell wrapper correctly
	if !strings.Contains(output, "git() {") || !strings.Contains(output, "commit-roaster roast") {
		t.Errorf("cmdInit did not output the expected shell script wrapper")
	}

	// B) PURE VERIFICATION: Assert that the .git/hooks directory was completely untouched
	content, err := os.ReadFile(fakeHookPath)
	if err != nil {
		t.Fatalf("Failed to read the existing hook after init: %v", err)
	}

	if string(content) != originalHookContent {
		t.Errorf("SECURITY/REGRESSION VIOLATION: cmdInit altered files in .git/hooks! Expected %q, got %q", originalHookContent, string(content))
	}

	// C) Check if it accidentally created any new files in the hooks directory
	files, _ := os.ReadDir(hooksDir)
	if len(files) > 1 {
		t.Errorf("SECURITY/REGRESSION VIOLATION: cmdInit created new rogue files in .git/hooks! Found %d files, expected 1.", len(files))
	}
}
