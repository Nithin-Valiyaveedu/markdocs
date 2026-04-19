package skill

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// SkillMeta holds the YAML frontmatter written to each skill file.
// Claude Code fields (description, when_to_use, user-invocable) come first
// so Claude Code reads them immediately. markdocs tracking fields follow.
type SkillMeta struct {
	// Claude Code fields
	Name          string `yaml:"name"`
	Description   string `yaml:"description,omitempty"`
	WhenToUse     string `yaml:"when_to_use,omitempty"`
	UserInvocable bool   `yaml:"user-invocable"`
	// markdocs tracking fields
	Category         string   `yaml:"category"`
	Sources          []string `yaml:"sources"`
	Compiled         string   `yaml:"compiled"` // RFC3339
	Checksum         string   `yaml:"checksum"` // "sha256:<hex>"
	Model            string   `yaml:"model"`
	Provider         string   `yaml:"provider"`
	ProjectFramework string   `yaml:"project_framework,omitempty"`
	MarkdocsVersion  string   `yaml:"markdocs_version"`
}

// Writer writes compiled skill files to disk.
type Writer interface {
	// Write creates the skill file and returns its absolute path.
	Write(library, category, content string, meta SkillMeta) (string, error)
}

// FSWriter writes skill files to the filesystem under projectRoot.
type FSWriter struct {
	projectRoot string
}

var _ Writer = (*FSWriter)(nil)

// NewFSWriter creates an FSWriter that writes skills relative to projectRoot.
func NewFSWriter(projectRoot string) *FSWriter {
	return &FSWriter{projectRoot: projectRoot}
}

// Write creates .claude/skills/<category>/<library>.md under the project root.
func (w *FSWriter) Write(library, category, content string, meta SkillMeta) (string, error) {
	safeLib := sanitizeName(library)
	safeCat := sanitizeName(category)

	dir := filepath.Join(w.projectRoot, ".claude", "skills", safeCat)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating skill directory %s: %w", dir, err)
	}

	path := filepath.Join(dir, safeLib+".md")

	frontmatter, err := yaml.Marshal(meta)
	if err != nil {
		return "", fmt.Errorf("marshaling frontmatter: %w", err)
	}

	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(frontmatter)
	sb.WriteString("---\n\n")
	sb.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		sb.WriteString("\n")
	}

	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		return "", fmt.Errorf("writing skill file %s: %w", path, err)
	}

	return path, nil
}

// ContentChecksum returns a "sha256:<hex>" checksum of content.
func ContentChecksum(content string) string {
	sum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("sha256:%x", sum)
}

// NewSkillMeta builds a SkillMeta from the given inputs.
func NewSkillMeta(library, provider, model, category, framework, description, whenToUse string, sources []string, scrapedContent string) SkillMeta {
	return SkillMeta{
		Name:             sanitizeName(library),
		Description:      description,
		WhenToUse:        whenToUse,
		UserInvocable:    false,
		Category:         category,
		Sources:          sources,
		Compiled:         time.Now().UTC().Format(time.RFC3339),
		Checksum:         ContentChecksum(scrapedContent),
		Model:            model,
		Provider:         provider,
		ProjectFramework: framework,
		MarkdocsVersion:  "0.3.0",
	}
}

// sanitizeName makes a library or category name safe for use as a filename.
func sanitizeName(name string) string {
	r := strings.NewReplacer("/", "-", " ", "-", "\\", "-", "@", "")
	result := strings.ToLower(r.Replace(name))
	return strings.Trim(result, "-")
}
