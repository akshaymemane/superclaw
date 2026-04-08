package builtin

// GitHubSkill teaches the agent to use the gh CLI safely.
// It is only useful when gh is installed and run_bash is allowed.
type GitHubSkill struct{}

func NewGitHubSkill() *GitHubSkill { return &GitHubSkill{} }

func (g *GitHubSkill) Name() string { return "github" }

func (g *GitHubSkill) Description() string {
	return "GitHub operations via the gh CLI (requires gh installed and run_bash allowed)."
}

func (g *GitHubSkill) Instructions() string {
	return `When working with GitHub via the gh CLI:
1. Verify gh is available before using it: run_bash "gh --version". If missing, tell the user and stop.
2. Use read-only commands first to understand the current state before taking action:
   - gh pr list, gh pr view <number>
   - gh issue list, gh issue view <number>
   - gh repo view
3. For creating or updating PRs and issues, confirm the content with the user before submitting.
4. Never run git push --force or git reset --hard without explicit user instruction.
5. Always run git diff or git status before committing to verify what will be included.
6. Keep commit messages concise and descriptive. Do not add boilerplate.
7. If a gh command fails with an auth error, tell the user to run "gh auth login" and stop.`
}
