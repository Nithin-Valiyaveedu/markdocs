package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleSkillFile = `---
name: react
category: frontend
sources:
    - https://react.dev
compiled: "2026-04-12T10:00:00Z"
checksum: sha256:abc123
model: llama3.2
provider: ollama
project_framework: next.js
markdocs_version: 0.2.1
---

# react — markdocs skill

## What This Is
React is a JavaScript library.
`

func TestReadFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "react.md")
	require.NoError(t, os.WriteFile(path, []byte(sampleSkillFile), 0o644))

	meta, err := ReadFrontmatter(path)
	require.NoError(t, err)
	assert.Equal(t, "react", meta.Name)
	assert.Equal(t, "frontend", meta.Category)
	assert.Equal(t, "llama3.2", meta.Model)
	assert.Equal(t, "ollama", meta.Provider)
	assert.Equal(t, "sha256:abc123", meta.Checksum)
	assert.Equal(t, "next.js", meta.ProjectFramework)
	assert.Len(t, meta.Sources, 1)
	assert.Equal(t, "https://react.dev", meta.Sources[0])
}

func TestReadFrontmatterMissingFile(t *testing.T) {
	_, err := ReadFrontmatter("/nonexistent/file.md")
	assert.Error(t, err)
}

func TestReadFrontmatterNoDelimiter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.md")
	require.NoError(t, os.WriteFile(path, []byte("# No frontmatter here\n"), 0o644))

	_, err := ReadFrontmatter(path)
	assert.Error(t, err)
}

func TestGlobSkills(t *testing.T) {
	dir := t.TempDir()

	// Write two skill files in different categories
	react := filepath.Join(dir, ".claude", "skills", "frontend")
	stripe := filepath.Join(dir, ".claude", "skills", "payments")
	require.NoError(t, os.MkdirAll(react, 0o755))
	require.NoError(t, os.MkdirAll(stripe, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(react, "react.md"), []byte(sampleSkillFile), 0o644))

	stripeFile := `---
name: stripe
category: payments
sources:
    - https://stripe.com/docs
compiled: "2026-04-12T10:00:00Z"
checksum: sha256:def456
model: gpt-4o-mini
provider: openai
markdocs_version: 0.2.1
---

# stripe — markdocs skill
`
	require.NoError(t, os.WriteFile(filepath.Join(stripe, "stripe.md"), []byte(stripeFile), 0o644))

	files, err := GlobSkills(dir)
	require.NoError(t, err)
	assert.Len(t, files, 2)

	names := make(map[string]bool)
	for _, f := range files {
		names[f.Meta.Name] = true
	}
	assert.True(t, names["react"])
	assert.True(t, names["stripe"])
}

func TestCompiledAge(t *testing.T) {
	// A timestamp from the distant past should return "Nd ago"
	age := CompiledAge("2020-01-01T00:00:00Z")
	assert.Contains(t, age, "d ago")

	// A recent timestamp should return minutes or hours
	import_now := "2026-04-12T10:00:00Z"
	age = CompiledAge(import_now)
	assert.NotEmpty(t, age)

	// Invalid timestamp
	age = CompiledAge("not-a-date")
	assert.Equal(t, "unknown", age)
}
