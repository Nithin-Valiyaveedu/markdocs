// Package resolver discovers authoritative documentation URLs for libraries by
// querying package registries (npm, PyPI, crates.io, pkg.go.dev) before
// falling back to web search. This gives maintainer-declared URLs rather than
// whatever a search engine surfaces.
package resolver

import (
	"context"
	"net/http"
	"strings"
	"time"
)

const userAgent = "markdocs/0.3.0 (https://github.com/Nithin-Valiyaveedu/markdocs)"

// Resolver queries package registries to find official documentation URLs.
// Base URLs are configurable so tests can inject httptest servers.
type Resolver struct {
	NPMBaseURL    string
	PyPIBaseURL   string
	CratesBaseURL string
	HTTPClient    *http.Client
}

// New returns a Resolver configured for the real public registries.
func New() *Resolver {
	return &Resolver{
		NPMBaseURL:    "https://registry.npmjs.org",
		PyPIBaseURL:   "https://pypi.org",
		CratesBaseURL: "https://crates.io",
		HTTPClient:    &http.Client{Timeout: 10 * time.Second},
	}
}

// Resolve returns candidate documentation URLs for the given library, ordered
// from most to least authoritative. It never returns an error unless all
// registry lookups fail — in that case callers should fall back to web search.
func (r *Resolver) Resolve(ctx context.Context, library string) ([]string, error) {
	var candidates []string

	switch {
	case isGoModule(library):
		candidates = append(candidates, r.resolveGo(library)...)
	case strings.HasPrefix(library, "@"):
		// Scoped npm package — only try npm
		candidates = append(candidates, r.resolveNPM(ctx, library)...)
	default:
		// Unknown ecosystem: try npm first (largest registry), then PyPI, then crates
		candidates = append(candidates, r.resolveNPM(ctx, library)...)
		candidates = append(candidates, r.resolvePyPI(ctx, library)...)
		candidates = append(candidates, r.resolveCrates(ctx, library)...)
	}

	return dedupe(candidates), nil
}

// isGoModule reports whether the library name looks like a Go module path.
func isGoModule(library string) bool {
	goPrefixes := []string{
		"github.com/", "golang.org/", "gopkg.in/",
		"k8s.io/", "sigs.k8s.io/", "go.uber.org/",
		"go.opentelemetry.io/", "cloud.google.com/",
	}
	for _, p := range goPrefixes {
		if strings.HasPrefix(library, p) {
			return true
		}
	}
	return false
}

// isRepoURL reports whether a URL points to a source code host rather than docs.
// We still include these, but docs-like URLs are placed first.
func isRepoURL(u string) bool {
	repoHosts := []string{"github.com", "gitlab.com", "bitbucket.org", "sr.ht"}
	for _, h := range repoHosts {
		if strings.Contains(u, h) {
			return true
		}
	}
	return false
}

// dedupe removes duplicate and empty URLs, preserving order.
// Non-repo (docs) URLs are returned before repo URLs.
func dedupe(urls []string) []string {
	seen := make(map[string]bool)
	var docs, repos []string
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		if isRepoURL(u) {
			repos = append(repos, u)
		} else {
			docs = append(docs, u)
		}
	}
	return append(docs, repos...)
}
