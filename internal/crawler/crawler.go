package crawler

import (
	"net/url"
	"strings"
	"wcheck/internal/engine"
)

// Crawler handles link discovery
type Crawler struct {
	BaseURL string
	Engine  *engine.Engine
}

// NewCrawler creates a new crawler for a base URL
func NewCrawler(baseURL string, eng *engine.Engine) *Crawler {
	return &Crawler{
		BaseURL: baseURL,
		Engine:  eng,
	}
}

// GetLinks returns a list of unique internal links found on the base URL
func (c *Crawler) GetLinks() ([]string, error) {
	rawHrefs, err := c.Engine.ExtractLinks(c.BaseURL)
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	uniqueLinks := make(map[string]struct{})
	uniqueLinks[c.BaseURL] = struct{}{} // Always include the base URL

	for _, href := range rawHrefs {
		u, err := url.Parse(href)
		if err != nil || href == "" {
			continue
		}

		// Resolve relative to base
		resolved := base.ResolveReference(u)

		// Remove fragment (anchor links)
		resolved.Fragment = ""
		resolved.RawQuery = "" // Optional: standardizing by removing queries too? User said # but usually better to clean.

		// Check if same domain
		if resolved.Host != base.Host {
			continue
		}

		// Normalize: remove trailing slash for deduplication consistency
		link := strings.TrimSuffix(resolved.String(), "/")
		if link == "" {
			continue
		}

		uniqueLinks[link] = struct{}{}
	}

	var links []string
	for link := range uniqueLinks {
		links = append(links, link)
	}

	return links, nil
}
