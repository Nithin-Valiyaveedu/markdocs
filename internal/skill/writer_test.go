package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestFSWriterCreatesFile(t *testing.T) {
	dir := t.TempDir()
	w := NewFSWriter(dir)

	meta := SkillMeta{
		Name:     "react",
		Category: "frontend",
		Sources:  []string{"https://react.dev"},
		Compiled: "2026-04-12T10:00:00Z",
		Checksum: "sha256:abc123",
		Model:    "test",
		Provider: "ollama",
	}
	path, err := w.Write("react", "frontend", "# React skill\n## What This Is\nA library.", meta)
	require.NoError(t, err)

	assert.FileExists(t, path)
	assert.Equal(t, filepath.Join(dir, ".claude", "skills", "frontend", "react.md"), path)

	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "---")
	assert.Contains(t, string(content), "# React skill")
}

func TestFSWriterFrontmatterParseable(t *testing.T) {
	dir := t.TempDir()
	w := NewFSWriter(dir)

	scraped := "documentation content for testing"
	meta := SkillMeta{
		Name:             "react",
		Category:         "frontend",
		Sources:          []string{"https://react.dev", "https://react.dev/reference"},
		Compiled:         "2026-04-12T10:00:00Z",
		Checksum:         ContentChecksum(scraped),
		Model:            "llama3.2",
		Provider:         "ollama",
		ProjectFramework: "next.js",
		MarkdocsVersion:  "0.2.0",
	}
	path, err := w.Write("react", "frontend", "# React\n## What This Is\nContent.", meta)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)
	parts := strings.SplitN(content, "---\n", 3)
	require.Len(t, parts, 3, "expected two --- delimiters")

	var parsed SkillMeta
	require.NoError(t, yaml.Unmarshal([]byte(parts[1]), &parsed))
	assert.Equal(t, meta.Name, parsed.Name)
	assert.Equal(t, meta.Model, parsed.Model)
	assert.Equal(t, meta.Category, parsed.Category)
	assert.Equal(t, meta.Checksum, parsed.Checksum)
	assert.Len(t, parsed.Sources, 2)
}

func TestFSWriterSanitizesName(t *testing.T) {
	dir := t.TempDir()
	w := NewFSWriter(dir)

	meta := SkillMeta{Compiled: "2026-04-12T10:00:00Z", Model: "test", Category: "ui"}
	path, err := w.Write("@shadcn/ui", "UI Components", "# Shadcn", meta)
	require.NoError(t, err)

	assert.NotContains(t, filepath.Base(path), "@")
	assert.FileExists(t, path)
}

func TestContentChecksumFormat(t *testing.T) {
	sum := ContentChecksum("hello world")
	assert.True(t, strings.HasPrefix(sum, "sha256:"))
	assert.Len(t, sum, len("sha256:")+64)
}
