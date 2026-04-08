# ARCHITECTURE.md

## Goal

`superclaw` should borrow the right ideas from `openclaw` without inheriting its full platform scope.

The target is:

- a local-first agent runtime
- a small fixed set of tools
- a small explicit set of skills
- container-first deployment with safe defaults
- bounded execution with strong guardrails

The target is not:

- a generic crawler framework
- a multi-channel messaging platform
- a plugin marketplace
- an unbounded agent with arbitrary tool access

## What To Copy From OpenClaw

These parts of `openclaw` are worth copying:

- The core framing: assistant runtime first, tools second, skills as operating guidance.
- A workspace contract where the agent runs inside a known directory.
- A tool registry with typed inputs and outputs.
- A skill loader that reads `SKILL.md`-style instructions from disk.
- A session/transcript model so runs are persistent and inspectable.
- A bounded agent loop: prompt, tool call, observation, next step, stop.

One part should be stricter than `openclaw`:

- `superclaw` should prefer containerized execution by default so the host system stays isolated from agent mistakes.

These parts should stay out of the first version:

- Multi-channel messaging.
- Gateway daemon plus WebSocket control plane.
- Mobile and desktop nodes.
- Canvas, voice, cron, browser automation, media generation.
- Large plugin and extension systems.
- Massive bundled skill catalog.

## OpenClaw vs Blaze-Claw

| Area | OpenClaw | Blaze-Claw |
| --- | --- | --- |
| Product shape | Personal AI assistant platform | Minimal local agent runtime |
| Runtime model | Gateway, sessions, channels, nodes, automations | Single local process with a simple run loop |
| Deployment model | Host-first with optional sandboxing and remote surfaces | Docker-first with minimal host exposure |
| Tool surface | Large built-in surface plus plugin tools | 4-5 built-in tools only |
| Skill surface | Bundled, managed, workspace, plugin, registry-backed | Local workspace skills with explicit allowlist |
| Ingress | WhatsApp, Telegram, Slack, Discord, CLI, apps, web | CLI first, optional HTTP later |
| Extensibility | Broad plugin ecosystem | Hardcoded interfaces, no marketplace |
| Autonomy | Can run long-lived background flows | Step-limited, budget-limited runs |
| Crawl/scrape support | Optional tool among many | Optional tool only, never core architecture |

## Minimum Important Tools

Keep the first version to these five tools max:

1. `workspace.read`
Reads files from the workspace and returns bounded content plus metadata.

2. `workspace.list`
Lists files and directories under the workspace with glob filtering and depth limits.

3. `shell.exec`
Runs a command in the workspace with timeout, output caps, and an allowlist of permitted binaries.

4. `workspace.patch`
Applies structured patches to files in the workspace. Prefer patch-based edits over raw writes.

5. `web.fetch`
Fetches a specific URL and returns normalized content. This is the only web capability needed at first.

### Why not more tools

- No browser automation in v1.
- No arbitrary HTTP client tool with full method control in v1.
- No cron, subagents, message delivery, or device tools in v1.
- No crawler pipeline tool. If we ever add crawl support, it should be a separate bounded tool built on top of `web.fetch`.
- No Docker socket access from inside the runtime.

### Optional v1.1 tool

If we later need one more capability, add:

- `web.search`

Do not add it before the five tools above are stable.

## Minimum Important Skills

Skills should stay small and map directly to the limited tool surface.

Start with only these three:

1. `coding`
Teaches safe repo exploration, patch-first editing, test-before-finish behavior, and when to use `shell.exec`.

2. `research`
Teaches disciplined use of `web.fetch`, source checking, and concise evidence-based summaries.

3. `github`
Optional skill enabled only when `gh` is installed and the runtime explicitly allows it through `shell.exec`.

### Skill rules

- Skills live in `skills/<name>/SKILL.md`.
- Skills are loaded only from the local workspace in v1.
- Skills must be allowlisted in config.
- No remote skill install.
- No plugin-provided skills.
- No dynamic skill discovery outside the configured directory.

