package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/akshaymemane/superclaw/internal/agent"
	"github.com/akshaymemane/superclaw/internal/config"
	"github.com/akshaymemane/superclaw/internal/session"
	"github.com/akshaymemane/superclaw/pkg/superclaw"
)

func main() {
	var (
		configPath = flag.String("config", config.DefaultPath, "path to superclaw.json config file")
		workDir    = flag.String("workdir", "", "constrain file operations to this directory (overrides config)")
		maxSteps   = flag.Int("max-steps", 0, "maximum LLM calls before halting (overrides config; 0 = use config/default)")
		toolsCSV   = flag.String("tools", "", "comma-separated allowlist of tools (overrides config; empty = use config/default)")
		skillsCSV  = flag.String("skills", "", "comma-separated skills to activate: summarize,research,extract,coding,github (overrides config)")
		quiet      = flag.Bool("quiet", false, "suppress progress output (tool calls not printed to stderr)")
	)
	flag.Parse()

	task := strings.Join(flag.Args(), " ")
	if task == "" {
		fmt.Fprintln(os.Stderr, "usage: superclaw [flags] <task>")
		fmt.Fprintln(os.Stderr, "\ntools:  fetch_url  web_search  read_file  write_file  run_bash  list_files  patch_file")
		fmt.Fprintln(os.Stderr, "skills: summarize  research   extract     coding    github")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Load config file (missing file is fine — returns empty File).
	cfg, err := config.LoadOrDefault(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config %s: %v\n", *configPath, err)
		os.Exit(1)
	}

	// Build agent config with file values as defaults, CLI flags as overrides.
	agentCfg := agent.DefaultConfig()
	if cfg.Model != "" {
		agentCfg.Model = agent.ModelFromString(cfg.Model)
	}
	if cfg.MaxSteps > 0 {
		agentCfg.MaxSteps = cfg.MaxSteps
	}
	if cfg.MaxTokens > 0 {
		agentCfg.MaxTokens = cfg.MaxTokens
	}
	if cfg.SystemPrompt != "" {
		agentCfg.SystemPrompt = cfg.SystemPrompt
	}
	if cfg.TimeoutSeconds > 0 {
		agentCfg.Timeout = superclaw.DurationFromSeconds(cfg.TimeoutSeconds)
	}
	// CLI flag overrides
	if *maxSteps > 0 {
		agentCfg.MaxSteps = *maxSteps
	}

	// Resolve workdir: CLI > config > "."
	wd := cfg.WorkDir
	if *workDir != "" {
		wd = *workDir
	}
	if wd == "" {
		wd = "."
	}

	// Resolve tools allowlist: CLI > config > nil (all)
	allowedTools := splitCSV(*toolsCSV)
	if len(allowedTools) == 0 {
		allowedTools = cfg.Tools
	}

	// Resolve skills: CLI > config
	activeSkills := splitCSV(*skillsCSV)
	if len(activeSkills) == 0 {
		activeSkills = cfg.Skills
	}

	// Build run options.
	opts := superclaw.Options{
		Config:        &agentCfg,
		WorkDir:       wd,
		Tools:         allowedTools,
		Skills:        activeSkills,
		MaxFetchCalls: cfg.MaxFetchCalls,
	}
	if !*quiet {
		opts.Hooks = agent.ProgressHooks(os.Stderr)
	}

	fmt.Fprintf(os.Stderr, "superclaw  task: %s\n", task)

	result := superclaw.Run(context.Background(), task, opts)

	// Persist the run record regardless of success/failure.
	store := session.NewStore(wd)
	status := "success"
	var runErr error
	if result.Err != nil {
		status = "error"
		runErr = result.Err
	}
	rec := session.NewRecord(task, result.Output, result.Steps, status, runErr)
	if appendErr := store.Append(rec); appendErr != nil {
		fmt.Fprintf(os.Stderr, "warning: could not save session record: %v\n", appendErr)
	}

	if result.Err != nil {
		fmt.Fprintf(os.Stderr, "error after %d steps: %v\n", result.Steps, result.Err)
		os.Exit(1)
	}

	fmt.Println(result.Output)
	fmt.Fprintf(os.Stderr, "done in %d steps\n", result.Steps)
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
