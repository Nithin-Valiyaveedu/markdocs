package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// SkillFile represents a parsed skill file including its path and metadata.
type SkillFile struct {
	Path string
	Meta SkillMeta
}

// ReadFrontmatter parses the YAML frontmatter from a skill file at path.
func ReadFrontmatter(path string) (*SkillMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return parseFrontmatter(string(data))
}

// parseFrontmatter extracts and parses YAML between the first pair of --- delimiters.
func parseFrontmatter(content string) (*SkillMeta, error) {
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("no frontmatter found (file must start with ---)")
	}
	rest := content[4:] // skip first "---\n"
	end := strings.Index(rest, "\n---\n")
	if end == -1 {
		return nil, fmt.Errorf("frontmatter closing --- not found")
	}
	yamlBlock := rest[:end]

	var meta SkillMeta
	if err := yaml.Unmarshal([]byte(yamlBlock), &meta); err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}
	return &meta, nil
}

// GlobSkills returns all skill files under projectRoot (.claude/skills/**/*.md).
func GlobSkills(projectRoot string) ([]SkillFile, error) {
	skillsDir := filepath.Join(projectRoot, ".claude", "skills")
	pattern := filepath.Join(skillsDir, "**", "*.md")

	// filepath.Glob doesn't support **, so use Walk
	var files []SkillFile
	err := filepath.Walk(skillsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable dirs
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		meta, err := ReadFrontmatter(path)
		if err != nil {
			return nil // skip malformed files
		}
		files = append(files, SkillFile{Path: path, Meta: *meta})
		return nil
	})
	_ = pattern
	if err != nil {
		return nil, fmt.Errorf("scanning skills: %w", err)
	}
	return files, nil
}

// CompiledAge returns a human-readable string for how long ago the skill was compiled.
func CompiledAge(compiledStr string) string {
	t, err := time.Parse(time.RFC3339, compiledStr)
	if err != nil {
		return "unknown"
	}
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
