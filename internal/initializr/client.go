package initializr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	MetadataMaxBytes = 5 << 20
	MetadataTimeout  = 15 * time.Second
	MediaType        = "application/vnd.initializr.v2.3+json"
)

type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
	UserAgent  string
}

func NewClient(baseURL, userAgent string, timeout time.Duration) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil || !u.IsAbs() || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return nil, fmt.Errorf("Initializr URL must be an absolute http or https URL")
	}
	if timeout <= 0 {
		timeout = MetadataTimeout
	}
	return &Client{BaseURL: u, HTTPClient: &http.Client{Timeout: timeout}, UserAgent: userAgent}, nil
}

func (c *Client) FetchMetadata(ctx context.Context) (Metadata, error) {
	u := *c.BaseURL
	u.Path = strings.TrimRight(u.Path, "/") + "/"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Metadata{}, err
	}
	req.Header.Set("Accept", MediaType)
	req.Header.Set("User-Agent", c.UserAgent)
	response, err := c.HTTPClient.Do(req)
	if err != nil {
		return Metadata{}, fmt.Errorf("requesting metadata from %s: %w", u.String(), err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 8192))
		return Metadata{}, fmt.Errorf("metadata request to %s returned %s: %s", u.String(), response.Status, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, MetadataMaxBytes+1))
	if err != nil {
		return Metadata{}, fmt.Errorf("reading Initializr metadata: %w", err)
	}
	if len(body) > MetadataMaxBytes {
		return Metadata{}, fmt.Errorf("metadata response exceeds the %d MiB limit", MetadataMaxBytes>>20)
	}
	var metadata Metadata
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	if err := decoder.Decode(&metadata); err != nil {
		return Metadata{}, fmt.Errorf("decoding Initializr metadata: %w", err)
	}
	if err := metadata.Validate(); err != nil {
		return Metadata{}, err
	}
	return metadata, nil
}
