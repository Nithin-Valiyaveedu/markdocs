package resolver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeNPM returns a test server that responds to registry.npmjs.org-style requests.
func fakeNPM(t *testing.T, homepage, repoURL string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"homepage": homepage,
			"repository": map[string]string{
				"url": repoURL,
			},
		})
	}))
}

func fakePyPI(t *testing.T, docURL, homepage string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"info": map[string]any{
				"project_urls": map[string]string{
					"Documentation": docURL,
				},
				"home_page": homepage,
			},
		})
	}))
}

func fakeCrates(t *testing.T, docURL, homepage, repo string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"crate": map[string]string{
				"documentation": docURL,
				"homepage":      homepage,
				"repository":    repo,
			},
		})
	}))
}

func fakeNotFound() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
}

func TestResolveNPM_WithDocsHomepage(t *testing.T) {
	srv := fakeNPM(t, "https://zustand.docs.pmnd.rs/", "git+https://github.com/pmndrs/zustand.git")
	defer srv.Close()

	r := &Resolver{NPMBaseURL: srv.URL, PyPIBaseURL: "http://unused", CratesBaseURL: "http://unused", HTTPClient: &http.Client{}}
	urls, err := r.Resolve(context.Background(), "zustand")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(urls) == 0 {
		t.Fatal("expected at least one URL")
	}
	// docs URL should come first (non-GitHub)
	if urls[0] != "https://zustand.docs.pmnd.rs/" {
		t.Errorf("expected docs URL first, got %q", urls[0])
	}
	// GitHub repo should still be present but after docs
	found := false
	for _, u := range urls {
		if strings.Contains(u, "github.com") {
			found = true
		}
	}
	if !found {
		t.Error("expected GitHub repo URL in results")
	}
}

func TestResolveNPM_GitHubHomepageOnly(t *testing.T) {
	// When homepage IS the GitHub repo, it should still be returned (many libs only have GH docs)
	srv := fakeNPM(t, "https://github.com/lodash/lodash", "git+https://github.com/lodash/lodash.git")
	defer srv.Close()

	r := &Resolver{NPMBaseURL: srv.URL, PyPIBaseURL: "http://unused", CratesBaseURL: "http://unused", HTTPClient: &http.Client{}}
	urls, err := r.Resolve(context.Background(), "lodash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should have the GitHub URL (deduplicated — homepage and repo are the same)
	if len(urls) == 0 {
		t.Fatal("expected at least one URL even when homepage is GitHub")
	}
	for _, u := range urls {
		if !strings.Contains(u, "github.com") {
			t.Errorf("unexpected non-GitHub URL %q when only GitHub is available", u)
		}
	}
}

func TestResolveNPM_NotFound(t *testing.T) {
	srv := fakeNotFound()
	defer srv.Close()

	r := &Resolver{NPMBaseURL: srv.URL, PyPIBaseURL: srv.URL, CratesBaseURL: srv.URL, HTTPClient: &http.Client{}}
	urls, err := r.Resolve(context.Background(), "nonexistent-package-xyz")
	if err != nil {
		t.Fatalf("Resolve should not return error on 404, got: %v", err)
	}
	// Empty is fine — caller falls back to web search
	if urls == nil {
		urls = []string{}
	}
	_ = urls
}

func TestResolvePyPI(t *testing.T) {
	srv := fakePyPI(t, "https://docs.python-requests.org/", "https://requests.readthedocs.io/")
	defer srv.Close()

	r := &Resolver{NPMBaseURL: "http://unused", PyPIBaseURL: srv.URL, CratesBaseURL: "http://unused", HTTPClient: &http.Client{}}
	// PyPI resolution is triggered for simple names (non-Go, non-scoped)
	urls := r.resolvePyPI(context.Background(), "requests")
	if len(urls) == 0 {
		t.Fatal("expected URLs from PyPI")
	}
	if urls[0] != "https://docs.python-requests.org/" {
		t.Errorf("expected Documentation URL first, got %q", urls[0])
	}
}

func TestResolveCrates(t *testing.T) {
	srv := fakeCrates(t, "https://docs.rs/tokio", "https://tokio.rs", "https://github.com/tokio-rs/tokio")
	defer srv.Close()

	r := &Resolver{NPMBaseURL: "http://unused", PyPIBaseURL: "http://unused", CratesBaseURL: srv.URL, HTTPClient: &http.Client{}}
	urls := r.resolveCrates(context.Background(), "tokio")
	if len(urls) == 0 {
		t.Fatal("expected URLs from crates.io")
	}
	if urls[0] != "https://docs.rs/tokio" {
		t.Errorf("expected documentation URL first, got %q", urls[0])
	}
}

func TestResolveGoModule(t *testing.T) {
	r := New()
	urls, err := r.Resolve(context.Background(), "github.com/gin-gonic/gin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(urls) == 0 {
		t.Fatal("expected pkg.go.dev URL")
	}
	if urls[0] != "https://pkg.go.dev/github.com/gin-gonic/gin" {
		t.Errorf("expected pkg.go.dev URL, got %q", urls[0])
	}
}

func TestDeduplication(t *testing.T) {
	// Same URL from homepage and repo should appear only once
	srv := fakeNPM(t, "https://github.com/foo/bar", "git+https://github.com/foo/bar.git")
	defer srv.Close()

	r := &Resolver{NPMBaseURL: srv.URL, PyPIBaseURL: "http://unused", CratesBaseURL: "http://unused", HTTPClient: &http.Client{}}
	urls, _ := r.Resolve(context.Background(), "foo")
	seen := make(map[string]int)
	for _, u := range urls {
		seen[u]++
		if seen[u] > 1 {
			t.Errorf("duplicate URL returned: %q", u)
		}
	}
}

func TestCleanGitURL(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"git+https://github.com/foo/bar.git", "https://github.com/foo/bar"},
		{"https://github.com/foo/bar", "https://github.com/foo/bar"},
		{"git://github.com/foo/bar.git", "https://github.com/foo/bar"},
		{"", ""},
		{"not-a-url", ""},
	}
	for _, tc := range cases {
		got := cleanGitURL(tc.input)
		if got != tc.want {
			t.Errorf("cleanGitURL(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestIsGoModule(t *testing.T) {
	goModules := []string{
		"github.com/gin-gonic/gin",
		"golang.org/x/net",
		"gopkg.in/yaml.v3",
		"k8s.io/client-go",
		"go.uber.org/zap",
	}
	notGoModules := []string{
		"react", "zustand", "@scope/pkg", "requests", "tokio",
	}
	for _, m := range goModules {
		if !isGoModule(m) {
			t.Errorf("isGoModule(%q) = false, want true", m)
		}
	}
	for _, m := range notGoModules {
		if isGoModule(m) {
			t.Errorf("isGoModule(%q) = true, want false", m)
		}
	}
}
