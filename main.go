package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
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
	case "init":
		cleanLegacyHooks()
		cmdInit()
	case "uninstall":
		cleanLegacyHooks()
		fmt.Println("  ✅ Legacy git hooks have been successfully cleaned up.")
		os.Exit(0)
	case "install":
		fmt.Fprintln(os.Stderr, "  ❌ The install command has been removed in favor of non-destructive shell integration.")
		fmt.Fprintln(os.Stderr, "     Run 'commit-roaster init' to learn how to add it to your terminal securely.")
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
	script := `
# commit-roaster shell integration
# Add this to your ~/.zshrc or ~/.bashrc:
# eval "$(commit-roaster init)"

git() {
  command git "$@"
  local ext_code=$?
  if [ "$1" = "commit" ] && [ $ext_code -eq 0 ]; then
    commit-roaster roast --last 1
  fi
  return $ext_code
}
`
	fmt.Println(strings.TrimSpace(script))
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
	fmt.Println("    init                   Output shell integration script (add to .zshrc/.bashrc)")
	fmt.Println("    uninstall              Clean up any legacy git hooks injected by older versions")
	fmt.Println("    roast [--last N]       Roast the last N commits (default: 1)")
	fmt.Println("    stats                  Analyze your full commit history")
	fmt.Println("    hook <file>            Roast a commit message file (for pre-commit/husky)")
	fmt.Println("    version                Print version")
	fmt.Println("    help                   Show this message")
	fmt.Println()
	fmt.Println(bold("  EXAMPLES"))
	fmt.Println("    eval \"$(commit-roaster init)\"      # set up shell wrapper automatically")
	fmt.Println("    commit-roaster roast --last 5      # roast last 5 commits")
	fmt.Println("    commit-roaster stats               # see your sin history")
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
