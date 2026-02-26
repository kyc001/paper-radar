package arxiv

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kyc001/paper-radar/internal/model"
)

const defaultBaseURL = "https://export.arxiv.org/api/query"

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

func (c *Client) Fetch(ctx context.Context, query string, maxResults int) ([]model.Paper, error) {
	params := url.Values{}
	params.Set("search_query", query)
	params.Set("sortBy", "submittedDate")
	params.Set("sortOrder", "descending")
	params.Set("max_results", fmt.Sprintf("%d", maxResults))

	endpoint := c.baseURL + "?" + params.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "paper-radar/0.1.0")

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return nil, fmt.Errorf("unexpected status %s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	var feed atomFeed
	if err := xml.NewDecoder(response.Body).Decode(&feed); err != nil {
		return nil, err
	}

	papers := make([]model.Paper, 0, len(feed.Entries))
	for _, entry := range feed.Entries {
		papers = append(papers, model.Paper{
			ID:          strings.TrimSpace(entry.ID),
			Title:       normalizeWhitespace(entry.Title),
			Summary:     normalizeWhitespace(entry.Summary),
			URL:         entry.URL(),
			PublishedAt: parseTime(entry.Published),
			UpdatedAt:   parseTime(entry.Updated),
		})
	}

	return papers, nil
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
