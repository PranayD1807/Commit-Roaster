package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
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

	// A) Ensure it actually outputs a shell wrapper correctly (bash or powershell depending on OS)
	if !strings.Contains(output, "commit-roaster roast") {
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

func TestCleanLegacyHooks(t *testing.T) {
	// 1. Setup isolated directories
	tmpHome := t.TempDir()
	tmpRepo := t.TempDir()
	
	// Hijack HOME so global git config goes to tmpHome instead of mutating the dev's real environment
	// Windows uses USERPROFILE for os.UserHomeDir(), Unix uses HOME
	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", tmpHome)

	// Create mocked global ~/.commit-roaster folder
	globalHooksDir := filepath.Join(tmpHome, ".commit-roaster", "hooks")
	if err := os.MkdirAll(globalHooksDir, 0755); err != nil {
		t.Fatalf("Failed to create global hooks dir: %v", err)
	}

	// Mock global git config using the safely overridden HOME
	cmd := exec.Command("git", "config", "--global", "core.hooksPath", globalHooksDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to setup mock global git config: %v", err)
	}

	// 2. Setup mock local repo
	originalWD, _ := os.Getwd()
	os.Chdir(tmpRepo)
	defer os.Chdir(originalWD)

	cmdInit := exec.Command("git", "init")
	if err := cmdInit.Run(); err != nil {
		t.Fatalf("Failed to initialize mock git repo: %v", err)
	}

	gitHooksDir := filepath.Join(tmpRepo, ".git", "hooks")
	os.MkdirAll(gitHooksDir, 0755)

	hookPath := filepath.Join(gitHooksDir, "commit-msg")
	backupPath := hookPath + ".backup"

	// Write a mock legacy hook containing the "commit-roaster" signature
	if err := os.WriteFile(hookPath, []byte("#!/bin/sh\ncommit-roaster hook \"$1\""), 0755); err != nil {
		t.Fatalf("Failed to write mock hook: %v", err)
	}
	
	// Write a mock backup of an original pre-commit/husky hook
	if err := os.WriteFile(backupPath, []byte("#!/bin/sh\n# original hook stuff"), 0755); err != nil {
		t.Fatalf("Failed to write backup hook: %v", err)
	}

	// 3. EXECUTE the cleanup logic
	cleanLegacyHooks()

	// 4. Verify global cleanup
	out, err := exec.Command("git", "config", "--global", "--get", "core.hooksPath").Output()
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		t.Errorf("Expected core.hooksPath to be unset globally, got: %s", string(out))
	}
	
	if _, err := os.Stat(filepath.Join(tmpHome, ".commit-roaster")); !os.IsNotExist(err) {
		t.Errorf("Expected ~/.commit-roaster directory to be completely deleted")
	}

	// 5. Verify local cleanup
	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("Restored hook file is missing. Expected .backup to be moved here: %v", err)
	}

	if string(data) != "#!/bin/sh\n# original hook stuff" {
		t.Errorf("Expected backup to be restored as the live hook, got: %s", string(data))
	}

	if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
		t.Errorf("Expected the .backup file to be moved/deleted after restoration")
	}
}

func TestCmdInitOutput(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	tests := []struct {
		name     string
		args     []string
		envShell string
		expected string
	}{
		{"Explicit Fish", []string{"commit-roaster", "init", "fish"}, "", "function git"},
		{"Explicit PowerShell", []string{"commit-roaster", "init", "powershell"}, "", "$LASTEXITCODE"},
		{"Explicit PWSH", []string{"commit-roaster", "init", "pwsh"}, "", "Invoke-Expression"},
		{"Implicit Zsh via Env", []string{"commit-roaster", "init"}, "/bin/zsh", "local ext_code=$?"},
		{"Implicit Fish via Env", []string{"commit-roaster", "init"}, "/usr/bin/fish", "set -l ext_code $status"},
		{"Explicit Bash", []string{"commit-roaster", "init", "bash"}, "", "git() {"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			t.Setenv("SHELL", tt.envShell)

			// Safely hijack stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			cmdInit()

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, but got syntax:\n%s", tt.expected, output)
			}
		})
	}
}

func TestSetupAndTeardown(t *testing.T) {
	tmpHome := t.TempDir()

	t.Setenv("HOME", tmpHome)
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("SHELL", "/bin/zsh")

	// 1. Run setup — should create .zshrc with the eval line
	// Capture stdout (setup prints to stdout)
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmdSetup()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	profilePath := filepath.Join(tmpHome, ".zshrc")
	data, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("Expected .zshrc to be created, got error: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, shellMarker) {
		t.Errorf("Expected shell marker in profile, got:\n%s", content)
	}
	if !strings.Contains(content, `eval "$(commit-roaster init)"`) {
		t.Errorf("Expected eval line in profile, got:\n%s", content)
	}

	// 2. Run setup again — should be idempotent
	oldStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w

	cmdSetup()

	w.Close()
	os.Stdout = oldStdout

	buf.Reset()
	io.Copy(&buf, r)

	data2, _ := os.ReadFile(profilePath)
	if strings.Count(string(data2), shellMarker) != 1 {
		t.Errorf("Setup ran twice but marker appeared %d times (expected 1)", strings.Count(string(data2), shellMarker))
	}

	// 3. Add some user content before and after to make sure teardown preserves it
	userContent := "# my custom alias\nalias ll='ls -la'\n"
	os.WriteFile(profilePath, []byte(userContent+string(data)), 0644)

	// 4. Run teardown
	oldStdout = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w

	cmdTeardown()

	w.Close()
	os.Stdout = oldStdout

	buf.Reset()
	io.Copy(&buf, r)

	finalData, err := os.ReadFile(profilePath)
	if err != nil {
		t.Fatalf("Profile file disappeared after teardown: %v", err)
	}

	finalContent := string(finalData)
	if strings.Contains(finalContent, shellMarker) {
		t.Errorf("Teardown failed to remove shell marker")
	}
	if strings.Contains(finalContent, "commit-roaster init") {
		t.Errorf("Teardown failed to remove eval line")
	}
	if !strings.Contains(finalContent, "alias ll='ls -la'") {
		t.Errorf("Teardown destroyed user content! Got:\n%s", finalContent)
	}
}
