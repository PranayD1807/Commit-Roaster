package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/PranayD1807/Commit-Roaster/internal/airoaster"
	"github.com/PranayD1807/Commit-Roaster/internal/config"
	"github.com/PranayD1807/Commit-Roaster/internal/roaster"
)

// CmdStats analyzes the full commit history.
func CmdStats() {
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

	cfg, cfgErr := config.Load()
	useAI := cfgErr == nil && cfg.AIEnabled

	if useAI {
		fmt.Println()
		fmt.Printf("  ⚠️  %s: Analyzing %d commits will use your AI tokens (%s / %s).\n",
			Bold("AI Roaster is enabled"), len(lines), config.ProviderDisplayName(cfg.Provider), cfg.Model)
		fmt.Print("     Proceed? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))

		if input != "y" && input != "yes" {
			fmt.Println("  ❌ Stats analysis aborted.")
			return
		}

		fmt.Println("  🤖 Analyzing commits with AI...")
	}

	var results []roaster.Result
	var lastErrKind airoaster.ErrorKind = -1 // track to avoid repeating the same warning
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		if useAI {
			roastMsg, err := airoaster.RoastCommitFn(cfg.Provider, cfg.Model, cfg.APIKey, line)
			if err == nil {
				if strings.ToUpper(strings.TrimSpace(roastMsg)) == "CLEAN" {
					results = append(results, roaster.Result{
						Message:    line,
						Violations: nil,
						Score:      10,
					})
				} else {
					results = append(results, roaster.Result{
						Message: line,
						Violations: []roaster.Violation{
							{
								Rule:  roaster.Rule{ID: "AI_ROAST", Name: "AI Roast", Description: "AI-generated roast"},
								Roast: roastMsg,
							},
						},
						Score: 5,
					})
				}
				continue
			}
			kind, aiMsg := airoaster.ClassifyError(err)
			if kind != lastErrKind {
				fmt.Fprintf(os.Stderr, "  ⚠️  AI Roast failed (%s). Falling back to rule-based roaster.\n", aiMsg)
				lastErrKind = kind
			}
		}

		results = append(results, roaster.Analyze(line))
	}

	roaster.PrintStats(results)
}
