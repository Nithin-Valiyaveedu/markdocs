package scraper

import (
	"fmt"
	"net/url"
	"strings"
)

// MinContentLength is the minimum number of characters required for scraped
// content to be considered successful. Below this threshold, the waterfall
// scraper tries the next layer.
const MinContentLength = 500

// Scraper fetches a URL and returns clean Markdown content.
type Scraper interface {
	Scrape(url string) (string, error)
}

// WaterfallScraper tries each layer in order, moving to the next when a layer
// returns content shorter than MinContentLength or an error.
type WaterfallScraper struct {
	layers []Scraper
}

var _ Scraper = (*WaterfallScraper)(nil)

// NewWaterfall returns a Scraper that tries HTTPScraper first, then RodScraper.
func NewWaterfall() Scraper {
	return &WaterfallScraper{
		layers: []Scraper{
			newHTTPScraper(),
			NewRodScraper(60),
		},
	}
}

// NewHTTPScraper returns only the Layer 1 HTTP scraper.
func NewHTTPScraper() Scraper {
	return newHTTPScraper()
}

// ValidateURL returns an error if rawURL is not a valid http/https URL.
func ValidateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL %q: %w", rawURL, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("invalid URL %q: must start with http:// or https://", rawURL)
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return fmt.Errorf("invalid URL %q: missing host", rawURL)
	}
	return nil
}

// Scrape attempts each layer in order and returns the first successful result.
func (w *WaterfallScraper) Scrape(url string) (string, error) {
	if err := ValidateURL(url); err != nil {
		return "", err
	}
	var lastErr error
	for _, layer := range w.layers {
		content, err := layer.Scrape(url)
		if err == nil && len(content) >= MinContentLength {
			return content, nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("content too short (%d chars)", len(content))
		}
	}
	if lastErr != nil {
		return "", lastErr
	}
	return "", fmt.Errorf("all scraper layers failed for %s", url)
}
