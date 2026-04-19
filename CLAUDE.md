# markdocs

Go CLI that helps you draft and create custom Claude Code skill files from existing documentation (`.claude/skills/<category>/<library>.md`).

## Commands

```bash
go build ./...              # build
go test ./...               # run all tests
go run main.go <command>    # run locally
```

## Key Architecture Decisions

- **No external scraping APIs** — own Go waterfall scraper only (Layer 1: net/http + go-readability → Layer 2: go-rod headless browser, fallback when content < 500 chars)
- **Provider-agnostic LLM interface** — `LLMProvider` interface in `internal/llm/provider.go`; never hardcode a specific provider
- **No SQLite / no CGO** — metadata lives entirely in each skill file's YAML frontmatter (`SkillMeta`), not a database
- **Single binary** — no runtime requirements for end users; distributed via npm and goreleaser
- **Skills are plain markdown** — written to `.claude/skills/<category>/<library>.md`; Claude Code picks them up automatically next session
- **Draft-review before write** — after LLM compilation, users review and optionally edit the skill in `$EDITOR` before it's written to disk; `--no-interactive` skips this

## Directory Structure

```
cmd/            → cobra commands (root, add, draft, scan, update, list, init, pipeline)
internal/
  config/       → read/write ~/.markdocs/config.json
  llm/          → LLMProvider interface + Anthropic / OpenAI / Ollama backends
  scraper/      → waterfall scraper (http.go → rod.go), markdown conversion
  skill/        → compiler, writer, reader, scanner, registry
  ui/           → banner, spinner, prompts, table, output styles
skills/
  markdocs/     → SKILL.md for markdocs itself
```

## Config

Stored at `~/.markdocs/config.json`. Providers: `anthropic`, `openai`, `ollama`.

Auto-detect order on startup:
1. `ANTHROPIC_API_KEY` env → Anthropic
2. `OPENAI_API_KEY` env → OpenAI
3. Ollama at `localhost:11434` → Ollama
4. Nothing → prompt `markdocs init`

OpenAI provider also covers Groq, Together, and any OpenAI-compatible endpoint via `base_url` override.

## Skill File Format

```yaml
---
name: <library>
category: frontend|backend|testing|infra|database|payments|auth|devtools
sources:
  - https://...
compiled: <RFC3339>
checksum: sha256:<hex>   # of scraped source content, used by `update`
model: <model-id>
provider: anthropic|openai|ollama
project_framework: <detected>
markdocs_version: 0.1.0
---
```

Followed by LLM-compiled markdown with sections: What This Is, Installation, Key Concepts, API / Usage Patterns, Your Project Config, Hidden Gotchas, Common Errors, Version Notes.

## Adding a New LLM Provider

1. Create `internal/llm/<name>.go` implementing `LLMProvider` (`Complete`, `Model`)
2. Add a case to `NewProvider()` in `internal/llm/provider.go`
3. Add detection logic in `internal/config/detect.go` if it has a standard env var
4. Add a `ProviderName` constant in `internal/config/config.go`

## Adding a New Command

1. Create `cmd/<name>.go` with a `cobra.Command`
2. Register it in `cmd/root.go` `init()`

See `cmd/draft.go` as a recent example.

## Shared Pipeline

`cmd/pipeline.go:runAddPipeline` contains the reusable scrape → compile → [review] → write flow used by `add`, `scan --add-all`, and `update`.

Signature: `runAddPipeline(ctx, library, url, provider, cwd, noInteractive bool) (*PipelineResult, error)`

- `noInteractive = false` (default for `add`): shows the draft review prompt (`ui.ReviewDraft`) after compile, before write. Returns `nil, nil` if the user discards.
- `noInteractive = true` (used by `scan --add-all` and `update`): skips review, writes immediately.

### Draft Review (`ui.ReviewDraft`)

Defined in `internal/ui/prompts.go`. Shows a bordered preview of the first 50 lines of the compiled skill, then prompts: Accept / Edit in $EDITOR / Discard.

- **Edit**: writes content to a temp `.md` file, opens `$EDITOR` (fallback `nano`), re-reads on exit
- **Accept**: passes markdown through unchanged
- **Discard**: returns `ReviewDiscard`; caller returns `nil, nil` to signal cancellation

`ReviewAction` constants: `ReviewAccept`, `ReviewEdit`, `ReviewDiscard` (in `internal/ui/prompts.go`).

## Code Style

- Errors: always wrap with context — `fmt.Errorf("doing X: %w", err)`
- No global state — pass dependencies explicitly
- Config loaded once in `PersistentPreRunE`, stored in `appConfig`, passed down
- `appConfig` is nil when running `init` (config doesn't exist yet)

## Testing

- Unit tests alongside source files (`*_test.go`)
- Scraper tests in `internal/scraper/scraper_test.go`
- Skill tests in `internal/skill/` (`compiler_test.go`, `writer_test.go`, `scanner_test.go`, `reader_test.go`)
- Integration tests should use recorded HTTP responses — no live network fetches in CI

## Release / Distribution

- `goreleaser` builds multi-platform binaries (`.goreleaser.yaml`)
- npm wrapper in `package.json` for `npx markdocs` distribution
- Version injected at build time: `-ldflags "-X 'github.com/Nithin-Valiyaveedu/markdocs/cmd.Version=<ver>'"`, falls back to `"dev"`
