package hook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const hookScript = `#!/bin/sh
# commit-roaster — git commit-msg hook
# Roasts your commit messages so you don't have to roast yourself later.
if command -v commit-roaster >/dev/null 2>&1; then
    commit-roaster hook "$1"
fi
exit 0
`

// Install sets up the commit-msg git hook.
// If global is true, uses core.hooksPath for all repos.
// Otherwise installs in the current repo's .git/hooks/.
func Install(global bool) error {
	if global {
		return installGlobal()
	}
	return installLocal()
}

// Uninstall removes the commit-msg git hook.
func Uninstall(global bool) error {
	if global {
		return uninstallGlobal()
	}
	return uninstallLocal()
}

func installGlobal() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot find home directory: %w", err)
	}

	hooksDir := filepath.Join(home, ".commit-roaster", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("cannot create hooks directory: %w", err)
	}

	hookPath := filepath.Join(hooksDir, "commit-msg")

	// Warn if existing core.hooksPath is set to something else.
	existing, _ := gitConfig("core.hooksPath")
	existing = strings.TrimSpace(existing)
	if existing != "" && existing != hooksDir {
		fmt.Printf("  ⚠️  Existing core.hooksPath detected: %s\n", existing)
		fmt.Printf("     This will be overridden. Previous hooks may stop working.\n\n")
	}

	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		return fmt.Errorf("cannot write hook script: %w", err)
	}

	if err := setGitConfig("core.hooksPath", hooksDir); err != nil {
		return fmt.Errorf("cannot set git config: %w", err)
	}

	fmt.Printf("  ✅ commit-roaster installed globally\n")
	fmt.Printf("     Hook: %s\n", hookPath)
	fmt.Printf("     All new commits will be roasted. You're welcome.\n\n")
	return nil
}

func installLocal() error {
	gitDir, err := findGitDir()
	if err != nil {
		return err
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("cannot create hooks directory: %w", err)
	}

	hookPath := filepath.Join(hooksDir, "commit-msg")

	if _, err := os.Stat(hookPath); err == nil {
		fmt.Printf("  ⚠️  Existing commit-msg hook found. Backing up to commit-msg.backup\n")
		backupPath := hookPath + ".backup"
		data, _ := os.ReadFile(hookPath)
		os.WriteFile(backupPath, data, 0755)
	}

	if err := os.WriteFile(hookPath, []byte(hookScript), 0755); err != nil {
		return fmt.Errorf("cannot write hook script: %w", err)
	}

	fmt.Printf("  ✅ commit-roaster installed in this repo\n")
	fmt.Printf("     Hook: %s\n", hookPath)
	fmt.Printf("     Your commits in this repo will now be roasted.\n\n")
	return nil
}

func uninstallGlobal() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot find home directory: %w", err)
	}

	hookPath := filepath.Join(home, ".commit-roaster", "hooks", "commit-msg")
	os.Remove(hookPath)

	// Unset core.hooksPath only if it points to our directory.
	existing, _ := gitConfig("core.hooksPath")
	expected := filepath.Join(home, ".commit-roaster", "hooks")
	if strings.TrimSpace(existing) == expected {
		unsetGitConfig("core.hooksPath")
	}

	fmt.Printf("  ✅ commit-roaster uninstalled globally\n")
	fmt.Printf("     Your commits are safe from roasting. For now.\n\n")
	return nil
}

func uninstallLocal() error {
	gitDir, err := findGitDir()
	if err != nil {
		return err
	}

	hookPath := filepath.Join(gitDir, "hooks", "commit-msg")

	// Only remove if it's our hook.
	data, err := os.ReadFile(hookPath)
	if err != nil {
		return fmt.Errorf("no commit-msg hook found in this repo")
	}
	if !strings.Contains(string(data), "commit-roaster") {
		return fmt.Errorf("commit-msg hook exists but wasn't installed by commit-roaster")
	}

	os.Remove(hookPath)

	// Restore backup if it exists.
	backupPath := hookPath + ".backup"
	if _, err := os.Stat(backupPath); err == nil {
		os.Rename(backupPath, hookPath)
		fmt.Printf("  ℹ️  Restored previous commit-msg hook from backup.\n")
	}

	fmt.Printf("  ✅ commit-roaster uninstalled from this repo\n\n")
	return nil
}

func findGitDir() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--git-dir").Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository (or git is not installed)")
	}
	return strings.TrimSpace(string(out)), nil
}

func gitConfig(key string) (string, error) {
	out, err := exec.Command("git", "config", "--global", "--get", key).Output()
	return string(out), err
}

func setGitConfig(key, value string) error {
	return exec.Command("git", "config", "--global", key, value).Run()
}

func unsetGitConfig(key string) error {
	return exec.Command("git", "config", "--global", "--unset", key).Run()
}
