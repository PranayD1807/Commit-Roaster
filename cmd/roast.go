package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/PranayD1807/Commit-Roaster/internal/airoaster"
	"github.com/PranayD1807/Commit-Roaster/internal/config"
	"github.com/PranayD1807/Commit-Roaster/internal/roaster"
)

// CmdRoast roasts the last N commits from git log.
func CmdRoast(n int) {
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

	cfg, cfgErr := config.Load()

	for _, line := range lines {
		parts := strings.SplitN(line, "|||", 2)
		if len(parts) != 2 {
			continue
		}
		hash := parts[0][:7]
		msg := parts[1]

		useAI := cfgErr == nil && cfg.AIEnabled

		if useAI {
			roastMsg, aiErr := airoaster.RoastCommitFn(cfg.Provider, cfg.Model, cfg.APIKey, msg)
			if aiErr == nil {
				if strings.ToUpper(strings.TrimSpace(roastMsg)) == "CLEAN" {
					fmt.Printf("  %s %s — %s %s\n", Dim("commit"), Dim(hash), Green("✓"), Dim(roaster.FirstLine(msg)))
				} else {
					fmt.Printf("  %s %s\n", Dim("commit"), Dim(hash))
					fmt.Printf("  🤖 %s (%s):\n", Bold("AI Roast"), Dim(cfg.Model))
					fmt.Printf("     %s\n\n", roastMsg)
				}
				continue
			}
			fmt.Fprintf(os.Stderr, "  ⚠️  AI Roast failed: %v. Falling back to rule-based roast.\n", aiErr)
		}

		result := roaster.Analyze(msg)
		if len(result.Violations) > 0 {
			fmt.Printf("  %s %s\n", Dim("commit"), Dim(hash))
			roaster.PrintResult(result)
		} else {
			fmt.Printf("  %s %s — %s %s\n", Dim("commit"), Dim(hash), Green("✓"), Dim(roaster.FirstLine(msg)))
		}
	}
}
