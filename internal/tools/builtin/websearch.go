package builtin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	searchMaxResults  = 8
	searchMaxBodyBytes = 128 * 1024 // 128 KB — enough to extract results
	searchURL         = "https://html.duckduckgo.com/html/"
)

// WebSearchTool runs a DuckDuckGo search and returns result titles, URLs, and
// snippets. No API key required.
type WebSearchTool struct{}

func NewWebSearchTool() *WebSearchTool { return &WebSearchTool{} }

func (w *WebSearchTool) Name() string { return "web_search" }
func (w *WebSearchTool) Description() string {
	return "Search the web using DuckDuckGo. Returns up to 8 results with titles, URLs, and snippets. No API key required."
}
func (w *WebSearchTool) Properties() map[string]any {
	return map[string]any{
		"query": map[string]any{
			"type":        "string",
			"description": "The search query.",
		},
	}
}

type searchInput struct {
	Query string `json:"query"`
}

type searchResult struct {
	Title   string
	URL     string
	Snippet string
}

func (w *WebSearchTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var in searchInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if in.Query == "" {
		return "", fmt.Errorf("query is required")
	}

	results, err := duckduckgoSearch(ctx, in.Query)
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "No results found.", nil
	}

	var sb strings.Builder
	for i, r := range results {
		fmt.Fprintf(&sb, "%d. %s\n   %s\n   %s\n\n", i+1, r.Title, r.URL, r.Snippet)
	}
	return strings.TrimRight(sb.String(), "\n"), nil
}

func duckduckgoSearch(ctx context.Context, query string) ([]searchResult, error) {
	form := url.Values{"q": {query}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, searchURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("web_search: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "superclaw/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("web_search: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("web_search: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, searchMaxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("web_search: read body: %w", err)
	}

	return parseSearchResults(string(data)), nil
}

// parseSearchResults extracts results from DuckDuckGo's HTML response.
// DuckDuckGo's HTML page uses class="result__a" for links and
// class="result__snippet" for snippets.
var (
	reTitleURL = regexp.MustCompile(`class="result__a"[^>]*href="([^"]+)"[^>]*>([^<]+)</a>`)
	reSnippet  = regexp.MustCompile(`class="result__snippet"[^>]*>([^<]+(?:<[^>]+>[^<]*</[^>]+>)*[^<]*)</[a-z]+>`)
	reTag      = regexp.MustCompile(`<[^>]+>`)
)

func parseSearchResults(html string) []searchResult {
	titleMatches := reTitleURL.FindAllStringSubmatch(html, searchMaxResults)
	snippetMatches := reSnippet.FindAllStringSubmatch(html, searchMaxResults)

	var results []searchResult
	for i, m := range titleMatches {
		if len(m) < 3 {
			continue
		}
		rawURL := m[1]
		title := strings.TrimSpace(m[2])

		// DuckDuckGo wraps URLs in a redirect; extract the actual URL.
		if strings.HasPrefix(rawURL, "//duckduckgo.com/l/") ||
			strings.Contains(rawURL, "uddg=") {
			if parsed, err := url.ParseQuery(strings.TrimPrefix(rawURL, "//duckduckgo.com/l/?")); err == nil {
				if uddg := parsed.Get("uddg"); uddg != "" {
					if decoded, err := url.QueryUnescape(uddg); err == nil {
						rawURL = decoded
					}
				}
			}
		}

		snippet := ""
		if i < len(snippetMatches) && len(snippetMatches[i]) >= 2 {
			snippet = strings.TrimSpace(reTag.ReplaceAllString(snippetMatches[i][1], ""))
		}

		results = append(results, searchResult{
			Title:   title,
			URL:     rawURL,
			Snippet: snippet,
		})
	}
	return results
}
