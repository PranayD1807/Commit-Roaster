package cmd_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PranayD1807/Commit-Roaster/cmd"
	"github.com/PranayD1807/Commit-Roaster/internal/airoaster"
	"github.com/PranayD1807/Commit-Roaster/internal/config"
)

func TestAIDisableByDefaultAndCommands(t *testing.T) {
	// Setup temporary home directory
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	// Save original args
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Mock Stdin & Stdout
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldStdin := os.Stdin
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		os.Stdin = oldStdin
	}()

	// 1. Verify show status (disabled by default)
	os.Args = []string{"commit-roaster", "ai", "status"}
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut
	cmd.CmdAI()
	wOut.Close()

	var buf bytes.Buffer
	io.Copy(&buf, rOut)
	statusOut := buf.String()
	if !strings.Contains(statusOut, "DISABLED") {
		t.Errorf("Expected status to be DISABLED by default, got:\n%s", statusOut)
	}

	// 2. Enable AI mode interactively
	os.Args = []string{"commit-roaster", "ai", "enable"}

	// Simulate user typing:
	// Choose Gemini: 1 [Enter]
	// Enter API Key: my-secret-key-123 [Enter]
	// Choose Gemini 2.5 Flash (index 1 of fallback list): 1 [Enter]
	userInput := "1\nmy-secret-key-123\n1\n"
	rIn, wIn, _ := os.Pipe()
	os.Stdin = rIn
	go func() {
		wIn.Write([]byte(userInput))
		wIn.Close()
	}()

	rOut, wOut, _ = os.Pipe()
	os.Stdout = wOut
	cmd.CmdAI()
	wOut.Close()

	buf.Reset()
	io.Copy(&buf, rOut)
	enableOut := buf.String()

	if !strings.Contains(enableOut, "AI roaster enabled with Google Gemini (gemini-3.1-flash-lite)") {
		t.Errorf("Unexpected output on enable: %s", enableOut)
	}

	// Verify status now shows ENABLED and masked key
	os.Args = []string{"commit-roaster", "ai", "status"}
	rOut, wOut, _ = os.Pipe()
	os.Stdout = wOut
	cmd.CmdAI()
	wOut.Close()

	buf.Reset()
	io.Copy(&buf, rOut)
	statusOut = buf.String()
	if !strings.Contains(statusOut, "ENABLED") || !strings.Contains(statusOut, "my-s****-123") {
		t.Errorf("Expected status to be ENABLED with masked key, got:\n%s", statusOut)
	}

	// 3. Disable AI mode
	os.Args = []string{"commit-roaster", "ai", "disable"}
	rOut, wOut, _ = os.Pipe()
	os.Stdout = wOut
	cmd.CmdAI()
	wOut.Close()

	buf.Reset()
	io.Copy(&buf, rOut)
	disableOut := buf.String()
	if !strings.Contains(disableOut, "disabled") {
		t.Errorf("Expected disable confirmation, got: %s", disableOut)
	}

	// Verify status is DISABLED again
	os.Args = []string{"commit-roaster", "ai", "status"}
	rOut, wOut, _ = os.Pipe()
	os.Stdout = wOut
	cmd.CmdAI()
	wOut.Close()

	buf.Reset()
	io.Copy(&buf, rOut)
	statusOut = buf.String()
	if !strings.Contains(statusOut, "DISABLED") {
		t.Errorf("Expected status to be DISABLED after disabling, got:\n%s", statusOut)
	}
}

