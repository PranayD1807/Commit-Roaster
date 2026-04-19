<div align="center">
  <h1>🔥 commit-roaster</h1>
  <p><strong>A CLI tool that roasts your terrible git commit messages.</strong></p>
  <p>No AI. No API keys. Just pure, hardcoded malice.</p>
</div>

<br />

`commit-roaster` acts as a Git hook that quietly judges your development choices. Whenever you write a commit message that is too vague, too short, overly ecstatic, or driven by despair, `commit-roaster` will jump in and deliver a scathing critique directly to your terminal.

The commits still go through—because breaking your workflow is evil—but your ego will take a hit. 

## Features
- **Shell Integration:** Automatically roasts you right after you run `git commit`. Zero conflicts with pre-commit hooks or Husky.
- **Fast & Local:** No AI, no network dependencies, completely offline heuristical matching entirely in Go.
- **Historical Analysis:** Run `stats` to view an audit of your commit history's overall lifespan hygiene.
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

### Enable Shell Integration (No Hooks Needed!)
Unlike fragile hook-based tools that break when combined with Husky or pre-commit, `commit-roaster` intercepts `git commit` safely via your shell, avoiding `.git/hooks` entirely. [(Read how this works)](HOW_IT_WORKS.md)

To set it up safely in your shell, add this single line to your `~/.zshrc` or `~/.bashrc`:

```bash
eval "$(commit-roaster init)"
```

Restart your terminal, and you're done!

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

1. Add your new `RuleID` constant in `roaster/rules.go`.
2. Register it in `DefaultRules()` and write a simple `checkMyRule(msg)` bool function.
3. Add 10-15 funny strings to the `roastTemplates` map inside `roaster/roasts.go`. 

Pull requests are actively encouraged!

---

## 🚀 Publishing New Versions
This project uses **GoReleaser** to automate building cross-platform binaries, uploading GitHub Releases, and updating the Homebrew Tap formula.

To publish a new version:

1. Commit all your changes completely.
2. Create and push a new lightweight git tag for the version:
   ```bash
   git tag -a v0.2.0 -m "Release v0.2.0"
   git push origin v0.2.0
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
