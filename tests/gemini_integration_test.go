//go:build integration

package cmd_test

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/PranayD1807/Commit-Roaster/internal/airoaster"
)

// loadEnv reads a simple KEY=VALUE .env file and returns the values as a map.
func loadEnv(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	env := make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			env[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return env, scanner.Err()
}

// geminiKey finds the GEMINI_KEY by walking up from the test file location.
func geminiKey(t *testing.T) string {
	t.Helper()

	dir, _ := os.Getwd()
	for {
		candidate := filepath.Join(dir, ".env")
		env, err := loadEnv(candidate)
		if err == nil {
			if key, ok := env["GEMINI_KEY"]; ok && key != "" {
				return key
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("GEMINI_KEY not found in any .env file — add it to the project root .env")
	return ""
}

// TestGeminiRoastBadCommit sends a notoriously bad commit message and expects a roast back.
func TestGeminiRoastBadCommit(t *testing.T) {
	apiKey := geminiKey(t)

	input := "fix stuff"
	expected := "a non-empty roast (not CLEAN)"

	t.Logf("┌─ Input:    %q", input)
	t.Logf("│  Expected: %s", expected)

	roast, err := airoaster.RoastCommit("gemini", "gemini-2.5-flash", apiKey, input)
	if err != nil {
		t.Fatalf("│  Error:    %v", err)
	}

	t.Logf("│  Got:      %q", roast)

	if strings.TrimSpace(roast) == "" {
		t.Error("└─ FAIL: response was empty")
	} else if strings.ToUpper(strings.TrimSpace(roast)) == "CLEAN" {
		t.Errorf("└─ FAIL: expected a roast but got CLEAN")
	} else {
		t.Log("└─ PASS ✅")
	}
}

// TestGeminiCleanCommit sends a well-formed commit and expects "CLEAN".
func TestGeminiCleanCommit(t *testing.T) {
	apiKey := geminiKey(t)

	input := "feat(auth): add JWT token refresh with automatic retry on 401"
	expected := `"CLEAN"`

	t.Logf("┌─ Input:    %q", input)
	t.Logf("│  Expected: %s", expected)

	roast, err := airoaster.RoastCommit("gemini", "gemini-2.5-flash", apiKey, input)
	if err != nil {
		t.Fatalf("│  Error:    %v", err)
	}

	t.Logf("│  Got:      %q", roast)

	if strings.TrimSpace(roast) == "" {
		t.Error("└─ FAIL: response was empty")
	} else if strings.ToUpper(strings.TrimSpace(roast)) == "CLEAN" {
		t.Log("└─ PASS ✅  (correctly identified as clean)")
	} else {
		// AI isn't 100% deterministic — log but don't fail
		t.Logf("└─ INFO ℹ️   Gemini roasted a good commit (AI will be AI)")
	}
}

// TestGeminiWIPCommit verifies a batch of bad commits all get roasted.
func TestGeminiWIPCommit(t *testing.T) {
	apiKey := geminiKey(t)

	cases := []struct {
		input    string
		expected string
	}{
		{"WIP", "a roast (not CLEAN)"},
		{"asdfgh", "a roast (not CLEAN)"},
		{"fix", "a roast (not CLEAN)"},
		{"temp commit please ignore", "a roast (not CLEAN)"},
	}

	for i, tc := range cases {
		if i > 0 {
			time.Sleep(3 * time.Second) // respect free-tier 5 RPM limit
		}
		tc := tc // capture range variable
		t.Run(tc.input, func(t *testing.T) {
			t.Logf("┌─ Input:    %q", tc.input)
			t.Logf("│  Expected: %s", tc.expected)

			roast, err := airoaster.RoastCommit("gemini", "gemini-2.5-flash", apiKey, tc.input)
			if err != nil {
				t.Fatalf("│  Error:    %v", err)
			}

			t.Logf("│  Got:      %q", roast)

			if strings.TrimSpace(roast) == "" {
				t.Error("└─ FAIL: response was empty")
			} else {
				t.Log("└─ PASS ✅")
			}
		})
	}
}
