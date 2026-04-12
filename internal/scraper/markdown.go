package scraper

import (
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

// ToMarkdown converts an HTML string to clean Markdown.
func ToMarkdown(html string) (string, error) {
	converter := md.NewConverter("", true, nil)
	result, err := converter.ConvertString(html)
	if err != nil {
		return "", err
	}
	return result, nil
}

// StripNoise removes noisy lines from Markdown (navigation menus, footers,
// script content, sidebar elements).
func StripNoise(content string) string {
	lines := strings.Split(content, "\n")
	var filtered []string
	for _, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		if isNoisyLine(lower) {
			continue
		}
		filtered = append(filtered, line)
	}
	// Collapse multiple blank lines into one
	var result []string
	prevBlank := false
	for _, line := range filtered {
		isBlank := strings.TrimSpace(line) == ""
		if isBlank && prevBlank {
			continue
		}
		result = append(result, line)
		prevBlank = isBlank
	}
	return strings.Join(result, "\n")
}

func isNoisyLine(lower string) bool {
	noisePrefixes := []string{
		"skip to",
		"cookie",
		"accept all",
		"privacy policy",
	}
	noiseContains := []string{
		"©",
		"all rights reserved",
	}
	for _, p := range noisePrefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	for _, c := range noiseContains {
		if strings.Contains(lower, c) {
			return true
		}
	}
	return false
}
