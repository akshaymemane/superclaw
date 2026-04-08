package builtin

// CodingSkill teaches the agent safe, minimal code editing practices.
type CodingSkill struct{}

func NewCodingSkill() *CodingSkill { return &CodingSkill{} }

func (c *CodingSkill) Name() string { return "coding" }

func (c *CodingSkill) Description() string {
	return "Safe, minimal code editing: read before edit, patch over overwrite, test after change."
}

func (c *CodingSkill) Instructions() string {
	return `When working with code:
1. Always read a file with read_file before editing it. Do not edit from memory.
2. Prefer patch_file over write_file. patch_file changes only what you intend; write_file overwrites the whole file.
3. After editing, verify the result with read_file to confirm the change looks correct.
4. Make the smallest change that satisfies the task. Do not refactor, reformat, or clean up code you were not asked to change.
5. If the project has a test command (e.g. "go test ./..." or "npm test"), run it with run_bash after making changes.
6. When exploring an unfamiliar codebase, use list_files first to understand the structure, then read specific files.
7. Never delete files unless explicitly instructed.`
}
