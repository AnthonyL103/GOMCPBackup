# GoMCP — Project One-Pager

## Problem

Building AI agents that can use external tools requires significant boilerplate: provider-specific API formatting, tool schema management, server lifecycle control, and conversation history tracking. Developers rewrite this for every project. Non-technical users are locked out entirely, and there's no simple way to leverage MCP-style tool orchestration without writing code.

## Solution

GoMCP is a Go-based Model Context Protocol framework that lets AI agents dynamically discover, invoke, and even **create** tool servers at runtime. Configuration is entirely YAML-driven, meaning configuration (prompts/declarations) never overlap with code and adding tools becomes as simple as writing/generating functions (language agnostic) and specifying the file path, function_names, and purpose. A planned hosted web product will extend this to non-technical users through a browser-based chat interface.

## Unique Value Proposition

| Differentiator | What It Means |
|----------------|---------------|
| **Runtime Server Generation** | The LLM generates, compiles, tests, and deploys new tool servers mid-conversation — no human intervention required |
| **Language-Agnostic Tools** | Any HTTP server (Go, Python, Node, etc.) works as a tool server — just expose `/execute/{handler}` |
| **Zero-Code Configuration** | YAML files define agents, servers, tools, and schemas — add capabilities without touching source code |
| **Multi-Provider Architecture** | Swap between Anthropic and OpenAI models with a single config change; provider-specific formatting handled internally |

## Key Features (Current State)

- **Multi-provider LLM support** — Anthropic Claude and OpenAI GPT families (11 models)
- **Dynamic tool loading** — tools parsed from YAML configs at startup
- **MCP server architecture** — extensible HTTP-based server system with process lifecycle management
- **Automatic tool chaining** — sequential (Anthropic) and parallel (OpenAI) tool execution loops
- **AI-driven server generation** — 5-stage validated pipeline: generate → test → deploy → register → cleanup
- **Type-safe schemas** — JSON Schema validation with recursive nested object/array support
- **Conversation history** — sliding-window chat management with full tool call/result tracking
- **Voice chat framework** — session state and input parsing scaffolded (integration pending)

## Architecture

```
User Input
    │
    ▼
┌──────────────────────────────────────────────────┐
│  Agent                                           │
│  ├── Config (YAML)    ├── Chat History           │
│  ├── Instructions     └── Voice Session (future) │
└──────┬────────────────────┬──────────────────────┘
       │                    │
       ▼                    ▼
┌──────────────┐    ┌──────────────────────┐
│  Transport   │    │  Registry            │
│  ├ Anthropic │    │  ├ Static Servers    │
│  └ OpenAI    │    │  └ Generated Servers │
└──────┬───────┘    └──────────┬───────────┘
       │                       │
       ▼                       ▼
   LLM APIs            MCP Tool Servers
  (Claude/GPT)         (HTTP on :PORT)
```

## Target Users

| Segment | Timeline | Access Mode |
|---------|----------|-------------|
| **Developers** building AI agents with tool integrations | Now | CLI + YAML config |
| **Non-technical users** needing AI-powered tool orchestration | Planned (8-week roadmap) | Web UI with chat interface |

## Tech Stack

| Layer | Current | Planned (Web Product) |
|-------|---------|----------------------|
| **Core** | Go 1.25 | Go HTTP API |
| **LLM Providers** | Anthropic, OpenAI | Same |
| **Config** | YAML (gopkg.in/yaml.v3) | Same + DB-backed |
| **Frontend** | CLI | React + Vite + TypeScript |
| **Auth & DB** | — | AWS RDS/Clickhouse |
| **Secrets** | Env vars / config | Encrypted Postgres + server-side key |
| **Hosting** | Local | AWS EC2 (backend — requires persistent child processes), AWS Amplify / CloudFront + S3 (frontend) |

## Roadmap Summary

```
Sprint 1 (Wk 1-2)    Sprint 2 (Wk 3-4)    Sprint 3 (Wk 5-6)    Sprint 4 (Wk 7-8)
┌───────────────┐    ┌───────────────┐    ┌───────────────┐    ┌───────────────┐
│  Web MVP      │    │  Multi-User   │    │  MCP Power    │    │  Production   │
│               │    │               │    │               │    │               │
│ • Chat UI     │ -> │ • Auth flows  │ -> │ • Tool gen UI │ -> │ • Rate limits │
│ • Go HTTP API │    │ • DB sessions │    │ • Remote APIs │    │ • Monitoring  │
│ • Deploy v1   │    │ • Secrets mgmt│    │ • Cloud arch  │    │ • Diagram MVP │
└───────────────┘    └───────────────┘    └───────────────┘    └───────────────┘
```

**Milestone targets:** Hosted chat (Wk 2) → Auth + secrets (Wk 4) → Tool gen + remote APIs (Wk 6) → Hardened launch (Wk 8)

## Constraints

- **Schedule:** 10–15 hours/week, 8 part-time weeks
- **Generated servers:** max 5 concurrent, ports allocated from 9000+
- **Server generation:** Go-only (currently)
- **Priority:** deploy quickly, reduce complexity — no Kubernetes, no over-engineering

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Scope creep beyond part-time capacity | Delayed delivery | Freeze advanced UI; weeks 1–4 strictly MVP + auth |
| Security mistakes in secret handling | Data exposure | Server-side encryption, never return plaintext, redact in logs |
| Tool generation produces broken servers | Poor UX | 5-stage validated pipeline with per-step LLM feedback loops |
| Hosting complexity | Ops burden | Managed platforms only (Render, Vercel, Supabase) |

## Success Criteria

- A non-technical user can send and receive messages from a browser (Sprint 1)
- Users see only their own data; secrets are never exposed (Sprint 2)
- Users can generate tool servers and configure remote APIs via UI (Sprint 3)
- System handles realistic class/project traffic with documented recovery steps (Sprint 4)
