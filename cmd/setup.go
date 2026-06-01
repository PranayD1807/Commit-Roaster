package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const ShellMarker = "# commit-roaster shell integration"

// CleanLegacyHooks removes any filesystem modifications made by versions < 1.0.0
func CleanLegacyHooks() {
	// 1. Clean Global core.hooksPath and ~/.commit-roaster folder
	home, err := os.UserHomeDir()
	if err == nil {
		expected := filepath.Join(home, ".commit-roaster", "hooks")
		out, _ := exec.Command("git", "config", "--global", "--get", "core.hooksPath").Output()
		if strings.TrimSpace(string(out)) == expected {
			exec.Command("git", "config", "--global", "--unset", "core.hooksPath").Run()
		}
		os.RemoveAll(filepath.Join(home, ".commit-roaster"))
	}

	// 2. Clean Local repository hook modifications
	out, err := exec.Command("git", "rev-parse", "--git-dir").Output()
	if err == nil {
		gitDir := strings.TrimSpace(string(out))
		hookPath := filepath.Join(gitDir, "hooks", "commit-msg")
		data, err := os.ReadFile(hookPath)
		if err == nil && strings.Contains(string(data), "commit-roaster") {
			os.Remove(hookPath)
			// Restore the previous hook if a backup was made
			backupPath := hookPath + ".backup"
			if _, err := os.Stat(backupPath); err == nil {
				os.Rename(backupPath, hookPath)
			}
		}
	}
}

// CmdInit outputs the shell integration script.
func CmdInit() {
	shell := "bash"
	if len(os.Args) >= 3 {
		shell = strings.ToLower(os.Args[2])
	} else if envShell := os.Getenv("SHELL"); envShell != "" {
		if strings.Contains(envShell, "fish") {
			shell = "fish"
		} else if strings.Contains(envShell, "zsh") {
			shell = "zsh"
		}
	} else if runtime.GOOS == "windows" {
		shell = "powershell"
	}

	var script string

	switch shell {
	case "fish":
		script = `
# commit-roaster integration for Fish
# Add to ~/.config/fish/config.fish: commit-roaster init fish | source
function git
  command git $argv
  set -l ext_code $status
  if test "$argv[1]" = "commit"; and test $ext_code -eq 0
    commit-roaster roast --last 1
  end
  return $ext_code
end`
	case "powershell", "pwsh", "ps":
		script = `
# commit-roaster integration for PowerShell
# Add to your $PROFILE: Invoke-Expression (&commit-roaster init powershell | Out-String)
function git {
    & git.exe @args
    $ext_code = $LASTEXITCODE
    if ($args.Count -gt 0 -and $args[0] -eq "commit" -and $ext_code -eq 0) {
        commit-roaster roast --last 1
    }
    exit $ext_code
}`
	default:
		// bash, zsh, sh
		script = `
# commit-roaster integration for Bash/Zsh
# Add to ~/.zshrc or ~/.bashrc: eval "$(commit-roaster init)"
git() {
  command git "$@"
  local ext_code=$?
  if [ "$1" = "commit" ] && [ $ext_code -eq 0 ]; then
    commit-roaster roast --last 1
  fi
  return $ext_code
}`
	}

	fmt.Println(strings.TrimSpace(script))
}

// DetectShellProfile returns the shell type and profile path for the current user.
func DetectShellProfile() (shellType string, profilePath string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "bash", ""
	}

	envShell := os.Getenv("SHELL")

	if strings.Contains(envShell, "fish") {
		return "fish", filepath.Join(home, ".config", "fish", "config.fish")
	}
	if strings.Contains(envShell, "zsh") {
		return "zsh", filepath.Join(home, ".zshrc")
	}
	if strings.Contains(envShell, "bash") || strings.Contains(envShell, "sh") {
		profile := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(profile); os.IsNotExist(err) {
			profile = filepath.Join(home, ".bash_profile")
		}
		return "bash", profile
	}

	if runtime.GOOS == "windows" {
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "echo $PROFILE").Output()
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			return "powershell", strings.TrimSpace(string(out))
		}
		return "powershell", filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	}

	profile := filepath.Join(home, ".bashrc")
	if _, err := os.Stat(profile); os.IsNotExist(err) {
		profile = filepath.Join(home, ".bash_profile")
	}
	return "bash", profile
}

