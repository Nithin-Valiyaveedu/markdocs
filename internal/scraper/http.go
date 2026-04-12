package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	readability "github.com/go-shiori/go-readability"
)

// HTTPScraper is Layer 1: fetches pages with net/http and extracts article
// content using go-readability.
type HTTPScraper struct {
	client  *http.Client
	timeout time.Duration
}

var _ Scraper = (*HTTPScraper)(nil)

// newHTTPScraper creates an HTTPScraper with a 30-second timeout.
func newHTTPScraper() *HTTPScraper {
	return newHTTPScraperWithTimeout(30 * time.Second)
}

// newHTTPScraperWithTimeout creates an HTTPScraper with the given timeout.
func newHTTPScraperWithTimeout(timeout time.Duration) *HTTPScraper {
	return &HTTPScraper{
		client:  &http.Client{Timeout: timeout},
		timeout: timeout,
	}
}

// Scrape fetches the URL, extracts the main article via go-readability,
// converts HTML to Markdown, and strips navigation/footer noise.
// Returns an error if the resulting content is shorter than MinContentLength.
func (s *HTTPScraper) Scrape(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parsing url %s: %w", rawURL, err)
	}

	resp, err := s.client.Get(rawURL)
	if err != nil {
		return "", fmt.Errorf("fetching %s: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("fetching %s: status %d", rawURL, resp.StatusCode)
	}

	article, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		return "", fmt.Errorf("extracting article from %s: %w", rawURL, err)
	}

	markdown, err := ToMarkdown(article.Content)
	if err != nil {
		return "", fmt.Errorf("converting html to markdown: %w", err)
	}

	clean := StripNoise(markdown)
	clean = strings.TrimSpace(clean)

	if len(clean) < MinContentLength {
		return "", fmt.Errorf("scraped content too short (%d chars) from %s", len(clean), rawURL)
	}

	return clean, nil
}
