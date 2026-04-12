package config

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// Temporarily override DefaultConfigPath via a file write
	cfg := &Config{
		Provider: ProviderAnthropic,
		Model:    "claude-haiku-4-5-20251001",
		APIKey:   "test-key",
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o600))

	// Read it back manually (can't easily override DefaultConfigPath in test,
	// so test the marshal/unmarshal round-trip directly)
	var loaded Config
	require.NoError(t, json.Unmarshal(data, &loaded))
	assert.Equal(t, cfg.Provider, loaded.Provider)
	assert.Equal(t, cfg.Model, loaded.Model)
	assert.Equal(t, cfg.APIKey, loaded.APIKey)
}

func TestSaveCreatesDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := &Config{
		Provider: ProviderOllama,
		Model:    "llama3.2",
		OllamaHost: "http://localhost:11434",
	}
	require.NoError(t, Save(cfg))

	path, err := DefaultConfigPath()
	require.NoError(t, err)
	assert.FileExists(t, path)

	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, cfg.Provider, loaded.Provider)
	assert.Equal(t, cfg.Model, loaded.Model)
}

func TestLoadMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// No config written; Load should fail
	_, err := Load()
	assert.Error(t, err)
}

func TestDetectAnthropic(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-test")
	t.Setenv("OPENAI_API_KEY", "")
	provider, ok := DetectProvider()
	assert.True(t, ok)
	assert.Equal(t, ProviderAnthropic, provider)
}

func TestDetectOpenAI(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "sk-test")
	provider, ok := DetectProvider()
	assert.True(t, ok)
	assert.Equal(t, ProviderOpenAI, provider)
}

func TestDetectOllama(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")

	// Start a mock Ollama server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	original := ollamaHealthURL
	ollamaHealthURL = srv.URL
	defer func() { ollamaHealthURL = original }()

	provider, ok := DetectProvider()
	assert.True(t, ok)
	assert.Equal(t, ProviderOllama, provider)
}

func TestDetectNone(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")

	// Point to a port that's definitely not listening
	original := ollamaHealthURL
	ollamaHealthURL = "http://localhost:19999/api/tags"
	defer func() { ollamaHealthURL = original }()

	_, ok := DetectProvider()
	assert.False(t, ok)
}
