// Package search provides web search functionality for documentation URL discovery.
package search

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var searchClient = &http.Client{
	Timeout: 15 * time.Second,
}

// DocURLs searches DuckDuckGo for documentation URLs for the given library.
// It returns up to maxResults candidate URLs filtered to likely doc sites.
// Falls back gracefully — returns (nil, nil) if search yields nothing usable.
func DocURLs(library string, maxResults int) ([]string, error) {
	query := fmt.Sprintf("%s official documentation", library)
	raw, err := ddgSearch(query)
	if err != nil {
		return nil, err
	}

	// Score and filter results
	var urls []string
	seen := make(map[string]bool)
	for _, u := range raw {
		if seen[u] {
			continue
		}
		seen[u] = true
		if isLikelyDocURL(u) {
			urls = append(urls, u)
		}
		if len(urls) >= maxResults {
			break
		}
	}

	// If strict filtering left nothing, fall back to all results
	if len(urls) == 0 {
		for _, u := range raw {
			if len(urls) >= maxResults {
				break
			}
			urls = append(urls, u)
		}
	}

	return urls, nil
}

// ValidateURLs returns only the URLs from candidates that respond with a 2xx
// status code, preserving order. Unreachable URLs are silently dropped.
func ValidateURLs(candidates []string) []string {
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	var valid []string
	for _, u := range candidates {
		resp, err := client.Head(u)
		if err != nil {
			// Try GET as fallback — some servers reject HEAD
			resp, err = client.Get(u)
			if err != nil {
				continue
			}
			resp.Body.Close()
		} else {
			resp.Body.Close()
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			valid = append(valid, u)
		}
	}
	return valid
}

// ddgSearch performs a DuckDuckGo HTML search and returns result URLs.
func ddgSearch(query string) ([]string, error) {
	endpoint := "https://html.duckduckgo.com/html/"
	form := url.Values{}
	form.Set("q", query)
	form.Set("kl", "us-en")

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("building ddg request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "markdocs/1.0 (documentation compiler)")

	resp, err := searchClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ddg search: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading ddg response: %w", err)
	}

	return extractDDGLinks(string(body)), nil
}

// extractDDGLinks parses DuckDuckGo HTML search results and returns result URLs.
func extractDDGLinks(body string) []string {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil
	}

	var urls []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href := attr.Val
					href = unwrapDDGRedirect(href)
					if href != "" && strings.HasPrefix(href, "http") {
						urls = append(urls, href)
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return urls
}

// unwrapDDGRedirect extracts the real destination from a DDG redirect URL.
// DDG wraps result links as //duckduckgo.com/l/?uddg=<encoded-url>&...
func unwrapDDGRedirect(href string) string {
	if !strings.Contains(href, "duckduckgo.com/l/") && !strings.Contains(href, "uddg=") {
		return href
	}
	parsed, err := url.Parse(href)
	if err != nil {
		return ""
	}
	if uddg := parsed.Query().Get("uddg"); uddg != "" {
		decoded, err := url.QueryUnescape(uddg)
		if err == nil {
			return decoded
		}
	}
	return ""
}

// docHostPatterns are hostname substrings that strongly indicate a documentation site.
var docHostPatterns = []string{
	"docs.", "doc.", "developer.", "developers.", "dev.",
	"api.", "reference.", "learn.", "guide.", "wiki.",
	"readthedocs.", ".github.io", "pkg.go.dev",
	"npmjs.com", "crates.io", "pypi.org",
}

// docPathPatterns are URL path substrings that indicate documentation.
var docPathPatterns = []string{
	"/docs", "/documentation", "/doc/", "/api/", "/reference/",
	"/guide/", "/getting-started", "/tutorial", "/manual",
}

// blocklistHosts are hosts that are almost never the right documentation source.
var blocklistHosts = []string{
	"youtube.com", "twitter.com", "x.com", "reddit.com",
	"stackoverflow.com", "medium.com", "dev.to", "linkedin.com",
	"facebook.com", "instagram.com", "amazon.com",
}

func isLikelyDocURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Host)
	path := strings.ToLower(parsed.Path)

	for _, blocked := range blocklistHosts {
		if strings.Contains(host, blocked) {
			return false
		}
	}
	for _, pat := range docHostPatterns {
		if strings.Contains(host, pat) {
			return true
		}
	}
	for _, pat := range docPathPatterns {
		if strings.Contains(path, pat) {
			return true
		}
	}
	return false
}
