---
name: markdocs
description: Use this skill when you need to compile library documentation into Claude Code skill files, scan a project for missing documentation skills, or update existing skills when dependencies change.
compatibility: Requires go 1.22+ OR npm OR brew. No API key required by default. Use --llm flag with a configured provider (markdocs init) for LLM-compiled output.
license: MIT
metadata:
  author: Nithin-Valiyaveedu
  version: "0.3.0"
---

# `markdocs` Skill

Compile library documentation into structured Claude Code skill files that Claude picks up automatically.

## Initial Setup

When this skill is invoked, confirm:
- `markdocs` is installed (see Installation below)
- For default structured mode: no API key needed
- For `--llm` mode: a provider is configured (`markdocs init`)

## Installation

```bash
# Go
go install github.com/Nithin-Valiyaveedu/markdocs@latest

# npm
npm install -g @Nithin-Valiyaveedu/markdocs

# Homebrew
brew install Nithin-Valiyaveedu/markdocs/markdocs
```

## Step 1: Add Skills (no API key required)

```bash
markdocs add <library>
```

Example: `markdocs add zustand` → writes `.claude/skills/frontend/zustand.md`

By default, markdocs uses **structured extraction** — it scrapes the official docs and maps headings to skill sections deterministically. No LLM, no API key.

Flags:
| Flag | Description |
|------|-------------|
| `--llm` | Use LLM compilation instead of structured extraction (requires `markdocs init`) |
| `--no-interactive` | Skip URL selection prompt, use first suggested URL |

## Step 2: Manage Skills

| Command | Description |
|---------|-------------|
| `markdocs scan` | Detect missing skills from package.json / go.mod |
| `markdocs scan --add-all` | Auto-add all missing skills |
| `markdocs list` | Show compiled skills + age |
| `markdocs list --stale` | Show only skills older than 7 days |
| `markdocs update <name>` | Recompile if source changed |
| `markdocs update --all` | Check and recompile all skills |

## How It Works

### Default (structured extraction)
1. Web search finds official documentation URLs for the library
2. User selects which URL to scrape (interactive)
3. Built-in Go scraper fetches and cleans the content
   - Layer 1: `net/http` + `go-readability` (static sites)
   - Layer 2: `go-rod` headless browser (JS-heavy sites, fallback)
4. Heading-based extractor maps doc sections to skill sections
5. Skill is written to `.claude/skills/<category>/<library>.md`
6. Claude Code picks it up automatically next session

### With `--llm`
Steps 1–3 are the same. Step 4 sends scraped content to the configured LLM provider (Anthropic, OpenAI, or Ollama) for richer compilation.

## Skill File Format

Skills are plain Markdown with YAML frontmatter. The frontmatter stores all metadata — no database required:

```markdown
---
name: react
description: React is a JavaScript library for building UIs...
when_to_use: user mentions react, building a component, JSX
user-invocable: false
category: frontend
sources:
  - https://react.dev/learn
compiled: 2026-04-19T10:23:00Z
checksum: sha256:abc123def456
model: ""
provider: structured
project_framework: next.js
markdocs_version: 0.2.1
---

# react — markdocs skill
...
```

## Step 3: Report

After running any command, confirm to the user:
- Which skills were compiled / updated / skipped
- Where skill files were written
- Any errors encountered
