package paperscool

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/kyc001/paper-radar/internal/model"
)

const defaultBaseURL = "https://papers.cool"

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 20 * time.Second},
		baseURL:    defaultBaseURL,
	}
}

func (c *Client) Fetch(ctx context.Context, topicQuery string, maxResults int, withKimi bool) ([]model.Paper, error) {
	feedURL := c.resolveFeedURL(topicQuery)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "paper-radar/0.2.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("unexpected status %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var feed atomFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, err
	}

	limit := maxResults
	if limit <= 0 || limit > len(feed.Entries) {
		limit = len(feed.Entries)
	}

	papers := make([]model.Paper, 0, limit)
	for i := 0; i < limit; i++ {
		entry := feed.Entries[i]
		paperID := extractPaperID(entry.ID)
		summary := normalizeWhitespace(entry.Summary)
		if withKimi && paperID != "" {
			if kimi, err := c.FetchKimiSummary(ctx, paperID); err == nil && strings.TrimSpace(kimi) != "" {
				summary = kimi
			}
		}

		papers = append(papers, model.Paper{
			ID:          strings.TrimSpace(entry.ID),
			Title:       normalizeWhitespace(entry.Title),
			Summary:     summary,
			URL:         entry.URL(),
			PublishedAt: parseTime(entry.Published),
			UpdatedAt:   parseTime(entry.Updated),
		})
	}

	return papers, nil
}

// FetchKimiSummary fetches and converts a Kimi summary for the given arXiv paper ID.
func (c *Client) FetchKimiSummary(ctx context.Context, paperID string) (string, error) {
	endpoint := fmt.Sprintf("%s/arxiv/kimi?paper=%s", c.baseURL, url.QueryEscape(paperID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "paper-radar/0.2.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("kimi status: %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", err
	}

	text := htmlToMarkdown(string(body))
	return strings.TrimSpace(text), nil
}

func (c *Client) resolveFeedURL(topicQuery string) string {
	q := strings.TrimSpace(topicQuery)
	if q == "" {
		return c.baseURL + "/arxiv/cs.AI/feed"
	}

	if strings.HasPrefix(q, "http://") || strings.HasPrefix(q, "https://") {
		return q
	}
	if strings.HasPrefix(q, "/") {
		return strings.TrimSuffix(c.baseURL, "/") + q
	}
	if strings.HasPrefix(q, "arxiv/") {
		if strings.HasSuffix(q, "/feed") {
			return c.baseURL + "/" + q
		}
		return c.baseURL + "/" + q + "/feed"
	}
	if strings.HasSuffix(q, "/feed") {
		return c.baseURL + "/" + strings.TrimPrefix(q, "/")
	}

	// treat query as arXiv category (e.g. cs.AI)
	return fmt.Sprintf("%s/arxiv/%s/feed", c.baseURL, q)
}

type atomFeed struct {
	Entries []atomEntry `xml:"entry"`
}

type atomEntry struct {
	ID        string     `xml:"id"`
	Title     string     `xml:"title"`
	Summary   string     `xml:"summary"`
	Published string     `xml:"published"`
	Updated   string     `xml:"updated"`
	Links     []atomLink `xml:"link"`
}

type atomLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

func (e atomEntry) URL() string {
	for _, link := range e.Links {
		if link.Rel == "alternate" && link.Href != "" {
			return link.Href
		}
	}
	for _, link := range e.Links {
		if link.Href != "" {
			return link.Href
		}
	}
	return strings.TrimSpace(e.ID)
}

func parseTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func normalizeWhitespace(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\t", " ")
	return strings.Join(strings.Fields(value), " ")
}

var (
	arxivIDRe      = regexp.MustCompile(`([0-9]{4}\.[0-9]{4,5})(v[0-9]+)?`)
	faqQRe         = regexp.MustCompile(`<p\s+class="faq-q">\s*<strong>(Q\d+)</strong>\s*[:：]\s*(.*?)\s*</p>`)
	faqAOpenRe     = regexp.MustCompile(`<div\s+class="faq-a">\s*`)
	faqACloseRe    = regexp.MustCompile(`\s*</div>`)
	htmlTagRe      = regexp.MustCompile(`<[^>]+>`)
	multiNewlineRe = regexp.MustCompile(`\n{3,}`)
	horizSpaceRe   = regexp.MustCompile(`[^\S\n]+`)
)

// htmlToMarkdown converts the Kimi API HTML response into clean markdown.
// The Kimi API returns HTML wrappers (<p class="faq-q">, <div class="faq-a">)
// around properly formatted markdown content with newlines, headings, lists, etc.
// We extract the markdown content while preserving its formatting.
func htmlToMarkdown(input string) string {
	text := input

	// Decode HTML entities first
	text = strings.NewReplacer(
		"&nbsp;", " ",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&#39;", "'",
		"&quot;", `"`,
		"&#x27;", "'",
	).Replace(text)

	// Convert FAQ question headers to markdown Q markers
	text = faqQRe.ReplaceAllString(text, "\n$1 : $2")

	// Remove FAQ answer div wrappers
	text = faqAOpenRe.ReplaceAllString(text, "\n")
	text = faqACloseRe.ReplaceAllString(text, "\n")

	// Strip remaining HTML tags (replace with empty, NOT space)
	text = htmlTagRe.ReplaceAllString(text, "")

	// Collapse horizontal whitespace (spaces/tabs) but preserve newlines
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(horizSpaceRe.ReplaceAllString(line, " "), " ")
	}
	text = strings.Join(lines, "\n")

	// Clean up excessive newlines (3+ → 2)
	text = multiNewlineRe.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

// stripHTML is a simple HTML stripper for non-Kimi content (backward compat).
func stripHTML(input string) string {
	text := strings.NewReplacer(
		"&nbsp;", " ",
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&#39;", "'",
		"&quot;", `"`,
	).Replace(input)
	text = htmlTagRe.ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func extractPaperID(entryID string) string {
	m := arxivIDRe.FindStringSubmatch(entryID)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}
