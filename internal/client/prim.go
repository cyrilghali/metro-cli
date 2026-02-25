package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	primBase    = "https://prim.iledefrance-mobilites.fr/marketplace"
	navitiaBase = primBase + "/v2/navitia"
)

type Client struct {
	apiKey string
	http   *http.Client
}

func New() (*Client, error) {
	key := os.Getenv("PRIM_TOKEN")
	if key == "" {
		return nil, fmt.Errorf("PRIM_TOKEN not set\n\nGet a free token at https://prim.iledefrance-mobilites.fr\nThen set it:\n  export PRIM_TOKEN=<your-token>\n\nTo make it permanent, add the line above to your shell profile (~/.bashrc, ~/.zshrc, etc.)")
	}

	return &Client{
		apiKey: key,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}, nil
}

// navitia makes a GET request to the Navitia v2 endpoint (no /coverage/ prefix).
func (c *Client) navitia(path string, params url.Values) ([]byte, error) {
	u := navitiaBase + "/" + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	return c.doGet(u)
}

// prim makes a GET request to the PRIM marketplace root endpoint.
func (c *Client) prim(path string, params url.Values) ([]byte, error) {
	u := primBase + "/" + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	return c.doGet(u)
}

func (c *Client) doGet(u string) ([]byte, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("not found (the API returned no results)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}

	return body, nil
}

func decode[T any](data []byte) (*T, error) {
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &result, nil
}
