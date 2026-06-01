package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/PranayD1807/Commit-Roaster/internal/airoaster"
	"github.com/PranayD1807/Commit-Roaster/internal/config"
	"github.com/PranayD1807/Commit-Roaster/internal/roaster"
)

// CmdHook is called by the git commit-msg hook. Reads the message file and roasts.
func CmdHook(msgFile string) {
	data, err := os.ReadFile(msgFile)
	if err != nil {
		return // fail silently so commits aren't blocked
	}

	msg := StripComments(string(data))
	if strings.TrimSpace(msg) == "" {
		return
	}

	cfg, err := config.Load()
	if err == nil && cfg.AIEnabled {
		roastMsg, err := airoaster.RoastCommitFn(cfg.Provider, cfg.Model, cfg.APIKey, msg)
		if err == nil {
			if strings.ToUpper(strings.TrimSpace(roastMsg)) != "CLEAN" {
				fmt.Println()
				fmt.Printf("  🤖 %s (%s):\n", Bold("AI Roast"), Dim(cfg.Model))
				fmt.Printf("     %s\n", roastMsg)
				fmt.Println()
			}
			return
		}
		// Only surface permanent errors (bad key) — transient errors fall back silently
		kind, aiMsg := airoaster.ClassifyError(err)
		if kind == airoaster.ErrInvalidKey {
			fmt.Fprintf(os.Stderr, "  ⚠️  AI Roast skipped (%s)\n", aiMsg)
		}
		// Fall back to rule-based
	}

	result := roaster.Analyze(msg)
	if len(result.Violations) > 0 {
		roaster.PrintResult(result)
	}
}

// StripComments removes lines starting with # (git's default comment char).
func StripComments(msg string) string {
	var lines []string
	for _, line := range strings.Split(msg, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}
