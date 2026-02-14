// Package scdl provides a client for interacting with SoundCloud's internal API.
package scdl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"

// Client provides methods to interact with the SoundCloud API.
type Client struct {
	httpClient *http.Client
	clientID   string
}

// NewClient creates a new SoundCloud client by extracting a client_id
// from the SoundCloud website.
func NewClient() (*Client, error) {
	c := &Client{
		httpClient: &http.Client{},
	}

	clientID, err := c.extractClientID()
	if err != nil {
		return nil, fmt.Errorf("extract client_id: %w", err)
	}
	c.clientID = clientID

	return c, nil
}

func (c *Client) get(url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}

var (
	assetRe    = regexp.MustCompile(`src="(https://a-v2\.sndcdn\.com/assets/[^\s"]+)"`)
	clientIDRe = regexp.MustCompile(`client_id:"([^"]+)"`)
)

func (c *Client) extractClientID() (string, error) {
	body, err := c.get("https://soundcloud.com")
	if err != nil {
		return "", fmt.Errorf("fetch soundcloud.com: %w", err)
	}

	matches := assetRe.FindAllSubmatch(body, -1)
	if len(matches) == 0 {
		return "", fmt.Errorf("no asset URLs found on soundcloud.com")
	}

	for _, match := range matches {
		assetURL := string(match[1])

		assetBody, err := c.get(assetURL)
		if err != nil {
			continue
		}

		if m := clientIDRe.FindSubmatch(assetBody); len(m) > 1 {
			return string(m[1]), nil
		}
	}

	return "", fmt.Errorf("client_id not found in any asset bundle")
}
