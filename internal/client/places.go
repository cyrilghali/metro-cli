package client

import (
	"fmt"
	"net/url"

	"github.com/cyrilghali/metro-cli/internal/model"
)

// SearchPlaces searches for places (stops, addresses) matching a query string.
// Uses the PRIM custom /marketplace/places endpoint (returns metro line info).
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

// PlacesNearby finds stop points near given coordinates, filtered to metro.
func (c *Client) PlacesNearby(lon, lat string, radius int) (*model.PlacesNearbyResponse, error) {
	path := fmt.Sprintf("coords/%s;%s/places_nearby", lon, lat)
	params := url.Values{}
	params.Set("distance", fmt.Sprintf("%d", radius))
	params.Add("type[]", "stop_point")
	params.Set("filter", metroFilter)
	params.Set("count", "20")
	params.Set("depth", "2")

	data, err := c.navitia(path, params)
	if err != nil {
		return nil, fmt.Errorf("places nearby: %w", err)
	}
	return decode[model.PlacesNearbyResponse](data)
}
