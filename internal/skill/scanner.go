package skill

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectDeps holds discovered dependencies from a single dependency file.
type ProjectDeps struct {
	Libraries []string
	Source    string // e.g. "package.json", "go.mod"
}

// ScanProject returns all libraries found in the project at root by reading
// known dependency files.
func ScanProject(root string) ([]ProjectDeps, error) {
	var results []ProjectDeps

	// package.json
	pkgPath := filepath.Join(root, "package.json")
	if libs, err := ScanPackageJSON(pkgPath); err == nil && len(libs) > 0 {
		results = append(results, ProjectDeps{Libraries: libs, Source: "package.json"})
	}

	// go.mod
	goModPath := filepath.Join(root, "go.mod")
	if libs, err := ScanGoMod(goModPath); err == nil && len(libs) > 0 {
		results = append(results, ProjectDeps{Libraries: libs, Source: "go.mod"})
	}

	// requirements.txt
	reqPath := filepath.Join(root, "requirements.txt")
	if libs, err := ScanRequirementsTxt(reqPath); err == nil && len(libs) > 0 {
		results = append(results, ProjectDeps{Libraries: libs, Source: "requirements.txt"})
	}

	return results, nil
}

// ScanPackageJSON parses a package.json file and returns all dependency names
// from both "dependencies" and "devDependencies".
func ScanPackageJSON(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	seen := make(map[string]bool)
	var libs []string
	for name := range pkg.Dependencies {
		if !seen[name] {
			seen[name] = true
			libs = append(libs, name)
		}
	}
	for name := range pkg.DevDependencies {
		if !seen[name] {
			seen[name] = true
			libs = append(libs, name)
		}
	}
	return libs, nil
}

// ScanGoMod parses a go.mod file and returns the short names of required modules
// (the last path segment of each module path, e.g. "gin" from "github.com/gin-gonic/gin").
func ScanGoMod(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var libs []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "require") && !strings.Contains(line, "github.com") && !strings.Contains(line, "golang.org") {
			continue
		}
		// Handle single-line require: require github.com/foo/bar v1.0.0
		// Handle block lines:         github.com/foo/bar v1.0.0
		parts := strings.Fields(line)
		for _, p := range parts {
			if strings.Contains(p, ".") && strings.Contains(p, "/") && !strings.HasPrefix(p, "v") {
				// Looks like a module path
				segments := strings.Split(p, "/")
				name := segments[len(segments)-1]
				// Strip version suffix if any (e.g. /v2)
				if strings.HasPrefix(name, "v") && len(name) <= 3 {
					if len(segments) >= 2 {
						name = segments[len(segments)-2]
					}
				}
				if name != "" && name != "require" {
					libs = append(libs, name)
				}
			}
		}
	}
	return libs, nil
}

// ScanRequirementsTxt parses a requirements.txt file and returns package names.
func ScanRequirementsTxt(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var libs []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip version specifiers: requests>=2.0, flask==1.0, etc.
		for _, sep := range []string{">=", "<=", "==", "!=", "~=", ">"} {
			if idx := strings.Index(line, sep); idx != -1 {
				line = line[:idx]
				break
			}
		}
		line = strings.TrimSpace(line)
		if line != "" {
			libs = append(libs, line)
		}
	}
	return libs, nil
}

// SkillExists reports whether a skill file for library already exists under projectRoot.
func SkillExists(projectRoot, library string) bool {
	safeLib := strings.ToLower(strings.ReplaceAll(library, "/", "-"))
	// Search across all category subdirectories
	skillsDir := filepath.Join(projectRoot, ".claude", "skills")
	found := false
	_ = filepath.Walk(skillsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.TrimSuffix(info.Name(), ".md") == safeLib {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// DetectFramework returns a short string describing the project's primary framework,
// detected from package.json, go.mod, etc. Returns "" if not detected.
func DetectFramework(projectRoot string) string {
	// Check package.json for known frameworks
	pkgPath := filepath.Join(projectRoot, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		var pkg struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if json.Unmarshal(data, &pkg) == nil {
			for name := range pkg.Dependencies {
				if fw := jsFramework(name); fw != "" {
					return fw
				}
			}
			for name := range pkg.DevDependencies {
				if fw := jsFramework(name); fw != "" {
					return fw
				}
			}
		}
	}

	// Check go.mod for known Go frameworks
	goModPath := filepath.Join(projectRoot, "go.mod")
	if data, err := os.ReadFile(goModPath); err == nil {
		content := string(data)
		goFrameworks := map[string]string{
			"github.com/gin-gonic/gin":    "gin",
			"github.com/gofiber/fiber":    "fiber",
			"github.com/labstack/echo":    "echo",
			"github.com/go-chi/chi":       "chi",
			"github.com/gorilla/mux":      "gorilla-mux",
		}
		for module, fw := range goFrameworks {
			if strings.Contains(content, module) {
				return fw
			}
		}
		if strings.Contains(content, "go ") {
			return "go"
		}
	}

	return ""
}

func jsFramework(name string) string {
	frameworks := map[string]string{
		"next":           "next.js",
		"nuxt":           "nuxt",
		"react":          "react",
		"vue":            "vue",
		"@angular/core":  "angular",
		"svelte":         "svelte",
		"remix":          "remix",
		"astro":          "astro",
		"express":        "express",
		"fastify":        "fastify",
		"koa":            "koa",
	}
	if fw, ok := frameworks[name]; ok {
		return fw
	}
	return ""
}
