package scraper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleArticleHTML contains enough content to exceed MinContentLength.
const sampleArticleHTML = `<!DOCTYPE html>
<html>
<head><title>Test Article</title></head>
<body>
<article>
<h1>Getting Started with React</h1>
<p>React is a JavaScript library for building user interfaces. It was created by Facebook and has become one of the most popular frontend frameworks.</p>
<p>React uses a component-based architecture where you build encapsulated components that manage their own state, then compose them to make complex UIs.</p>
<h2>Installation</h2>
<p>You can install React using npm or yarn. Run the following command in your terminal to create a new React application with all the necessary tooling pre-configured for you.</p>
<pre><code>npx create-react-app my-app</code></pre>
<h2>Key Concepts</h2>
<p>JSX is a syntax extension for JavaScript that allows you to write HTML-like code inside your JavaScript files. Components can be either functional or class-based, though functional components with hooks are now the recommended approach.</p>
<p>State management in React is handled through the useState hook for local state, and libraries like Redux or Zustand for global application state that needs to be shared across many components.</p>
</article>
</body>
</html>`

func TestHTTPScraperSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, sampleArticleHTML)
	}))
	defer srv.Close()

	s := newHTTPScraper()
	content, err := s.Scrape(srv.URL)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(content), MinContentLength)
	assert.Contains(t, content, "React")
}

func TestHTTPScraperShortContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "<html><body><p>Short.</p></body></html>")
	}))
	defer srv.Close()

	s := newHTTPScraper()
	_, err := s.Scrape(srv.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too short")
}

func TestHTTPScraperBadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	s := newHTTPScraper()
	_, err := s.Scrape(srv.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestToMarkdown(t *testing.T) {
	html := `<h1>Hello</h1><p>This is a <strong>test</strong> paragraph.</p>`
	result, err := ToMarkdown(html)
	require.NoError(t, err)
	assert.Contains(t, result, "Hello")
	assert.Contains(t, result, "test")
}

func TestStripNoise(t *testing.T) {
	input := `# Documentation
Skip to main content
## Introduction
This is the real content.
Cookie preferences
## Usage
Use it like this.
© 2024 Company. All rights reserved.`

	result := StripNoise(input)
	assert.Contains(t, result, "# Documentation")
	assert.Contains(t, result, "This is the real content")
	assert.NotContains(t, result, "Skip to main content")
	assert.NotContains(t, result, "Cookie preferences")
	assert.NotContains(t, result, "All rights reserved")
}

// MockScraper is a test double for Scraper.
type MockScraper struct {
	content string
	err     error
	called  bool
}

var _ Scraper = (*MockScraper)(nil)

func (m *MockScraper) Scrape(url string) (string, error) {
	m.called = true
	return m.content, m.err
}

func TestWaterfallFallsToLayer2(t *testing.T) {
	// Layer 1 returns content that's too short
	layer1 := &MockScraper{content: "short"}
	// Layer 2 returns adequate content
	longContent := strings.Repeat("x", MinContentLength+1)
	layer2 := &MockScraper{content: longContent}

	w := &WaterfallScraper{layers: []Scraper{layer1, layer2}}
	content, err := w.Scrape("http://example.com")
	require.NoError(t, err)
	assert.True(t, layer1.called, "layer1 should have been called")
	assert.True(t, layer2.called, "layer2 should have been called")
	assert.Equal(t, longContent, content)
}

func TestWaterfallSucceedsOnLayer1(t *testing.T) {
	longContent := strings.Repeat("x", MinContentLength+1)
	layer1 := &MockScraper{content: longContent}
	layer2 := &MockScraper{content: "not reached"}

	w := &WaterfallScraper{layers: []Scraper{layer1, layer2}}
	content, err := w.Scrape("http://example.com")
	require.NoError(t, err)
	assert.False(t, layer2.called, "layer2 should NOT have been called")
	assert.Equal(t, longContent, content)
}
