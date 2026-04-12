---
name: markdocs
description: Use this skill when you need to compile library documentation into Claude Code skill files, scan a project for missing documentation skills, or update existing skills when dependencies change.
compatibility: Requires go 1.22+ OR npm OR brew. Requires an LLM provider configured via `markdocs init` (supports Anthropic, OpenAI, OpenAI-compatible endpoints, and local Ollama).
license: MIT
metadata:
  author: Nithin-Valiyaveedu
  version: "0.1.0"
---

# `markdocs` Skill

Compile library documentation into structured Claude Code skill files that Claude picks up automatically.

## Initial Setup

When this skill is invoked, confirm:
- `markdocs` is installed (see Installation below)
- A provider is configured (`markdocs init`)

## Installation

```bash
# Go
go install github.com/Nithin-Valiyaveedu/markdocs@latest

# npm
npm install -g @Nithin-Valiyaveedu/markdocs

# Homebrew
brew install Nithin-Valiyaveedu/markdocs/markdocs
```

## Step 1: Configure a Provider

```bash
markdocs init
```

Auto-detects from environment (`ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, or local Ollama). Supports OpenAI-compatible endpoints (Groq, Together AI) via base URL override.

## Step 2: Add Skills

```bash
markdocs add <library>
```

Example: `markdocs add shadcn` → writes `.claude/skills/ui/shadcn.md`

Flags:
| Flag | Description |
|------|-------------|
| `--no-interactive` | Skip URL selection, use first suggested URL |

## Step 3: Manage Skills

| Command | Description |
|---------|-------------|
| `markdocs scan` | Detect missing skills from package.json / go.mod |
| `markdocs scan --add-all` | Auto-add all missing skills |
| `markdocs list` | Show compiled skills + age |
| `markdocs list --stale` | Show only skills older than 7 days |
| `markdocs update <name>` | Recompile if source changed |
| `markdocs update --all` | Check and recompile all skills |

## How It Works

1. LLM discovers official documentation URLs for the library
2. User selects which URL(s) to scrape (interactive)
3. Built-in Go scraper fetches and cleans the content (no external APIs)
   - Layer 1: `net/http` + `go-readability` (static sites)
   - Layer 2: `go-rod` headless browser (JS-heavy sites, fallback)
4. LLM compiles scraped content into a structured skill file
5. Skill is written to `.claude/skills/<category>/<library>.md`
6. Claude Code picks it up automatically next session

## Skill File Format

Skills are plain Markdown with YAML frontmatter. The frontmatter stores all metadata — no database required:

```markdown
---
name: react
category: frontend
sources:
  - https://react.dev/learn
compiled: 2026-04-12T10:23:00Z
checksum: sha256:abc123def456
model: llama3.2
provider: ollama
project_framework: next.js
markdocs_version: 0.1.0
---

# react — markdocs skill
...
```

## Step 4: Report

After running any command, confirm to the user:
- Which skills were compiled / updated / skipped
- Where skill files were written
- Any errors encountered
