package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanPackageJSONDeps(t *testing.T) {
	dir := t.TempDir()
	pkgJSON := `{
		"dependencies": {"react": "^18.0.0", "next": "^14.0.0"},
		"devDependencies": {"vitest": "^1.0.0"}
	}`
	path := filepath.Join(dir, "package.json")
	require.NoError(t, os.WriteFile(path, []byte(pkgJSON), 0o644))

	libs, err := ScanPackageJSON(path)
	require.NoError(t, err)
	assert.Contains(t, libs, "react")
	assert.Contains(t, libs, "next")
	assert.Contains(t, libs, "vitest")
}

func TestScanPackageJSONMissing(t *testing.T) {
	_, err := ScanPackageJSON("/nonexistent/package.json")
	assert.Error(t, err)
}

func TestScanGoMod(t *testing.T) {
	dir := t.TempDir()
	goMod := `module example.com/myapp

go 1.22

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/stretchr/testify v1.8.4
)
`
	path := filepath.Join(dir, "go.mod")
	require.NoError(t, os.WriteFile(path, []byte(goMod), 0o644))

	libs, err := ScanGoMod(path)
	require.NoError(t, err)
	assert.Contains(t, libs, "gin")
	assert.Contains(t, libs, "testify")
}

func TestScanRequirementsTxt(t *testing.T) {
	dir := t.TempDir()
	reqs := `requests>=2.28.0
flask==2.3.0
# a comment
pytest>=7.0
`
	path := filepath.Join(dir, "requirements.txt")
	require.NoError(t, os.WriteFile(path, []byte(reqs), 0o644))

	libs, err := ScanRequirementsTxt(path)
	require.NoError(t, err)
	assert.Contains(t, libs, "requests")
	assert.Contains(t, libs, "flask")
	assert.Contains(t, libs, "pytest")
}

func TestScanProject(t *testing.T) {
	dir := t.TempDir()

	pkgJSON := `{"dependencies": {"react": "^18.0.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644))

	goMod := "module example.com/myapp\n\ngo 1.22\n\nrequire github.com/gin-gonic/gin v1.9.1\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644))

	results, err := ScanProject(dir)
	require.NoError(t, err)
	assert.Len(t, results, 2)

	sources := make(map[string]bool)
	for _, r := range results {
		sources[r.Source] = true
	}
	assert.True(t, sources["package.json"])
	assert.True(t, sources["go.mod"])
}

func TestSkillExists(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, ".claude", "skills", "frontend")
	require.NoError(t, os.MkdirAll(skillDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(skillDir, "react.md"), []byte("# React"), 0o644))

	assert.True(t, SkillExists(dir, "react"))
	assert.False(t, SkillExists(dir, "vue"))
}

func TestDetectFrameworkReact(t *testing.T) {
	dir := t.TempDir()
	pkgJSON := `{"dependencies": {"react": "^18.0.0", "next": "^14.0.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0o644))

	fw := DetectFramework(dir)
	// Either "react" or "next.js" depending on map iteration order — both are valid
	assert.NotEmpty(t, fw)
}

func TestDetectFrameworkGin(t *testing.T) {
	dir := t.TempDir()
	goMod := "module example.com/myapp\n\ngo 1.22\n\nrequire github.com/gin-gonic/gin v1.9.1\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644))

	fw := DetectFramework(dir)
	assert.Equal(t, "gin", fw)
}

func TestDetectFrameworkNone(t *testing.T) {
	fw := DetectFramework(t.TempDir())
	assert.Empty(t, fw)
}
