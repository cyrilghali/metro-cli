package client

import (
	"fmt"
	"net/url"

	"github.com/cyrilghali/metro-cli/internal/model"
)

// Departures fetches next departures at a stop area, optionally filtered by mode.
// If modeFilter is empty, all transport modes are returned.
func (c *Client) Departures(stopAreaID string, count int, modeFilter string) (*model.DeparturesResponse, error) {
	path := fmt.Sprintf("stop_areas/%s/departures", url.PathEscape(stopAreaID))
	params := url.Values{}
	params.Set("count", fmt.Sprintf("%d", count))
	params.Set("data_freshness", "realtime")
	params.Set("depth", "2")
	if modeFilter != "" {
		params.Set("filter", modeFilter)
	}

	data, err := c.navitia(path, params)
	if err != nil {
		return nil, fmt.Errorf("fetching departures: %w", err)
	}
	return decode[model.DeparturesResponse](data)
}
