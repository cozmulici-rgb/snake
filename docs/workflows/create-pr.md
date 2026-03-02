# Create a Pull Request (PR)

This is the project memory for creating PRs reliably.

## Preconditions
- Your branch is pushed to `origin`.
- Your working tree is clean: `git status --short` should be empty.
- You know the base branch (usually `main`).

## Quick CLI Flow (GitHub CLI)
1. Authenticate once per machine/session if needed:
```powershell
gh auth login
```
2. Verify branch and push:
```powershell
git branch --show-current
git push -u origin <your-branch>
```
3. Create PR:
```powershell
gh pr create --base main --head <your-branch> --title "<PR title>" --body "<PR summary>"
```
4. Open/inspect PR:
```powershell
gh pr view --web
```

## Browser Flow (No `gh` auth required)
1. Push your branch:
```powershell
git push -u origin <your-branch>
```
2. Open:
`https://github.com/cozmulici-rgb/snake/compare/main...<your-branch>?expand=1`
3. Fill title/body and submit PR.

## Useful Checks
```powershell
git status --short
git log --oneline -5
gh pr list --head <your-branch> --state open
```

## Notes
- If `gh pr create` fails with auth error, run `gh auth login`.
- Avoid including unrelated local changes in the PR.
