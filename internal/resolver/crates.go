package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type cratesResponse struct {
	Crate struct {
		Documentation string `json:"documentation"`
		Homepage      string `json:"homepage"`
		Repository    string `json:"repository"`
	} `json:"crate"`
}

// resolveCrates queries crates.io for the library's documentation URL.
func (r *Resolver) resolveCrates(ctx context.Context, library string) []string {
	apiURL := fmt.Sprintf("%s/api/v1/crates/%s", r.CratesBaseURL, library)

	req, err := newRequest(ctx, apiURL)
	if err != nil {
		return nil
	}

	resp, err := r.HTTPClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return nil
	}
	defer resp.Body.Close()

	var pkg cratesResponse
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return nil
	}

	var urls []string
	if pkg.Crate.Documentation != "" {
		urls = append(urls, pkg.Crate.Documentation)
	}
	if pkg.Crate.Homepage != "" {
		urls = append(urls, pkg.Crate.Homepage)
	}
	if pkg.Crate.Repository != "" {
		urls = append(urls, pkg.Crate.Repository)
	}
	return urls
}

// resolveGo constructs the pkg.go.dev URL for Go modules directly from the module path.
func (r *Resolver) resolveGo(library string) []string {
	return []string{fmt.Sprintf("https://pkg.go.dev/%s", library)}
}

// newRequest creates an HTTP GET request with the markdocs User-Agent header.
func newRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	return req, nil
}
