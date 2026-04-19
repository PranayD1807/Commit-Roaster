package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/PranayD1807/Commit-Roaster/roaster"
)

const version = "1.0.0"

func main() {
	rand.Seed(time.Now().UnixNano())

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "setup":
		cleanLegacyHooks()
		cmdSetup()
	case "teardown":
		cmdTeardown()
	case "init":
		cmdInit()
	case "uninstall":
		cleanLegacyHooks()
		fmt.Println("  ✅ Legacy git hooks have been successfully cleaned up.")
		fmt.Println("     Run 'commit-roaster teardown' to also remove the shell integration.")
		os.Exit(0)
	case "install":
		fmt.Fprintln(os.Stderr, "  ❌ The install command has been replaced with 'commit-roaster setup'.")
		fmt.Fprintln(os.Stderr, "     Run 'commit-roaster setup' to get started.")
		os.Exit(1)
	case "roast":
		n := getFlagInt(1, "--last", "-n")
		cmdRoast(n)
	case "stats":
		cmdStats()
	case "hook":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: commit-roaster hook <message-file>")
			os.Exit(1)
		}
		cmdHook(os.Args[2])
	case "version", "--version", "-v":
		fmt.Printf("  commit-roaster v%s 🔥\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "  Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// cleanLegacyHooks removes any filesystem modifications made by versions < 1.0.0
func cleanLegacyHooks() {
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

// cmdInit outputs the shell integration script.
func cmdInit() {
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

const shellMarker = "# commit-roaster shell integration"

// detectShellProfile returns the shell type and profile path for the current user.
func detectShellProfile() (shellType string, profilePath string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "bash", ""
	}

	// Check $SHELL env first (works cross-platform, respects explicit overrides)
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

	// If $SHELL is not set, fall back to OS detection
	if runtime.GOOS == "windows" {
		out, err := exec.Command("powershell", "-NoProfile", "-Command", "echo $PROFILE").Output()
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			return "powershell", strings.TrimSpace(string(out))
		}
		return "powershell", filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	}

	// Final fallback: bash
	profile := filepath.Join(home, ".bashrc")
	if _, err := os.Stat(profile); os.IsNotExist(err) {
		profile = filepath.Join(home, ".bash_profile")
	}
	return "bash", profile
}

// evalLine returns the line to append to the shell profile.
func evalLine(shellType string) string {
	switch shellType {
	case "fish":
		return "commit-roaster init fish | source"
	case "powershell", "pwsh":
		return "Invoke-Expression (&commit-roaster init powershell | Out-String)"
	default:
		return `eval "$(commit-roaster init)"`
	}
}

// cmdSetup automatically appends the shell integration line to the user's profile.
func cmdSetup() {
	shellType, profilePath := detectShellProfile()

	if profilePath == "" {
		fmt.Fprintln(os.Stderr, "  ❌ Could not detect your shell profile. Please add manually:")
		fmt.Fprintf(os.Stderr, "     %s\n", evalLine(shellType))
		os.Exit(1)
	}

	// Check if already installed
	data, _ := os.ReadFile(profilePath)
	if strings.Contains(string(data), shellMarker) {
		fmt.Println("  ✅ commit-roaster is already set up in your shell!")
		fmt.Printf("     Profile: %s\n", profilePath)
		return
	}

	line := evalLine(shellType)
	snippet := fmt.Sprintf("\n%s\n%s\n", shellMarker, line)

	// Ensure parent directory exists (for fish config)
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

// cmdTeardown removes commit-roaster from the user's shell profile.
func cmdTeardown() {
	_, profilePath := detectShellProfile()

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
	if !strings.Contains(content, shellMarker) {
		fmt.Println("  ℹ️  commit-roaster is not installed in your shell profile.")
		return
	}

	// Remove the marker line and the line immediately after it
	lines := strings.Split(content, "\n")
	var cleaned []string
	skipNext := false
	for _, line := range lines {
		if strings.TrimSpace(line) == shellMarker {
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

	cleanLegacyHooks()

	fmt.Println("  ✅ commit-roaster has been removed from your shell.")
	fmt.Printf("     Profile: %s\n", profilePath)
	fmt.Println("     Your commits are safe from roasting. For now. 😏")
	fmt.Println()
}

// cmdHook is called by the git commit-msg hook. Reads the message file and roasts.
func cmdHook(msgFile string) {
	data, err := os.ReadFile(msgFile)
	if err != nil {
		return // fail silently so commits aren't blocked
	}

	msg := stripComments(string(data))
	if strings.TrimSpace(msg) == "" {
		return
	}

	result := roaster.Analyze(msg)
	if len(result.Violations) > 0 {
		roaster.PrintResult(result)
	}
}

// cmdRoast roasts the last N commits from git log.
func cmdRoast(n int) {
	if n < 1 {
		n = 1
	}
	if n > 100 {
		n = 100
	}

	out, err := exec.Command("git", "log", fmt.Sprintf("-n%d", n), "--format=%H|||%s").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "  ❌ Not a git repository or no commits found.")
		os.Exit(1)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		fmt.Println("  No commits found.")
		return
	}

	for _, line := range lines {
		parts := strings.SplitN(line, "|||", 2)
		if len(parts) != 2 {
			continue
		}
		hash := parts[0][:7]
		msg := parts[1]

		result := roaster.Analyze(msg)
		if len(result.Violations) > 0 {
			fmt.Printf("  %s %s\n", dim("commit"), dim(hash))
			roaster.PrintResult(result)
		} else {
			fmt.Printf("  %s %s — %s %s\n", dim("commit"), dim(hash), green("✓"), dim(roaster.FirstLine(msg)))
		}
	}
}

// cmdStats analyzes the full commit history.
func cmdStats() {
	out, err := exec.Command("git", "log", "--format=%s").Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "  ❌ Not a git repository or no commits found.")
		os.Exit(1)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		fmt.Println("  No commits found.")
		return
	}

	var results []roaster.Result
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			results = append(results, roaster.Analyze(line))
		}
	}

	roaster.PrintStats(results)
}