func TestStatsPromptAndExecution(t *testing.T) {
	// Setup temporary home directory
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	// We must run in a mock git repo to avoid failing the git log commands
	tempRepo := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tempRepo)
	defer os.Chdir(originalWD)

	// Initialize git repo and create a mock commit
	_ = exec.Command("git", "init").Run()
	_ = exec.Command("git", "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "config", "user.name", "Test").Run()
	_ = os.WriteFile(filepath.Join(tempRepo, "a.txt"), []byte("a"), 0644)
	_ = exec.Command("git", "add", "a.txt").Run()
	_ = exec.Command("git", "commit", "-m", "Initial mock commit").Run()

	// Enable AI mode by saving a config directly
	cfg := config.Config{
		AIEnabled: true,
		Provider:  "gemini",
		Model:     "gemini-3.1-flash-lite",
		APIKey:    "test-key",
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Mock RoastCommitFn to return a static roast
	calledAI := false
	origRoastCommitFn := airoaster.RoastCommitFn
	defer func() { airoaster.RoastCommitFn = origRoastCommitFn }()

	airoaster.RoastCommitFn = func(provider, model, apiKey, commitMsg string) (string, error) {
		calledAI = true
		return "This is a sarcastic mock roast.", nil
	}

	// Mock Stdin/Stdout
	oldStdout := os.Stdout
	oldStdin := os.Stdin
	defer func() {
		os.Stdout = oldStdout
		os.Stdin = oldStdin
	}()

	// 1. Cancel stats prompt (typing 'n')
	rIn, wIn, _ := os.Pipe()
	os.Stdin = rIn
	go func() {
		wIn.Write([]byte("n\n"))
		wIn.Close()
	}()

	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut
	cmd.CmdStats()
	wOut.Close()

	var buf bytes.Buffer
	io.Copy(&buf, rOut)
	output := buf.String()

	if !strings.Contains(output, "analysis aborted") {
		t.Errorf("Expected stats to abort, but got:\n%s", output)
	}
	if calledAI {
		t.Errorf("AI should not have been called on aborted stats")
	}

	// 2. Accept stats prompt (typing 'y')
	calledAI = false
	rIn, wIn, _ = os.Pipe()
	os.Stdin = rIn
	go func() {
		wIn.Write([]byte("y\n"))
		wIn.Close()
	}()

	rOut, wOut, _ = os.Pipe()
	os.Stdout = wOut
	cmd.CmdStats()
	wOut.Close()

	buf.Reset()
	io.Copy(&buf, rOut)
	output = buf.String()

	if !strings.Contains(output, "Analyzing commits with AI...") || !strings.Contains(output, "Average score") {
		t.Errorf("Expected stats execution with AI, but got:\n%s", output)
	}
	if !calledAI {
		t.Errorf("AI should have been called on accepted stats")
	}
}

func TestRoastAndHookWithAI(t *testing.T) {
	// Setup temporary home directory
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	// Mock git repository for CmdRoast
	tempRepo := t.TempDir()
	originalWD, _ := os.Getwd()
	os.Chdir(tempRepo)
	defer os.Chdir(originalWD)

	_ = exec.Command("git", "init").Run()
	_ = exec.Command("git", "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "config", "user.name", "Test").Run()
	_ = os.WriteFile(filepath.Join(tempRepo, "a.txt"), []byte("a"), 0644)
	_ = exec.Command("git", "add", "a.txt").Run()
	_ = exec.Command("git", "commit", "-m", "fix: logic error in roast").Run()

	// 1. With AI enabled
	cfg := config.Config{
		AIEnabled: true,
		Provider:  "gemini",
		Model:     "gemini-3.1-flash-lite",
		APIKey:    "test-key",
	}
	_ = config.Save(cfg)

	// Mock RoastCommitFn
	aiReturn := "This is a sarcastic mock roast."
	origRoastCommitFn := airoaster.RoastCommitFn
	defer func() { airoaster.RoastCommitFn = origRoastCommitFn }()

	airoaster.RoastCommitFn = func(provider, model, apiKey, commitMsg string) (string, error) {
		return aiReturn, nil
	}

	// Capture stdout
	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut
	cmd.CmdRoast(1)
	wOut.Close()

	var buf bytes.Buffer
	io.Copy(&buf, rOut)
	output := buf.String()

	if !strings.Contains(output, "AI Roast") || !strings.Contains(output, "This is a sarcastic mock roast.") {
		t.Errorf("Expected AI Roast in CmdRoast output, got:\n%s", output)
	}

	// Test CLEAN response
	aiReturn = "CLEAN"
	rOut, wOut, _ = os.Pipe()
	os.Stdout = wOut
	cmd.CmdRoast(1)
	wOut.Close()

	buf.Reset()
	io.Copy(&buf, rOut)
	output = buf.String()

	if !strings.Contains(output, "✓") || strings.Contains(output, "AI Roast") {
		t.Errorf("Expected clean commit checkmark output, got:\n%s", output)
	}

	// Test CmdHook with AI enabled
	msgFile := filepath.Join(tempRepo, "commit-msg.txt")
	_ = os.WriteFile(msgFile, []byte("some bad commit message"), 0644)

	aiReturn = "Mock hook roast from Claude."
	rOut, wOut, _ = os.Pipe()
	os.Stdout = wOut
	cmd.CmdHook(msgFile)
	wOut.Close()

	buf.Reset()
	io.Copy(&buf, rOut)
	output = buf.String()

	if !strings.Contains(output, "AI Roast") || !strings.Contains(output, "Mock hook roast from Claude.") {
		t.Errorf("Expected hook roast output, got:\n%s", output)
	}

	// Test CmdHook with CLEAN message
	aiReturn = "CLEAN"
	rOut, wOut, _ = os.Pipe()
	os.Stdout = wOut
	cmd.CmdHook(msgFile)
	wOut.Close()

	buf.Reset()
	io.Copy(&buf, rOut)
	output = buf.String()

	if strings.TrimSpace(output) != "" {
		t.Errorf("Expected empty hook output for CLEAN message, got:\n%s", output)
	}
}