## Target Runtime Architecture

The architecture should look like this:

```text
user/cli
  -> runtime
      -> session manager
      -> prompt builder
      -> model client
      -> tool registry
      -> skill registry
      -> transcript store
```

### Core execution loop

For each run:

1. Load config and workspace.
2. Resolve enabled skills for this agent.
3. Build the system prompt from base instructions plus selected skills.
4. Load recent session history.
5. Call the model.
6. If the model returns a tool call, validate it against the registry and policy.
7. Execute the tool with budgets and timeouts.
8. Append tool result to the transcript.
9. Repeat until final answer or max steps reached.
10. Persist the final assistant message and run metadata.

### Hard runtime limits

The runtime should enforce:

- max tool calls per run
- max wall-clock run duration
- max bytes returned per tool
- max file size read
- max command duration
- max URLs fetched per run

## Containerization And Host Safety

Docker should be a supported deployment target from day one.

The design goal is:

- the agent runs inside a container by default
- the host filesystem is not broadly mounted
- the container can be stopped and removed without affecting the host
- dangerous host capabilities are unavailable unless the user opts in

### Security stance

Docker is a containment layer, not a perfect security boundary.

We should document and design for:

- safe-by-default container settings
- a narrow workspace mount
- a non-root runtime user
- no privileged mode
- no host PID namespace
- no host network mode
- no Docker socket mount
- dropped Linux capabilities
- read-only root filesystem where practical
- explicit writable paths for runtime data only

### Recommended Docker profile

The default container deployment should use:

- `USER` set to a non-root user
- `read_only: true` or `--read-only`
- a bind mount only for the chosen workspace directory
- a separate named volume for transcripts, caches, and state
- `tmpfs` mounts for temporary files
- `security_opt: ["no-new-privileges:true"]`
- `cap_drop: ["ALL"]`
- CPU, memory, and PID limits
- a restart policy only if the user explicitly wants long-running behavior

### Workspace isolation model

Inside Docker, the runtime should assume:

- `/workspace` is the only user project mount
- `/data` stores transcripts, config, and runtime state
- `/tmp` is temporary and disposable

The built-in tools should respect these boundaries:

- `workspace.read`, `workspace.list`, and `workspace.patch` operate only under `/workspace`
- `shell.exec` runs with `/workspace` as cwd and cannot escape the allowed roots
- `web.fetch` has no direct filesystem access

### Crisis containment

If the agent behaves badly, the user should be able to recover by:

1. stopping the container
2. removing the container
3. deleting the isolated `/data` volume if needed
4. keeping the host system otherwise untouched

That means we should avoid designs that depend on:

- host-level daemons
- broad host mounts like `/` or `$HOME`
- privileged containers
- host package installation
- persistent background processes outside the container

### Two operating modes

Support two explicit modes:

1. `docker-safe`
Runs in Docker with strict mounts and reduced privileges. This should be the recommended default.

2. `local-dev`
Runs directly on the host for contributors developing the runtime itself.

`local-dev` should be clearly documented as less isolated.

## Suggested Go Layout

```text
cmd/superclaw/
  main.go

internal/config/
  config.go
  load.go

internal/runtime/
  runner.go
  loop.go
  policy.go

internal/session/
  manager.go
  transcript.go

internal/prompt/
  builder.go
  bootstrap.go

internal/model/
  client.go
  types.go

internal/tools/
  registry.go
  policy.go
  types.go
  workspace_read.go
  workspace_list.go
  workspace_patch.go
  shell_exec.go
  web_fetch.go

internal/skills/
  loader.go
  parser.go
  registry.go

internal/store/
  jsonl.go

deploy/
  docker/
    Dockerfile
    docker-compose.yml
    entrypoint.sh

pkg/superclaw/
  runtime.go
```

## Key Interfaces

