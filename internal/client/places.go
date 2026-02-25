package client

import (
	"fmt"
	"net/url"

	"github.com/cyrilghali/metro-cli/internal/model"
)

// SearchPlaces searches for places (stops, addresses) matching a query string.
// Uses the PRIM custom /marketplace/places endpoint (returns line info).
func (c *Client) SearchPlaces(query string) (*model.PRIMPlacesResponse, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("coverage", "fr-idf")

	data, err := c.prim("places", params)
	if err != nil {
		return nil, fmt.Errorf("searching places: %w", err)
	}
	return decode[model.PRIMPlacesResponse](data)
}

// NavitiaPlaces searches using Navitia's places endpoint (returns WGS84 coords).
// Used for addresses where we need proper lon/lat for nearby lookups.
func (c *Client) NavitiaPlaces(query string) (*model.NavitiaPlacesResponse, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Add("type[]", "address")
	params.Add("type[]", "stop_area")
	params.Set("count", "5")

	data, err := c.navitia("places", params)
	if err != nil {
		return nil, fmt.Errorf("navitia places: %w", err)
	}
	return decode[model.NavitiaPlacesResponse](data)
}

// PlacesNearby finds stop points near given coordinates, optionally filtered by mode.
// If modeFilter is empty, all stop points are returned.
func (c *Client) PlacesNearby(lon, lat string, radius int, modeFilter string) (*model.PlacesNearbyResponse, error) {
	path := fmt.Sprintf("coords/%s;%s/places_nearby", url.PathEscape(lon), url.PathEscape(lat))
	params := url.Values{}
	params.Set("distance", fmt.Sprintf("%d", radius))
	params.Add("type[]", "stop_point")
	if modeFilter != "" {
		params.Set("filter", modeFilter)
	}
	params.Set("count", "30")
	params.Set("depth", "2")

	data, err := c.navitia(path, params)
	if err != nil {
		return nil, fmt.Errorf("places nearby: %w", err)
	}
	return decode[model.PlacesNearbyResponse](data)
}
