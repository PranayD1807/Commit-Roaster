<div align="center">
  <h1>🔥 commit-roaster</h1>
  <p><strong>A CLI tool that roasts your terrible git commit messages.</strong></p>
  <p>No AI. No API keys. Just pure, hardcoded malice.</p>
</div>

<br />

`commit-roaster` acts as a Git hook that quietly judges your development choices. Whenever you write a commit message that is too vague, too short, overly ecstatic, or driven by despair, `commit-roaster` will jump in and deliver a scathing critique directly to your terminal.

The commits still go through—because breaking your workflow is evil—but your ego will take a hit. 

## Features
- **Git Hook Integration:** Automatically roasts you right as you press Enter.
- **Fast & Local:** No AI, no network dependencies, completely offline heuristical matching entirely in Go.
- **Historical Analysis:** Run `stats` to view an audit of your commit history's overall lifespan hygiene.
- **12 Brutal Rules:** More than 140 randomly rotated roasts so you rarely get the same insult twice.

---

## 📦 Installation

Since it's written in Go, you can build from source, or install globally via `make`:

```bash
git clone https://github.com/mellow/commit-roaster.git
cd commit-roaster
make install
```

### Enable the Git Hook
To actually get roasted automatically, install the hook:

```bash
# Install globally (recommended - affects all repos)
commit-roaster install --global

# Or install only in the current repository
commit-roaster install
```

*(To remove it at any time, run `commit-roaster uninstall [--global]`)*

---

## 🚀 Usage

Wait for it to roast you organically when committing:
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

## 🪪 License

MIT. 

*Enjoy being roasted.*
