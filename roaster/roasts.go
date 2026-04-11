package roaster

import "math/rand"

// roastTemplates maps each rule to a pool of snarky messages.
var roastTemplates = map[RuleID][]string{
	RuleTooShort: {
		"Ah yes, '%s'. Future you will love this treasure hunt.",
		"'%s'? A commit message so short, it's basically a cry for help.",
		"Even a tweet has more character than this. Literally.",
		"Bold move, writing a commit message shorter than your attention span.",
		"'%s' — the literary equivalent of a shrug emoji.",
		"One day you'll read this in git blame and feel nothing. Because there's nothing to feel.",
	},
	RuleTooVague: {
		"'%s' could mean literally anything. Schrödinger's commit.",
		"If your commit message was a GPS, it would just say 'go somewhere'.",
		"This commit message has less context than a politician's promise.",
		"'%s' — future archaeologists will be baffled by this artifact.",
		"You might as well have written 'I did a thing' and called it a day.",
		"Your commit message is so vague it could be a horoscope.",
	},
	RuleAllCaps: {
		"We get it, you're excited. Indoor voice, please.",
		"Calm down. The keyboard didn't do anything to you.",
		"CAPS LOCK: cruise control for cool? Not in commit messages.",
		"Your commit message just yelled at the entire git log.",
		"Take a deep breath. Count to ten. Then rewrite this in lowercase.",
	},
	RuleNoVerb: {
		"What about it? Did it hurt you? What did you DO?",
		"That's a noun, not a commit message. Verbs exist for a reason.",
		"This reads like a word cloud, not a description of changes.",
		"Your commit message is a statement of existence, not an action.",
		"Git commits describe what you DID, not what you're THINKING about.",
	},
	RuleProfanity: {
		"Save the rage for code review, champ.",
		"Your git log is not your therapist.",
		"Commit messages are permanent. HR can see these.",
		"Future you reading this in git blame will be so proud.",
		"Git stores this forever. Just like the internet. Choose wisely.",
	},
	RuleWallOfText: {
		"This isn't your diary. That's what PR descriptions are for.",
		"Your commit message has more words than most README files.",
		"TL;DR: You need a TL;DR for your commit message.",
		"A commit message, not a novel. Save the prose for your memoir.",
		"Subject lines should be punchy, not a paragraph. There's a body for that.",
	},
	RuleTrailingPeriod: {
		"This isn't an essay. Drop the period, Shakespeare.",
		"A period at the end of a subject line? How... formal.",
		"The period adds nothing but passive-aggressive energy.",
		"Ending with a period is the git equivalent of 'Regards.'",
		"Conventional commits don't end with periods. Now you know.",
	},
}

// GetRoast returns a random roast for the given rule, with the message substituted.
func GetRoast(id RuleID, msg string) string {
	templates, ok := roastTemplates[id]
	if !ok || len(templates) == 0 {
		return "Your commit message needs work."
	}
	tmpl := templates[rand.Intn(len(templates))]
	// Substitute %s with the commit subject (truncated).
	subject := FirstLine(msg)
	if len(subject) > 40 {
		subject = subject[:37] + "..."
	}
	// Simple manual substitution to avoid importing fmt just for this.
	result := ""
	for i := 0; i < len(tmpl); i++ {
		if i+1 < len(tmpl) && tmpl[i] == '%' && tmpl[i+1] == 's' {
			result += subject
			i++ // skip 's'
		} else {
			result += string(tmpl[i])
		}
	}
	return result
}
