package config

import (
	"net/http"
	"os"
	"time"
)

// ollamaHealthURL is the URL used to check if Ollama is running locally.
// It's a package-level variable so tests can override it.
var ollamaHealthURL = "http://localhost:11434/api/tags"

// DetectProvider inspects environment variables and a local Ollama instance
// to determine which LLM provider is available. Returns the provider name
// and true if one is found, or ("", false) if nothing is detected.
//
// Detection order:
//  1. ANTHROPIC_API_KEY env var → ProviderAnthropic
//  2. OPENAI_API_KEY env var → ProviderOpenAI
//  3. Ollama running at localhost:11434 → ProviderOllama
func DetectProvider() (ProviderName, bool) {
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return ProviderAnthropic, true
	}
	if os.Getenv("OPENAI_API_KEY") != "" {
		return ProviderOpenAI, true
	}
	if ollamaRunning() {
		return ProviderOllama, true
	}
	return "", false
}

func ollamaRunning() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(ollamaHealthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
