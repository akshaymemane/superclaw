# superclaw

**A minimal, opinionated Go agent runtime.**

superclaw runs LLM-powered tasks with bounded steps, curated tools, and Docker-first isolation. It plans, acts, and stops — it does not run forever.

[![Go 1.23+](https://img.shields.io/badge/go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/license-MIT-orange)](LICENSE)
[![Anthropic](https://img.shields.io/badge/powered%20by-Anthropic-blueviolet)](https://anthropic.com)

---

## What it is

superclaw is an agent runtime, not a crawler and not an autonomous agent with unlimited reach. It wraps a tight execution loop:

```
load config → build system prompt → call model → execute tools → observe → repeat → stop
```

Hard limits are enforced at every step: maximum LLM calls, wall-clock timeout, maximum URLs fetched. When the task is done (or limits are hit), it stops and writes a session record.

---

## Install

```bash
go install github.com/akshaymemane/superclaw/cmd/superclaw@latest
```

**Requirements**: Go 1.23+, an [Anthropic API key](https://console.anthropic.com/).

---

## Quick start

```bash
export ANTHROPIC_API_KEY=sk-ant-...

# Run a task
superclaw "read the README and write a one-paragraph summary to summary.md"

# Constrain to a specific directory
superclaw -workdir ./myproject "list the Go files and explain each package"

# Activate skills
superclaw -skills coding "find and fix the bug in internal/parser/parser.go"

# With a config file
cp superclaw.json.example superclaw.json
superclaw "your task here"
```

---

## Tools

superclaw ships with 7 built-in tools. Pass a comma-separated allowlist with `-tools` or in `superclaw.json` to restrict which tools are active.

| Tool | Description |
|---|---|
| `fetch_url` | HTTP GET with 32 KB cap, retries on 5xx |
| `web_search` | DuckDuckGo search, no API key required |
| `read_file` | Read a file (64 KB cap, path-constrained) |
| `write_file` | Write or overwrite a file |
| `patch_file` | Targeted str-replace edit (old_text must match exactly once) |
| `list_files` | List files with glob filtering and depth limit |
| `run_bash` | Run a shell command (30s timeout, blocked patterns list) |

---

## Skills

Skills inject focused instructions into the system prompt. They do not add tools or change the tool surface.

| Skill | Description |
|---|---|
| `summarize` | Lead with conclusion, structured headers, no filler |
| `research` | 2–4 sources, cross-reference, cite URLs |
| `extract` | Schema-first extraction, `NOT_FOUND` for missing fields |
| `coding` | Read before edit, prefer `patch_file`, smallest change, run tests |
| `github` | Verify `gh` installed, read-only first, confirm before submit |

Activate via `-skills` flag or `superclaw.json`:

```bash
superclaw -skills coding,github "open a PR with the fix"
```

---

## Configuration

Copy `superclaw.json.example` to `superclaw.json` in your working directory. All fields are optional.

```json
{
  "model": "claude-opus-4-6",
  "max_steps": 20,
  "max_tokens": 16000,
  "timeout_seconds": 300,
  "work_dir": ".",
  "max_fetch_calls": 10,
  "tools": ["fetch_url", "web_search", "read_file", "write_file", "run_bash", "list_files", "patch_file"],
  "skills": ["coding"],
  "system_prompt": ""
}
```

### CLI flags

All flags override the corresponding `superclaw.json` value.

| Flag | Description |
|---|---|
| `-config` | Path to config file (default: `superclaw.json`) |
| `-workdir` | Constrain file operations to this directory |
| `-max-steps` | Override `max_steps` |
| `-tools` | Comma-separated tool allowlist |
| `-skills` | Comma-separated skills to activate |
| `-quiet` | Suppress progress output to stderr |

---

## Docker

The recommended deployment model. The agent runs as a non-root user with a read-only root filesystem, capability drops, and a narrow workspace mount.

```bash
docker run \
  --rm \
  --read-only \
  --tmpfs /tmp:rw,noexec,nosuid,size=64m \
  --cap-drop ALL \
  --security-opt no-new-privileges:true \
  --pids-limit 256 \
  --memory 512m \
  -e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY \
  -v $(pwd):/workspace \
  -v superclaw-data:/data \
  ghcr.io/akshaymemane/superclaw:latest \
  "your task here"
```

Or with Docker Compose:

```bash
cd deploy/docker
HOST_WORKDIR=$(pwd)/../../ docker compose run --rm superclaw "your task here"
```

Build the image locally:

```bash
docker build -f deploy/docker/Dockerfile -t superclaw .
```

---

## Session history

Every run appends a record to `.superclaw/runs.jsonl` in the work directory:

```json
{"id":"20260408-143022.123","timestamp":"2026-04-08T14:30:22Z","task":"...","result":"...","steps":4,"status":"success"}
```

Read history:

```bash
cat .superclaw/runs.jsonl | python3 -m json.tool
```

---

## Retry behavior

superclaw retries automatically on transient failures:

- **LLM calls**: up to 4 attempts, exponential backoff with full jitter, on HTTP 429/500/502/503/504 and network errors. Context cancellation and 4xx client errors are not retried.
- **`fetch_url`**: up to 3 attempts on HTTP 5xx and connection failures.

---

## Architecture

```
cmd/superclaw/        CLI entrypoint
internal/
  agent/               Execution loop, LLM client, retry, config, hooks, state
  config/              superclaw.json loader
  session/             JSONL transcript store
  skills/              Skill interface, registry, 5 built-ins
  tools/               Tool interface, registry, 7 built-ins
pkg/superclaw/             Public Run() API
deploy/docker/         Dockerfile, docker-compose.yml, entrypoint.sh
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for the full design rationale and a comparison with openclaw.

---

## Development

```bash
go build ./...          # build
go test ./...           # run all tests
go vet ./...            # static analysis
go test -run TestName ./internal/agent/   # single test
```

---

## License

MIT. See [LICENSE](LICENSE).
