package scraper

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// RodScraper is Layer 2: uses a headless browser (go-rod) for JS-heavy sites.
type RodScraper struct {
	timeoutSeconds int
}

var _ Scraper = (*RodScraper)(nil)

// NewRodScraper creates a RodScraper with the given timeout in seconds.
func NewRodScraper(timeoutSeconds int) *RodScraper {
	return &RodScraper{timeoutSeconds: timeoutSeconds}
}

// Scrape launches a headless browser, renders the page, and returns clean Markdown.
func (r *RodScraper) Scrape(url string) (string, error) {
	timeout := time.Duration(r.timeoutSeconds) * time.Second

	// Try to find an installed browser; if not available, return a clear error
	path, found := launcher.LookPath()
	if !found {
		return "", fmt.Errorf("rod scraper: no browser found — install Chromium or Chrome to scrape JS-heavy sites")
	}

	u := launcher.New().Bin(path).MustLaunch()

	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	page := browser.Timeout(timeout).MustPage(url)
	defer page.MustClose()

	if err := page.WaitLoad(); err != nil {
		return "", fmt.Errorf("rod waiting for page load %s: %w", url, err)
	}

	html, err := page.HTML()
	if err != nil {
		return "", fmt.Errorf("rod getting html from %s: %w", url, err)
	}

	markdown, err := ToMarkdown(html)
	if err != nil {
		return "", fmt.Errorf("converting rod html to markdown: %w", err)
	}

	clean := strings.TrimSpace(StripNoise(markdown))
	if len(clean) < MinContentLength {
		return "", fmt.Errorf("rod scraped content too short (%d chars) from %s", len(clean), url)
	}
	return clean, nil
}
