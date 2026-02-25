package client

import (
	"fmt"
	"net/url"

	"github.com/cyrilghali/metro-cli/internal/model"
)

// MetroLines fetches all metro lines with their associated disruptions.
func (c *Client) MetroLines() (*model.LinesResponse, error) {
	params := url.Values{}
	params.Set("filter", metroFilter)
	params.Set("count", "20")
	params.Set("depth", "1")

	data, err := c.navitia("lines", params)
	if err != nil {
		return nil, fmt.Errorf("fetching metro lines: %w", err)
	}
	return decode[model.LinesResponse](data)
}

// DisruptionsForLine fetches disruptions for a specific line.
func (c *Client) DisruptionsForLine(lineID string) (*model.DeparturesResponse, error) {
	path := fmt.Sprintf("lines/%s/disruptions", lineID)
	params := url.Values{}
	params.Set("count", "20")

	data, err := c.navitia(path, params)
	if err != nil {
		return nil, fmt.Errorf("fetching disruptions: %w", err)
	}
	return decode[model.DeparturesResponse](data)
}
