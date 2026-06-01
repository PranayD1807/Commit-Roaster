<div align="center">
  <h1>🔥 commit-roaster</h1>
  <p><strong>A CLI tool that roasts your terrible git commit messages.</strong></p>
  <p>Rule-based by default. Optionally powered by Gemini, Claude, or ChatGPT for extra creative pain.</p>
</div>

<br />

`commit-roaster` quietly judges your commit messages every time you run `git commit` in your terminal. It uses lightweight shell integration — no git hooks, no file modifications, no conflicts with Husky or pre-commit.

The commits still go through — because breaking your workflow is evil — but your ego will take a hit.

## Features
- **One-Command Setup:** Run `commit-roaster setup` and you're done. Works on Bash, Zsh, Fish, and PowerShell.
- **Zero Conflicts:** Doesn't touch `.git/hooks`. Works alongside Husky, pre-commit, Lefthook, and any other hook manager.
- **Fast & Offline by Default:** Rule-based heuristics, no network dependencies, instant feedback.
- **Optional AI Mode:** Enable Gemini, Claude, or ChatGPT for creative, LLM-powered roasts — disabled by default.
- **Cross-Platform:** Supports macOS, Linux, and Windows out of the box.
- **Historical Analysis:** Run `stats` to view an audit of your commit history's overall hygiene.
- **12 Brutal Rules:** More than 140 randomly rotated roasts so you rarely get the same insult twice.

---

## 📦 Installation

### Method 1: Homebrew (macOS / Linux)
Because the tool is hosted on a custom tap, you can install it using:
```bash
brew install PranayD1807/tap/commit-roaster
```

### Method 2: Go Install
If you already have Go installed, this is the quickest method:
```bash
go install github.com/PranayD1807/Commit-Roaster@latest
```

### Method 3: Build from Source
```bash
git clone https://github.com/PranayD1807/Commit-Roaster.git
cd Commit-Roaster
make install
```

### Enable Roasting (One Command!)
After installing the binary, just run:
```bash
commit-roaster setup
```

That's it. It auto-detects your shell (Bash, Zsh, Fish, or PowerShell), appends the integration line to the right profile file, and you're done. [(How does this work?)](HOW_IT_WORKS.md)

To fully remove it later:
```bash
commit-roaster teardown
```

<details>
<summary><strong>Manual setup (advanced)</strong></summary>

If you prefer to add the integration yourself, use `commit-roaster init`:

**Bash & Zsh** (`~/.zshrc` or `~/.bashrc`):
```bash
eval "$(commit-roaster init)"
```

**Fish** (`~/.config/fish/config.fish`):
```fish
commit-roaster init fish | source
```

**PowerShell** (`$PROFILE`):
```powershell
Invoke-Expression (&commit-roaster init powershell | Out-String)
```
</details>


---

## 🚀 Usage

Wait for it to roast you organically in the terminal when committing:
```bash
$ git commit -m "update"

  🔥 commit-roaster says:
  Message: "update"

  💀 Too Short — 'update'? A commit message so short, it's basically a cry for help.
  💀 Too Vague — You might as well have written 'I did a thing' and called it a day.

  😬 Score: 6/10 — Mediocre at best.
```

### Manual Audit Commands

**1. Roast recent commits in the current repo:**
```bash
# Roast the last commit
commit-roaster roast

# Roast the last 5 commits
commit-roaster roast --last 5
```

**2. See your permanent record:**
```bash
commit-roaster stats
```
*(This command audits your entire git history in the current repository and gives you a breakdown of your commit message violations and an overall grade).*

---

## 🤖 AI Roaster Mode

Tired of local, hardcoded insults? Enable **AI Roaster Mode** to get creative, LLM-powered roasts from Google Gemini, Claude, or ChatGPT.

### Configuration

AI mode is disabled by default. To enable it:
```bash
commit-roaster ai enable
```
This interactive setup will guide you through:
1. Selecting your AI provider (Gemini, Claude, or ChatGPT).
2. Selecting from the latest supported models (e.g., `gemini-2.5-flash`, `claude-sonnet-4-20250514`, `gpt-4.1`).
3. Entering your API key.

To check your current AI roaster configuration and status:
```bash
commit-roaster ai status
```

To disable AI mode and revert to the local rule-based roaster:
```bash
commit-roaster ai disable
```

### 🔒 Security

* **Owner-Only Permissions**: Your API key and choices are saved locally at `~/.config/commit-roaster/config.json` with strict `0600` permissions (owner read/write).
* **Safe Output**: Your API key is masked when viewing `ai status`.
* **Zero Logging**: API keys never leave your machine except when sent securely over HTTPS to the respective provider APIs.
* **Fail Safe**: If the AI model fails (network error, invalid key), it falls back gracefully and silently to the local rule-based engine, never blocking your commits.

---

## 📖 The Roast Rules

`commit-roaster` analyzes your subject line against 12 different heuristics:

1. **Too Short (<10 chars):** Let's use our words.
2. **Too Vague:** `fix`, `update`, `stuff`... Schroedinger's commits.
3. **ALL CAPS:** Please stop yelling at the version control tree.
4. **No Verb:** Must include an action (or a conventional commit prefix).
5. **Profanity:** Keep it PG-13, HR is watching.
6. **Wall of Text:** If it's over 150 characters, use the commit body instead.
7. **Trailing Period:** This isn't formal poetry. Stop adding periods to subject lines.
8. **Ticket Only:** Example: `JIRA-1234`. Providing zero readable context.
9. **Despair / YOLO:** Mentions of `pray`, `yolo`, `hopefully`, `black magic`.
10. **File Name Only:** E.g., `src/main.js`. We know what file changed. *What* did you do to it?
11. **Exclamation Overdose:** More than one exclamation point. Calm down. 
12. **Default Message:** GitHub web edits like `Update README.md` or git's auto-generated merge subjects. 

---

## 🛠 Contributing

Want to make the roasts more savage? Adding new rules is incredibly easy!

1. Add your new `RuleID` constant in `internal/roaster/rules.go`.
2. Register it in `DefaultRules()` and write a simple `checkMyRule(msg)` bool function.
3. Add 10-15 funny strings to the `roastTemplates` map inside `internal/roaster/roasts.go`.

### Running Tests

**Unit tests** (no API key required):
```bash
go test ./...
```

**Integration tests** against the live Gemini API (requires a free key from [aistudio.google.com](https://aistudio.google.com)):
```bash
# Add GEMINI_KEY=your_key to a .env file in the project root, then:
go test -v -tags integration ./tests/ -run TestGemini -timeout 120s
```

Pull requests are actively encouraged!

---

## 🚀 Publishing New Versions
This project uses **GoReleaser** to automate building cross-platform binaries, uploading GitHub Releases, and updating the Homebrew Tap formula.

To publish a new version:

1. Commit all your changes completely.
2. Create and push a new lightweight git tag for the version:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. Run GoReleaser to automatically build and deploy:
   ```bash
   goreleaser release --clean
   ```

*Note: You must have a `GITHUB_TOKEN` exported in your terminal with `repo` scopes for GoReleaser to successfully push the release and update the tap.*

---

## 🪪 License

MIT. 

*Enjoy being roasted.*
