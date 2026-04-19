package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type npmPackage struct {
	Homepage   string `json:"homepage"`
	Repository struct {
		URL string `json:"url"`
	} `json:"repository"`
	Bugs struct {
		URL string `json:"url"`
	} `json:"bugs"`
}

// resolveNPM queries the npm registry for the library's homepage and repository URL.
func (r *Resolver) resolveNPM(ctx context.Context, library string) []string {
	// Scoped packages like @scope/name need the slash URL-encoded
	encodedName := strings.ReplaceAll(library, "/", "%2F")
	apiURL := fmt.Sprintf("%s/%s", r.NPMBaseURL, encodedName)

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

	var pkg npmPackage
	if err := json.NewDecoder(resp.Body).Decode(&pkg); err != nil {
		return nil
	}

	var urls []string
	if pkg.Homepage != "" {
		urls = append(urls, pkg.Homepage)
	}
	if repoURL := cleanGitURL(pkg.Repository.URL); repoURL != "" {
		urls = append(urls, repoURL)
	}
	return urls
}

// cleanGitURL converts a git remote URL to an HTTPS URL.
// e.g. "git+https://github.com/foo/bar.git" → "https://github.com/foo/bar"
func cleanGitURL(raw string) string {
	raw = strings.TrimPrefix(raw, "git+")
	raw = strings.TrimSuffix(raw, ".git")
	raw = strings.TrimPrefix(raw, "git://")
	if strings.HasPrefix(raw, "github.com/") {
		raw = "https://" + raw
	}
	// Validate it's an HTTP URL
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return ""
	}
	return raw
}
