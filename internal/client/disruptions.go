package client

import (
	"fmt"
	"net/url"

	"github.com/cyrilghali/metro-cli/internal/model"
)

// Lines fetches lines with their associated disruptions, optionally filtered by mode.
// If modeFilter is empty, all lines are returned (paginated).
func (c *Client) Lines(modeFilter string, count int) (*model.LinesResponse, error) {
	params := url.Values{}
	if modeFilter != "" {
		params.Set("filter", modeFilter)
	}
	params.Set("count", fmt.Sprintf("%d", count))
	params.Set("depth", "1")

	data, err := c.navitia("lines", params)
	if err != nil {
		return nil, fmt.Errorf("fetching lines: %w", err)
	}
	return decode[model.LinesResponse](data)
}
