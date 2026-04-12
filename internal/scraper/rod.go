package scraper

import "errors"

// RodScraper is Layer 2: uses a headless browser (go-rod) for JS-heavy sites.
// Full implementation is added in Phase 11.
type RodScraper struct {
	timeoutSeconds int
}

var _ Scraper = (*RodScraper)(nil)

// NewRodScraper creates a RodScraper with the given timeout in seconds.
func NewRodScraper(timeoutSeconds int) *RodScraper {
	return &RodScraper{timeoutSeconds: timeoutSeconds}
}

// Scrape launches a headless browser and renders the page before extracting content.
func (r *RodScraper) Scrape(url string) (string, error) {
	return "", errors.New("rod scraper: not yet implemented — coming in Phase 11")
}
