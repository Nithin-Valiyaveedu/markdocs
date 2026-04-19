package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// ReviewAction represents what the user chose during draft review.
type ReviewAction int

const (
	ReviewAccept  ReviewAction = iota
	ReviewEdit
	ReviewDiscard
)

// ReviewDraft shows a preview of the compiled skill draft and prompts the user
// to Accept, Edit in $EDITOR, or Discard before the file is written to disk.
func ReviewDraft(draftMarkdown string) (ReviewAction, string, error) {
	Section("Draft Preview")
	fmt.Println(renderDraftPreview(draftMarkdown))

	const (
		optAccept  = "Accept as-is"
		optEdit    = "Edit in $EDITOR"
		optDiscard = "Discard"
	)

	var chosen string
	err := huh.NewSelect[string]().
		Title("What would you like to do with this draft?").
		Options(
			huh.NewOption(optAccept, optAccept),
			huh.NewOption(optEdit, optEdit),
			huh.NewOption(optDiscard, optDiscard),
		).
		Value(&chosen).
		Run()
	if err != nil {
		return ReviewDiscard, "", fmt.Errorf("review prompt: %w", err)
	}

	switch chosen {
	case optEdit:
		edited, err := openInEditor(draftMarkdown)
		if err != nil {
			return ReviewDiscard, "", fmt.Errorf("editor: %w", err)
		}
		return ReviewEdit, edited, nil
	case optDiscard:
		return ReviewDiscard, "", nil
	default:
		return ReviewAccept, draftMarkdown, nil
	}
}

// renderDraftPreview returns a lipgloss-bordered preview of the first 50 lines.
func renderDraftPreview(markdown string) string {
	lines := strings.Split(markdown, "\n")
	const maxLines = 50
	var body string
	if len(lines) > maxLines {
		body = strings.Join(lines[:maxLines], "\n") +
			"\n" + StyleMuted.Render(fmt.Sprintf("  … %d more lines", len(lines)-maxLines))
	} else {
		body = markdown
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Padding(0, 1).
		MaxWidth(100).
		Render(body)
}

// openInEditor writes content to a temp file, opens it in $EDITOR (fallback: nano),
// waits for the editor to exit, then returns the (possibly modified) content.
func openInEditor(content string) (string, error) {
	tmp, err := os.CreateTemp("", "markdocs-draft-*.md")
	if err != nil {
		return "", fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(content); err != nil {
		return "", fmt.Errorf("writing temp file: %w", err)
	}
	tmp.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	editorCmd := exec.Command(editor, tmp.Name())
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	if err := editorCmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %w", err)
	}

	result, err := os.ReadFile(tmp.Name())
	if err != nil {
		return "", fmt.Errorf("reading edited file: %w", err)
	}
	return string(result), nil
}

// Select shows an interactive single-select prompt and returns the chosen value.
func Select(title string, options []string) (string, error) {
	opts := make([]huh.Option[string], len(options))
	for i, o := range options {
		opts[i] = huh.NewOption(o, o)
	}
	var result string
	err := huh.NewSelect[string]().
		Title(title).
		Options(opts...).
		Value(&result).
		Run()
	return result, err
}

// SelectWithManual shows a select prompt with an extra "Enter manually" option.
// If the user picks "Enter manually", they are prompted to type a value.
func SelectWithManual(title string, options []string) (string, error) {
	const manualOption = "↵  Enter manually"
	all := append(options, manualOption)

	chosen, err := Select(title, all)
	if err != nil {
		return "", err
	}
	if chosen == manualOption {
		return Input("Enter URL", "https://")
	}
	return chosen, nil
}

// Input shows a text input prompt with a placeholder default.
func Input(title, placeholder string) (string, error) {
	var result string
	err := huh.NewInput().
		Title(title).
		Placeholder(placeholder).
		Value(&result).
		Run()
	return result, err
}

// InputWithDefault shows a text input prompt with a pre-filled default value.
func InputWithDefault(title, defaultVal string) (string, error) {
	result := defaultVal
	err := huh.NewInput().
		Title(title).
		Value(&result).
		Run()
	return result, err
}

// Secret shows a masked input prompt (for API keys, passwords).
func Secret(title string) (string, error) {
	var result string
	err := huh.NewInput().
		Title(title).
		EchoMode(huh.EchoModePassword).
		Value(&result).
		Run()
	return result, err
}

// Confirm shows a yes/no confirmation prompt.
func Confirm(question string) (bool, error) {
	var result bool
	err := huh.NewConfirm().
		Title(question).
		Affirmative("Yes").
		Negative("No").
		Value(&result).
		Run()
	return result, err
}

// SelectProvider shows a provider selection menu and returns the selected index.
func SelectProvider(detected string) (int, error) {
	items := []string{
		"Anthropic (Claude)",
		"OpenAI (GPT)",
		"OpenAI-compatible (Groq, Together, etc.)",
		"Ollama (local)",
	}
	opts := make([]huh.Option[int], len(items))
	for i, item := range items {
		label := item
		if detected != "" && containsProvider(detected, item) {
			label = item + "  " + StyleSuccess.Render("← detected")
		}
		opts[i] = huh.NewOption(label, i)
	}

	var idx int
	err := huh.NewSelect[int]().
		Title("Select LLM provider").
		Options(opts...).
		Value(&idx).
		Run()
	return idx, err
}

func containsProvider(detected, item string) bool {
	switch detected {
	case "anthropic":
		return item == "Anthropic (Claude)"
	case "openai":
		return item == "OpenAI (GPT)"
	case "ollama":
		return item == "Ollama (local)"
	}
	return false
}

// UseDetected asks if the user wants to use an auto-detected provider.
func UseDetected(providerName string) (bool, error) {
	return Confirm(fmt.Sprintf("Detected %s from environment — use it?", providerName))
}
