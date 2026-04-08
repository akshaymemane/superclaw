# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**superclaw** should be built as a compact [openclaw](https://github.com/openclaw/openclaw)-style agent runtime in Go, not as a generic web crawler/scraper.

The primary goal is an opinionated, minimal agent that can plan, call a small number of well-defined tools, and use a curated set of skills. If crawling or scraping exists at all, it should be only one bounded tool in that system, not the product's central architecture.

See `ARCHITECTURE.md` for the concrete comparison to OpenClaw, the minimum tool and skill set, and the target Go package layout.

Container-first safety is part of the design. Prefer a Docker deployment model with strict mounts and limited privileges so agent failures do not broadly affect the host machine.

## Commands

```bash
go run ./cmd/superclaw        # run the blaze-claw runtime
go build ./...                 # build all binaries
go test ./...                  # run all tests
go test ./internal/...         # run internal package tests
go test -run TestName ./pkg/.. # run a single test by name
go vet ./...                   # static analysis
```

## Intended Architecture

Prefer a standard Go project layout:

```
cmd/blaze-claw/       # main entrypoint
internal/             # private packages (runtime, tools, skills, execution loop)
pkg/                  # public/reusable packages (small exportable API surface)
```

Key design principles to follow:
- **Not crawler-first**: do not default to designing fetch → parse → extract → enqueue pipelines as the core system. That is only acceptable for a specific tool, not the full architecture.
- **Limited tool surface**: keep the first version intentionally small, for example 3-6 tools max. Favor strongly typed, auditable tools over arbitrary tool execution.
- **Limited skill surface**: support only a curated, explicit set of skills. Avoid dynamic skill discovery, broad plugin ecosystems, or uncontrolled capability growth.
- **Agent loop over crawl loop**: prioritize a simple execution model such as plan → act → observe → stop, with bounded steps, budgets, and cancellation.
- **Human-controlled scope**: every tool call should have clear input/output contracts, timeouts, and guardrails. The runtime should be easy to inspect and reason about.
- **Container-first isolation**: design the runtime so it works cleanly inside Docker with a narrow workspace mount, isolated state volume, and no need for privileged host access.
- **Concurrency where it helps**: use goroutines/channels for bounded parallel tool execution or background work, but do not optimize for maximum crawl throughput as the primary design goal.
- **Minimal dependencies**: prefer stdlib and small focused packages. Add dependencies only when they clearly support the agent runtime.

## Non-Goals

- Do not turn this into a general-purpose crawler framework.
- Do not build an unbounded autonomous agent with arbitrary tool access.
- Do not introduce a large marketplace of tools or skills in the initial design.
- Do not optimize for scraping throughput before the core runtime, tool registry, and skill model are solid.
- Do not require privileged Docker features or broad host mounts for normal operation.

## Initial Focus

If the codebase is being scaffolded from scratch, prioritize these pieces first:

1. `agent` or `runtime` core that owns config, state, and the execution loop.
2. Small tool registry with a hardcoded or manifest-backed allowlist.
3. Small skill registry with explicit loading and validation.
4. Shared interfaces for tool invocation, skill selection, and execution results.
5. Container-safe path and process boundaries that work cleanly in Docker.
6. Tests that prove the runtime stays bounded and only uses approved tools/skills.
