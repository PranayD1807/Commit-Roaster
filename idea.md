# commit-roaster 🔥

A CLI tool that **roasts bad git commit messages** using rule-based pattern matching and pre-written snarky responses. No API keys, no network calls — pure local heuristics.

**Install once via Homebrew. Apply globally. Every bad commit gets roasted.**

---

## Core Roast Rules

| Rule | Detects | Example Roast |
|---|---|---|
| **Too short** | `fix`, `wip`, `update` | *"Ah yes, 'fix'. Future you will love this treasure hunt."* |
| **Too vague** | `misc changes`, `stuff`, `things` | *"'Stuff'? Are you writing code or cleaning your garage?"* |
| **All caps** | `FIXED THE BUG` | *"We get it, you're excited. Indoor voice, please."* |
| **No verb** | `button`, `header` | *"What about the button? Did it hurt you?"* |
| **Profanity** | Detectable swear words | *"Save the rage for code review, champ."* |
| **Repeat offender** | Same message used N times | *"You've used 'fix' 12 times. Therapy might help."* |
| **Wall of text** | 200+ character messages | *"This isn't your diary. That's what PR descriptions are for."* |
| **Trailing period** | `Fixed the bug.` | *"This isn't an essay. Drop the period, Shakespeare."* |

Each rule triggers a random roast from a pool of pre-written templates — keeps it fresh.

---

## Architecture

**Language: Go** — single binary, no runtime, perfect for Homebrew distribution.

```
commit-roaster/
├── main.go              # CLI entrypoint
├── roaster/
│   ├── analyzer.go      # Rule engine — checks commit message against rules
│   ├── rules.go         # Individual rule definitions
│   └── roasts.go        # Pre-written roast template pools per rule
├── hook/
│   └── install.go       # `commit-roaster install` — sets up git hook globally
├── go.mod
├── Makefile
└── Formula/
    └── commit-roaster.rb # Homebrew formula
```

---

## Usage

```bash
# Install once
brew install commit-roaster

# Apply globally to all repos
commit-roaster install --global

# Now every commit gets roasted automatically
git commit -m "fix"
# 🔥 Roast: "Ah yes, 'fix'. A commit message so vague it could be a horoscope."

# Or roast manually / retroactively
commit-roaster roast              # roasts last commit
commit-roaster roast --last 5     # roasts last 5 commits
commit-roaster stats              # shows your commit message sin history
```

---

## Design Decisions

### Hook Type: `commit-msg` (not `post-commit`)
Fires *before* the commit finalizes — roast is printed, commit still goes through. Non-blocking by default.

### Roast Severity Levels
`--spicy mild|medium|hot` flag to control how savage the roasts are.

### Stats Mode
`commit-roaster stats` — analyze full commit history, give an overall "commit hygiene" score with a breakdown of violations. Fun shareable output.

### Zero Config by Default
Optional `.commit-roaster.yml` to disable specific rules or add custom patterns. But works out of the box with no setup.

### No Network. Ever.
Everything runs locally, instantly. No API keys, no telemetry, no external calls.

---

## Open Questions

- [ ] Should the roast block the commit or just print? (Non-blocking default, `--strict` mode to block?)
- [ ] Roast tone preference? (Sarcastic, friendly, Gordon Ramsay mode?)
- [ ] Any additional rules to add?