// EvalLine returns the line to append to the shell profile.
func EvalLine(shellType string) string {
	switch shellType {
	case "fish":
		return "commit-roaster init fish | source"
	case "powershell", "pwsh":
		return "Invoke-Expression (&commit-roaster init powershell | Out-String)"
	default:
		return `eval "$(commit-roaster init)"`
	}
}

// CmdSetup automatically appends the shell integration line to the user's profile.
func CmdSetup() {
	shellType, profilePath := DetectShellProfile()

	if profilePath == "" {
		fmt.Fprintln(os.Stderr, "  ❌ Could not detect your shell profile. Please add manually:")
		fmt.Fprintf(os.Stderr, "     %s\n", EvalLine(shellType))
		os.Exit(1)
	}

	data, _ := os.ReadFile(profilePath)
	if strings.Contains(string(data), ShellMarker) {
		fmt.Println("  ✅ commit-roaster is already set up in your shell!")
		fmt.Printf("     Profile: %s\n", profilePath)
		return
	}

	line := EvalLine(shellType)
	snippet := fmt.Sprintf("\n%s\n%s\n", ShellMarker, line)

	os.MkdirAll(filepath.Dir(profilePath), 0755)

	f, err := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ❌ Could not write to %s: %v\n", profilePath, err)
		fmt.Fprintf(os.Stderr, "     Please add this line manually: %s\n", line)
		os.Exit(1)
	}
	defer f.Close()

	if _, err := f.WriteString(snippet); err != nil {
		fmt.Fprintf(os.Stderr, "  ❌ Failed to write to %s: %v\n", profilePath, err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("  ✅ commit-roaster is now active!\n")
	fmt.Printf("     Shell:   %s\n", shellType)
	fmt.Printf("     Profile: %s\n", profilePath)
	fmt.Println()
	fmt.Println("  🔄 Restart your terminal or run:")
	if shellType == "fish" {
		fmt.Printf("     source %s\n", profilePath)
	} else if shellType == "powershell" || shellType == "pwsh" {
		fmt.Printf("     . $PROFILE\n")
	} else {
		fmt.Printf("     source %s\n", profilePath)
	}
	fmt.Println()
	fmt.Println("  Every git commit will now be roasted. You're welcome. 🔥")
	fmt.Println()
}

// CmdTeardown removes commit-roaster from the user's shell profile.
func CmdTeardown() {
	_, profilePath := DetectShellProfile()

	if profilePath == "" {
		fmt.Fprintln(os.Stderr, "  ❌ Could not detect your shell profile.")
		os.Exit(1)
	}

	data, err := os.ReadFile(profilePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  ❌ Could not read %s: %v\n", profilePath, err)
		os.Exit(1)
	}

	content := string(data)
	if !strings.Contains(content, ShellMarker) {
		fmt.Println("  ℹ️  commit-roaster is not installed in your shell profile.")
		return
	}

	lines := strings.Split(content, "\n")
	var cleaned []string
	skipNext := false
	for _, line := range lines {
		if strings.TrimSpace(line) == ShellMarker {
			skipNext = true
			continue
		}
		if skipNext {
			skipNext = false
			continue
		}
		cleaned = append(cleaned, line)
	}

	newContent := strings.Join(cleaned, "\n")
	if err := os.WriteFile(profilePath, []byte(newContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "  ❌ Could not write to %s: %v\n", profilePath, err)
		os.Exit(1)
	}

	CleanLegacyHooks()

	fmt.Println("  ✅ commit-roaster has been removed from your shell.")
	fmt.Printf("     Profile: %s\n", profilePath)
	fmt.Println("     Your commits are safe from roasting. For now. 😏")
	fmt.Println()
}
