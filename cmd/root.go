package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

// Version is set at build time by GoReleaser via -ldflags.
// When running locally with `go run` or `go build`, it defaults to "dev".
var Version = "dev"

// Run is the entry point called by main(). It seeds the RNG and dispatches commands.
func Run() {
	rand.Seed(time.Now().UnixNano())

	if len(os.Args) < 2 {
		PrintUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "ai":
		CmdAI()
	case "setup":
		CleanLegacyHooks()
		CmdSetup()
	case "teardown":
		CmdTeardown()
	case "init":
		CmdInit()
	case "uninstall":
		CleanLegacyHooks()
		fmt.Println("  ✅ Legacy git hooks have been successfully cleaned up.")
		fmt.Println("     Run 'commit-roaster teardown' to also remove the shell integration.")
		os.Exit(0)
	case "install":
		fmt.Fprintln(os.Stderr, "  ❌ The install command has been replaced with 'commit-roaster setup'.")
		fmt.Fprintln(os.Stderr, "     Run 'commit-roaster setup' to get started.")
		os.Exit(1)
	case "roast":
		n := GetFlagInt(1, "--last", "-n")
		CmdRoast(n)
	case "stats":
		CmdStats()
	case "hook":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: commit-roaster hook <message-file>")
			os.Exit(1)
		}
		CmdHook(os.Args[2])
	case "version", "--version", "-v":
		fmt.Printf("  commit-roaster v%s 🔥\n", Version)
	case "help", "--help", "-h":
		PrintUsage()
	default:
		fmt.Fprintf(os.Stderr, "  Unknown command: %s\n\n", os.Args[1])
		PrintUsage()
		os.Exit(1)
	}
}

func PrintUsage() {
	fmt.Println()
	fmt.Printf("  %s v%s\n\n", Bold("🔥 commit-roaster"), Version)
	fmt.Println("  Roasts your bad git commit messages. No AI, no API keys, no mercy. Or enable AI mode for extra pain.")
	fmt.Println()
	fmt.Println("  COMMANDS")
	fmt.Println("    setup                  One-command install — adds shell integration automatically")
	fmt.Println("    teardown               Remove shell integration and clean up everything")
	fmt.Println("    init [shell]           Output shell script for bash/zsh, fish, or powershell")
	fmt.Println("    roast [--last N]       Roast the last N commits (default: 1)")
	fmt.Println("    stats                  Analyze your full commit history")
	fmt.Println("    hook <file>            Roast a commit message file (for pre-commit/husky)")
	fmt.Println("    ai [status|enable|...] Configure AI Roaster mode (Gemini, Claude, ChatGPT)")
	fmt.Println("    version                Print version")
	fmt.Println("    help                   Show this message")
	fmt.Println()
	fmt.Println("  QUICK START")
	fmt.Println("    commit-roaster setup                      # one-command install")
	fmt.Println("    commit-roaster teardown                   # fully remove")
	fmt.Println()
	fmt.Println("  EXAMPLES")
	fmt.Println("    commit-roaster ai enable                  # set up AI Roaster mode")
	fmt.Println("    commit-roaster roast --last 5             # roast last 5 commits")
	fmt.Println("    commit-roaster stats                      # see your sin history")
	fmt.Println()
}
