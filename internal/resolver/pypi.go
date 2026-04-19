package resolver

import (
	"context"
	"encoding/json"
	"fmt"
)

type pypiResponse struct {
	Info struct {
		HomePage    string            `json:"home_page"`
		ProjectURLs map[string]string `json:"project_urls"`
	} `json:"info"`
}

// docURLKeys are the project_urls keys we consider official documentation, in priority order.
var docURLKeys = []string{
	"Documentation", "Docs", "documentation", "docs",
	"Homepage", "Home", "homepage",
}

// resolveP yPI queries PyPI for the library's documentation URL.
func (r *Resolver) resolvePyPI(ctx context.Context, library string) []string {
	apiURL := fmt.Sprintf("%s/pypi/%s/json", r.PyPIBaseURL, library)

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

	var pkg pypiResponse
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return nil
	}

	var urls []string
	// Check project_urls in priority order
	for _, key := range docURLKeys {
		if u := pkg.Info.ProjectURLs[key]; u != "" {
			urls = append(urls, u)
			break
		}
	}
	if pkg.Info.HomePage != "" {
		urls = append(urls, pkg.Info.HomePage)
	}
	return urls
}