```go
type Tool interface {
    Name() string
    Schema() ToolSchema
    Invoke(ctx context.Context, call ToolCall) (ToolResult, error)
}

type Skill struct {
    Name        string
    Description string
    Instructions string
}

type Runner interface {
    Run(ctx context.Context, req RunRequest) (RunResult, error)
}
```

## Config Shape

The initial config should stay small:

```yaml
workspace: ./
model:
  provider: openai
  name: gpt-5
runtime:
  max_steps: 12
  max_run_duration: 5m
tools:
  allow:
    - workspace.read
    - workspace.list
    - workspace.patch
    - shell.exec
    - web.fetch
skills:
  allow:
    - coding
    - research
runtime_mode: docker-safe
```

## Docker Packaging Requirements

When we implement packaging, the Docker assets should provide:

- a small production image
- a non-root runtime user
- a mounted `/workspace`
- a mounted `/data`
- healthcheck support
- environment-based model provider configuration
- no requirement for privileged Docker features

Example shape:

```text
docker run \
  --read-only \
  --tmpfs /tmp:rw,noexec,nosuid,size=64m \
  --cap-drop ALL \
  --security-opt no-new-privileges:true \
  --pids-limit 256 \
  --memory 512m \
  -v $(pwd):/workspace \
  -v superclaw-data:/data \
  superclaw:latest
```

This should be the documented baseline, not an advanced setup.

## Roadmap Order

Build in this order:

1. Config loader
2. Session/transcript store
3. Tool registry and the five built-in tools
4. Skill loader and allowlist enforcement
5. Single-session agent loop
6. Docker packaging and container-safe path enforcement
7. CLI entrypoint
8. Tests for bounds, policies, transcript persistence, and container-safe path rules

Only after that should we consider:

- `web.search`
- a small HTTP API
- one external channel
- a bounded crawl tool built on top of `web.fetch`

---

## Current Implementation Status

Package names differ slightly from the suggested layout above — they will converge as the runtime matures.

### Package map (planned → actual)

| Planned | Actual | Notes |
|---|---|---|
| `internal/runtime/` | `internal/agent/` | Loop, Config, State, Hooks, LLMClient, retry |
| `internal/model/` | `internal/agent/llm.go` | `LLMClient` interface + retry wrapper |
| `internal/tools/` | `internal/tools/` | Tool interface, Registry, 7 builtins |
| `internal/skills/` | `internal/skills/` | Skill interface, Registry, 5 builtins |
| `internal/config/` | `internal/config/` | `superclaw.json` loader (JSON, not YAML) |
| `internal/session/` | `internal/session/` | JSONL transcript store |
| `pkg/superclaw/` | `pkg/superclaw/` | Public `Run(ctx, task, opts)` API |

### Tool name map (planned → actual)

| Planned | Actual |
|---|---|
| `workspace.read` | `read_file` |
| `workspace.list` | `list_files` |
| `workspace.patch` | `patch_file` (exact str-replace; `write_file` kept for full overwrites) |
| `shell.exec` | `run_bash` |
| `web.fetch` | `fetch_url` |
| `web.search` | `web_search` (DuckDuckGo HTML, no API key) |

### What is working

- Agent loop: step limit, wall-clock timeout, tool dispatch, allowlist enforcement
- LLM retry: exponential backoff with full jitter on 429/5xx and network errors
- HTTP retry: `fetch_url` retries on 5xx and connection failures
- Config file (`superclaw.json`): model, steps, tokens, timeout, max_fetch_calls, workdir, tools, skills
- Progress hooks printed to stderr; session records appended to `.superclaw/runs.jsonl`
- Five builtin skills: `summarize`, `research`, `extract`, `coding`, `github`
- Docker: two-stage image, non-root user, `/workspace` + `/data` volumes, read-only rootfs, capability drops
- Unit tests: agent loop, tool registry, file tools, config, session store, skills

### Not yet implemented

- HTTP API / programmatic ingress (v1.1)
- Multi-turn session history injected into prompts (v1.1)
- Bounded crawl tool built on `fetch_url` (v1.1)
