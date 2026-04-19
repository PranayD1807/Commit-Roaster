# How `commit-roaster` Works (And Why We Don't Use Git Hooks)

`commit-roaster` is designed to be completely automatic while remaining **100% non-destructive**. To achieve this, it uses a Shell Integration approach rather than traditional Git Hooks.

## The Problem with Git Hooks
Initially, `commit-roaster` attempted to install itself as a native `commit-msg` Git hook (inside `.git/hooks/commit-msg`). 

However, modern version control workflows frequently rely on tooling to manage hooks automatically:
- **Husky** (JavaScript / Node.js ecosystems)
- **pre-commit** (Python / language-agnostic frameworks)
- **Lefthook** (Go / Ruby ecosystems)

Git rigidly allows exactly **one executable** per hook phase (`commit-msg`). Because these tools actively generate, monitor, and overwrite the `.git/hooks` directory to manage their own logic, any attempts by `commit-roaster` to inject or backup files would eventually be wiped out, or worse, conflict and break the developer's repository setup.

## The Solution: Shell Interception
Rather than fighting established ecosystem managers for dominance over `.git/hooks`, `commit-roaster` sidesteps Git internal hooks entirely by executing at the **Shell level**.

By evaluating the `commit-roaster init` command in your terminal profile (`~/.zshrc` or `~/.bashrc`), we define a lightweight wrapper function around the `git` command itself.

### The Shell Function Explained

When you run `eval "$(commit-roaster init)"`, it loads the following invisible wrapper into your current shell session:

```bash
git() {
  # 1. Pass all arguments directly to the REAL git executable natively
  command git "$@"
  
  # 2. Capture the exit code of that real git command
  local ext_code=$?
  
  # 3. If it was specifically a "commit" command AND it succeeded (0)...
  if [ "$1" = "commit" ] && [ $ext_code -eq 0 ]; then
  
    # 4. Jump in and roast the message!
    commit-roaster roast --last 1
  fi
  
  # 5. Return the exact exit code so you never notice the interception
  return $ext_code
}
```

### Why This Is Better
1. **Bulletproof Compatibility:** Because the wrapper delegates immediately to `command git`, all your native Husky scripts, `commitlint` passes, and `pre-commit` formatters execute 100% naturally. We don't interfere with them.
2. **Post-Commit Roasting:** By waiting for `$ext_code -eq 0`, `commit-roaster` ensures the actual commit is completely saved and accepted by your CI standards before evaluating it. It simply reads the `HEAD` commit and roasts you.
3. **Global by Default:** Since it binds to your terminal profile, the roaster works locally on every single repository you clone automatically—no setup required per repo.
4. **No Destructive Edits**: It doesn't modify a single file inside your repository. No `core.hooksPath` hijacking, no `.git/hooks` overwriting. You are safe.