// stripComments removes lines starting with # (git's default comment char).
func stripComments(msg string) string {
	var lines []string
	for _, line := range strings.Split(msg, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func printUsage() {
	fmt.Println()
	fmt.Printf("  %s v%s\n\n", bold("🔥 commit-roaster"), version)
	fmt.Println("  Roasts your bad git commit messages. No AI, no API keys, no mercy.")
	fmt.Println()
	fmt.Println("  COMMANDS")
	fmt.Println("    setup                  One-command install — adds shell integration automatically")
	fmt.Println("    teardown               Remove shell integration and clean up everything")
	fmt.Println("    init [shell]           Output shell script for bash/zsh, fish, or powershell")
	fmt.Println("    roast [--last N]       Roast the last N commits (default: 1)")
	fmt.Println("    stats                  Analyze your full commit history")
	fmt.Println("    hook <file>            Roast a commit message file (for pre-commit/husky)")
	fmt.Println("    version                Print version")
	fmt.Println("    help                   Show this message")
	fmt.Println()
	fmt.Println("  QUICK START")
	fmt.Println("    commit-roaster setup                      # one-command install")
	fmt.Println("    commit-roaster teardown                   # fully remove")
	fmt.Println()
	fmt.Println("  EXAMPLES")
	fmt.Println("    commit-roaster roast --last 5              # roast last 5 commits")
	fmt.Println("    commit-roaster stats                      # see your sin history")
	fmt.Println()
}

// --- flag helpers ---

func hasFlag(flags ...string) bool {
	for _, arg := range os.Args[2:] {
		for _, f := range flags {
			if arg == f {
				return true
			}
		}
	}
	return false
}

func getFlagInt(def int, flags ...string) int {
	args := os.Args[2:]
	for i, arg := range args {
		for _, f := range flags {
			if arg == f && i+1 < len(args) {
				if v, err := strconv.Atoi(args[i+1]); err == nil {
					return v
				}
			}
		}
	}
	return def
}

// ANSI helpers (duplicated from roaster for main package use)
func bold(s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return "\033[1m" + s + "\033[0m"
}

func dim(s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return "\033[2m" + s + "\033[0m"
}

func green(s string) string {
	if os.Getenv("NO_COLOR") != "" {
		return s
	}
	return "\033[32m" + s + "\033[0m"
}
