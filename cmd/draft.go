package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Nithin-Valiyaveedu/markdocs/internal/skill"
	"github.com/Nithin-Valiyaveedu/markdocs/internal/ui"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var draftCmd = &cobra.Command{
	Use:   "draft <name>",
	Short: "Compile your own notes or docs into a skill file.",
	Long: `Prompts you to paste documentation, notes, or any text content.
The LLM compiles it into a structured skill file, shows a draft preview,
and lets you edit before writing to .claude/skills/<category>/<name>.md.

You can also pipe content directly:

  cat docs.md | markdocs draft mylib`,
	Args: cobra.ExactArgs(1),
	RunE: runDraft,
}

func runDraft(cmd *cobra.Command, args []string) error {
	library := args[0]
	ctx := cmd.Context()

	compiler, providerName, modelName := newCompiler()

	// Step 1: Get content from stdin (pipe) or interactive textarea
	var content string
	if isStdinPipe() {
		raw, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		content = string(raw)
		ui.Info(fmt.Sprintf("Read %d chars from stdin", len(content)))
	} else {
		ui.Info("Paste your documentation, notes, or any relevant content below.")
		ui.Info("Press Ctrl+D or Esc to finish.")
		ui.Blank()
		err := huh.NewText().
			Title(fmt.Sprintf("Content for %q skill", library)).
			Placeholder("Paste documentation, API references, examples...").
			Value(&content).
			Run()
		if err != nil {
			return fmt.Errorf("content input: %w", err)
		}
	}

	if len(strings.TrimSpace(content)) < 50 {
		ui.Warning("Content is too short — add more detail and try again.")
		return nil
	}

	// Step 2: Compile
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	var compiled *skill.CompileOutput
	err = ui.Spin("Compiling skill...", func() error {
		framework := skill.DetectFramework(cwd)
		var err error
		compiled, err = compiler.Compile(ctx, skill.CompileInput{
			Library:          library,
			URL:              "user-provided notes",
			ScrapedMarkdown:  content,
			ProjectFramework: framework,
			Model:            modelName,
		})
		return err
	})
	if err != nil {
		return fmt.Errorf("compiling skill for %s: %w", library, err)
	}
	ui.StepDone(1, fmt.Sprintf("Compiled skill (category: %s)", compiled.Category))

	// Step 3: Review draft
	action, finalMarkdown, err := ui.ReviewDraft(compiled.Markdown)
	if err != nil {
		return fmt.Errorf("draft review: %w", err)
	}
	if action == ui.ReviewDiscard {
		ui.Warning("Discarded — no skill written.")
		return nil
	}
	if action == ui.ReviewEdit {
		compiled.Markdown = finalMarkdown
	}

	// Step 4: Write
	framework := skill.DetectFramework(cwd)
	meta := skill.NewSkillMeta(
		library,
		providerName,
		modelName,
		compiled.Category,
		framework,
		compiled.Description,
		compiled.WhenToUse,
		[]string{"user-provided notes"},
		content,
	)
	writer := skill.NewFSWriter(cwd)
	path, err := writer.Write(library, compiled.Category, compiled.Markdown, meta)
	if err != nil {
		return fmt.Errorf("writing skill for %s: %w", library, err)
	}

	ui.Blank()
	ui.Success(fmt.Sprintf("Skill ready: %s", path))
	ui.Info("Claude Code will pick it up automatically next session.")
	return nil
}

// isStdinPipe reports whether stdin is a pipe (non-interactive).
func isStdinPipe() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice == 0
}
