package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

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
