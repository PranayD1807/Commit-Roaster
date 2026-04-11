package roaster

import (
	"fmt"
	"os"
	"strings"
)

// Violation represents a single rule failure.
type Violation struct {
	Rule  Rule
	Roast string
}

// Result holds the analysis of a single commit message.
type Result struct {
	Message    string
	Violations []Violation
	Score      int // 0-10
}

// Analyze checks a commit message against all default rules.
func Analyze(msg string) Result {
	rules := DefaultRules()
	var violations []Violation

	for _, r := range rules {
		if r.Check(msg) {
			violations = append(violations, Violation{
				Rule:  r,
				Roast: GetRoast(r.ID, msg),
			})
		}
	}

	score := 10 - (len(violations) * 2)
	if score < 0 {
		score = 0
	}

	return Result{
		Message:    msg,
		Violations: violations,
		Score:      score,
	}
}

// ANSI color helpers
var useColor = true

func init() {
	if os.Getenv("NO_COLOR") != "" {
		useColor = false
	}
}

func c(code, text string) string {
	if !useColor {
		return text
	}
	return code + text + "\033[0m"
}

func bold(s string) string   { return c("\033[1m", s) }
func dim(s string) string    { return c("\033[2m", s) }
func red(s string) string    { return c("\033[31m", s) }
func green(s string) string  { return c("\033[32m", s) }
func yellow(s string) string { return c("\033[33m", s) }
func cyan(s string) string   { return c("\033[36m", s) }

// PrintResult prints a formatted roast result to stdout.
func PrintResult(r Result) {
	if len(r.Violations) == 0 {
		fmt.Printf("\n  %s  %s\n\n", green("✓"), dim("Clean commit. No roast today."))
		return
	}

	subject := FirstLine(r.Message)
	if len(subject) > 60 {
		subject = subject[:57] + "..."
	}

	fmt.Printf("\n  %s %s\n", bold("🔥 commit-roaster"), dim("says:"))
	fmt.Printf("  %s %s\n\n", dim("Message:"), cyan("\""+subject+"\""))

	for _, v := range r.Violations {
		fmt.Printf("  %s %s — %s\n", red("💀"), bold(v.Rule.Name), v.Roast)
	}

	fmt.Printf("\n  %s %s\n\n", scoreEmoji(r.Score), scoreText(r.Score))
}

// PrintStats prints aggregate stats for multiple results.
func PrintStats(results []Result) {
	if len(results) == 0 {
		fmt.Println("  No commits found.")
		return
	}

	totalScore := 0
	ruleCounts := make(map[RuleID]int)
	cleanCount := 0

	for _, r := range results {
		totalScore += r.Score
		if len(r.Violations) == 0 {
			cleanCount++
		}
		for _, v := range r.Violations {
			ruleCounts[v.Rule.ID]++
		}
	}

	avg := float64(totalScore) / float64(len(results))

	fmt.Printf("\n  %s\n", bold("🔥 commit-roaster stats"))
	fmt.Printf("  %s\n\n", dim(strings.Repeat("━", 36)))

	fmt.Printf("  📊 Commits analyzed:  %s\n", bold(fmt.Sprintf("%d", len(results))))
	fmt.Printf("  ✅ Clean commits:     %s %s\n", bold(fmt.Sprintf("%d", cleanCount)), dim(fmt.Sprintf("(%.0f%%)", float64(cleanCount)/float64(len(results))*100)))
	fmt.Printf("  🏆 Average score:     %s\n\n", bold(fmt.Sprintf("%.1f/10", avg)))

	if len(ruleCounts) > 0 {
		fmt.Printf("  %s\n", bold("Top Violations:"))
		rules := DefaultRules()
		for _, rule := range rules {
			count, ok := ruleCounts[rule.ID]
			if !ok {
				continue
			}
			pct := float64(count) / float64(len(results)) * 100
			bar := strings.Repeat("█", int(pct/5))
			fmt.Printf("  %s %-16s %s %s %s\n",
				red("💀"), rule.Name,
				yellow(fmt.Sprintf("%3d", count)),
				dim(fmt.Sprintf("(%2.0f%%)", pct)),
				dim(bar),
			)
		}
	}

	fmt.Printf("\n  %s %s\n\n", scoreEmoji(int(avg)), overallText(avg))
}

func scoreEmoji(score int) string {
	switch {
	case score >= 9:
		return "🏆"
	case score >= 7:
		return "👍"
	case score >= 5:
		return "😬"
	case score >= 3:
		return "💩"
	default:
		return "☠️"
	}
}

func scoreText(score int) string {
	s := fmt.Sprintf("Score: %d/10", score)
	switch {
	case score >= 9:
		return green(s) + dim(" — Surprisingly decent.")
	case score >= 7:
		return green(s) + dim(" — Not bad, could be better.")
	case score >= 5:
		return yellow(s) + dim(" — Mediocre at best.")
	case score >= 3:
		return red(s) + dim(" — Your git log weeps.")
	default:
		return red(s) + dim(" — Absolute carnage.")
	}
}

func overallText(avg float64) string {
	switch {
	case avg >= 9:
		return green(fmt.Sprintf("%.1f/10", avg)) + dim(" — Your commit messages are chef's kiss.")
	case avg >= 7:
		return green(fmt.Sprintf("%.1f/10", avg)) + dim(" — Pretty solid commit hygiene.")
	case avg >= 5:
		return yellow(fmt.Sprintf("%.1f/10", avg)) + dim(" — Room for improvement. Lots of it.")
	case avg >= 3:
		return red(fmt.Sprintf("%.1f/10", avg)) + dim(" — Your git log is a crime scene.")
	default:
		return red(fmt.Sprintf("%.1f/10", avg)) + dim(" — Have you considered a career change?")
	}
}
