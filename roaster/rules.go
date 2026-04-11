package roaster

import (
	"strings"
	"unicode"
)

// RuleID identifies a specific roast rule.
type RuleID string

const (
	RuleTooShort       RuleID = "too_short"
	RuleTooVague       RuleID = "too_vague"
	RuleAllCaps        RuleID = "all_caps"
	RuleNoVerb         RuleID = "no_verb"
	RuleProfanity      RuleID = "profanity"
	RuleWallOfText     RuleID = "wall_of_text"
	RuleTrailingPeriod RuleID = "trailing_period"
)

// Rule represents a single commit message check.
type Rule struct {
	ID          RuleID
	Name        string
	Description string
	Check       func(msg string) bool
}

// DefaultRules returns all built-in roast rules.
func DefaultRules() []Rule {
	return []Rule{
		{ID: RuleTooShort, Name: "Too Short", Description: "Commit message is suspiciously brief", Check: checkTooShort},
		{ID: RuleTooVague, Name: "Too Vague", Description: "Commit message is uselessly generic", Check: checkTooVague},
		{ID: RuleAllCaps, Name: "ALL CAPS", Description: "Why are you yelling?", Check: checkAllCaps},
		{ID: RuleNoVerb, Name: "No Verb", Description: "Doesn't start with an action", Check: checkNoVerb},
		{ID: RuleProfanity, Name: "Profanity", Description: "Contains some choice words", Check: checkProfanity},
		{ID: RuleWallOfText, Name: "Wall of Text", Description: "Subject line is way too long", Check: checkWallOfText},
		{ID: RuleTrailingPeriod, Name: "Trailing Period", Description: "Subject line ends with a period", Check: checkTrailingPeriod},
	}
}

func checkTooShort(msg string) bool {
	return len(strings.TrimSpace(FirstLine(msg))) < 10
}

func checkTooVague(msg string) bool {
	subject := strings.ToLower(strings.TrimSpace(FirstLine(msg)))
	stripped := strings.ToLower(strings.TrimSpace(StripConventionalPrefix(FirstLine(msg))))
	vagueExact := []string{
		"fix", "fixed", "fixes", "fixing",
		"update", "updated", "updates", "updating",
		"change", "changed", "changes",
		"wip", "work in progress",
		"stuff", "things", "misc",
		"minor", "minor changes", "minor fix",
		"tweaks", "cleanup", "clean up",
		"temp", "tmp", "asdf", "asd",
		"test", "testing", "commit", "save", "done",
		".", "-", "...", "---",
		"initial commit", "no message", "empty",
	}
	for _, p := range vagueExact {
		if subject == p || stripped == p {
			return true
		}
	}
	return false
}

func checkAllCaps(msg string) bool {
	subject := strings.TrimSpace(FirstLine(msg))
	if len(subject) < 6 {
		return false
	}
	upper, letters := 0, 0
	for _, r := range subject {
		if unicode.IsLetter(r) {
			letters++
			if unicode.IsUpper(r) {
				upper++
			}
		}
	}
	if letters < 4 {
		return false
	}
	return float64(upper)/float64(letters) > 0.7
}

func checkNoVerb(msg string) bool {
	subject := strings.TrimSpace(FirstLine(msg))
	if len(subject) == 0 {
		return false
	}
	stripped := StripConventionalPrefix(subject)
	if stripped != subject {
		return false // had a valid conventional prefix
	}
	words := strings.Fields(strings.ToLower(subject))
	if len(words) == 0 {
		return false
	}
	nonVerbs := map[string]bool{
		"the": true, "a": true, "an": true, "some": true,
		"this": true, "that": true, "these": true, "those": true,
		"my": true, "our": true, "their": true, "its": true,
		"lol": true, "oops": true, "hmm": true, "idk": true,
	}
	return nonVerbs[words[0]]
}

func checkProfanity(msg string) bool {
	words := strings.Fields(strings.ToLower(FirstLine(msg)))
	profane := map[string]bool{
		"fuck": true, "fucking": true, "fucked": true, "fucker": true,
		"shit": true, "shitty": true, "bullshit": true,
		"damn": true, "dammit": true, "goddamn": true,
		"ass": true, "asshole": true,
		"crap": true, "crappy": true,
		"bitch": true, "wtf": true, "stfu": true,
	}
	for _, w := range words {
		cleaned := strings.Trim(w, ".,!?;:'\"()[]{}#@")
		if profane[cleaned] {
			return true
		}
	}
	return false
}

func checkWallOfText(msg string) bool {
	return len(strings.TrimSpace(FirstLine(msg))) > 150
}

func checkTrailingPeriod(msg string) bool {
	subject := strings.TrimSpace(FirstLine(msg))
	return len(subject) > 1 && strings.HasSuffix(subject, ".")
}

// FirstLine returns the first line of a message.
func FirstLine(msg string) string {
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		return msg[:i]
	}
	return msg
}

// StripConventionalPrefix removes "feat:", "fix(scope):" etc.
func StripConventionalPrefix(subject string) string {
	colonIdx := strings.IndexByte(subject, ':')
	if colonIdx <= 0 || colonIdx > 25 {
		return subject
	}
	prefix := strings.TrimRight(subject[:colonIdx], "!")
	if p := strings.IndexByte(prefix, '('); p >= 0 {
		prefix = prefix[:p]
	}
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	types := map[string]bool{
		"feat": true, "fix": true, "docs": true, "style": true,
		"refactor": true, "perf": true, "test": true, "build": true,
		"ci": true, "chore": true, "revert": true,
	}
	if types[prefix] {
		return strings.TrimSpace(subject[colonIdx+1:])
	}
	return subject
}
