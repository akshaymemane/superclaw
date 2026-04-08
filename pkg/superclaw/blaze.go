// Package superclaw provides the public API for running a superclaw agent.
// Use this package from external programs; internal packages are not stable.
package superclaw

import (
	"context"
	"time"

	"github.com/akshaymemane/superclaw/internal/agent"
	"github.com/akshaymemane/superclaw/internal/skills"
	skillsbuiltin "github.com/akshaymemane/superclaw/internal/skills/builtin"
	"github.com/akshaymemane/superclaw/internal/tools"
	"github.com/akshaymemane/superclaw/internal/tools/builtin"
)

// RunResult holds the outcome of a single agent run.
type RunResult struct {
	Steps  int
	Output string
	Err    error
}

// Options configures a Run call.
type Options struct {
	// APIKey is the Anthropic API key.
	// Falls back to the ANTHROPIC_API_KEY environment variable when empty.
	APIKey string
	// Config overrides the default agent config. Fields left at zero value use
	// agent.DefaultConfig() values.
	Config *agent.Config
	// Tools is the allowlist of tool names for this run.
	// Defaults to all built-in tools when nil.
	Tools []string
	// Skills is the list of built-in skill names to activate.
	// Available: "summarize", "research", "extract", "coding", "github".
	Skills []string
	// WorkDir constrains read_file, write_file, list_files, and patch_file to
	// this directory. Defaults to the current directory when empty.
	WorkDir string
	// Hooks receives lifecycle events. All fields are optional.
	Hooks agent.Hooks
	// MaxFetchCalls caps the total fetch_url calls per run. 0 = unlimited.
	MaxFetchCalls int
}

var defaultTools = []string{
	"fetch_url", "web_search", "read_file", "write_file",
	"run_bash", "list_files", "patch_file",
}

// Run executes a task with the configured tool and skill set.
func Run(ctx context.Context, task string, opts Options) RunResult {
	cfg := agent.DefaultConfig()
	if opts.Config != nil {
		cfg = *opts.Config
	}
	cfg.Hooks = opts.Hooks

	var fetchTool *builtin.FetchTool
	if opts.MaxFetchCalls > 0 {
		fetchTool = builtin.NewFetchToolWithLimit(opts.MaxFetchCalls)
	} else {
		fetchTool = builtin.NewFetchTool()
	}

	tr := tools.NewRegistry()
	tr.Register(fetchTool)
	tr.Register(builtin.NewWebSearchTool())
	tr.Register(builtin.NewReadFileTool(opts.WorkDir))
	tr.Register(builtin.NewWriteFileTool(opts.WorkDir))
	tr.Register(builtin.NewBashTool())
	tr.Register(builtin.NewListFilesTool(opts.WorkDir))
	tr.Register(builtin.NewPatchFileTool(opts.WorkDir))

	allowed := opts.Tools
	if len(allowed) == 0 {
		allowed = defaultTools
	}
	tr.Allow(allowed...)

	sr := skills.NewRegistry()
	sr.Register(skillsbuiltin.NewSummarizeSkill())
	sr.Register(skillsbuiltin.NewResearchSkill())
	sr.Register(skillsbuiltin.NewExtractSkill())
	sr.Register(skillsbuiltin.NewCodingSkill())
	sr.Register(skillsbuiltin.NewGitHubSkill())

	if len(opts.Skills) > 0 {
		if err := sr.Activate(opts.Skills...); err != nil {
			return RunResult{Err: err}
		}
	}

	a := agent.New(cfg, tr, sr, opts.APIKey)
	state, err := a.Run(ctx, task)
	return RunResult{
		Steps:  state.Steps,
		Output: state.Result,
		Err:    err,
	}
}

// DurationFromSeconds converts an int seconds value to a time.Duration.
// Returns 0 (no timeout) when seconds is 0.
func DurationFromSeconds(seconds int) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}
