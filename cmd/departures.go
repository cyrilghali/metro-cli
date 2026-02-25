package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cyrilghali/metro-cli/internal/client"
	"github.com/cyrilghali/metro-cli/internal/config"
	"github.com/cyrilghali/metro-cli/internal/display"
	"github.com/cyrilghali/metro-cli/internal/location"
	"github.com/cyrilghali/metro-cli/internal/model"
	"github.com/spf13/cobra"
)

var here bool

var departuresCmd = &cobra.Command{
	Use:   "departures [station or address]",
	Short: "Show next metro departures near a location",
	Long: `Show next metro departures near a station or address.

Examples:
  metro departures chatelet
  metro departures "gare de lyon"
  metro departures "73 rue rivoli"
  metro departures --here             # auto-detect location via browser
  metro departures                    # uses default station from config`,
	RunE: runDepartures,
}

func init() {
	departuresCmd.Flags().BoolVar(&here, "here", false, "auto-detect your location via browser geolocation")
	rootCmd.AddCommand(departuresCmd)
}

func runDepartures(cmd *cobra.Command, args []string) error {
	c, err := client.New()
	if err != nil {
		return err
	}

	// --here: use browser geolocation
	if here {
		return runDeparturesHere(c)
	}

	// Station/address search
	query := strings.Join(args, " ")
	if query == "" {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		if cfg.DefaultStation == "" {
			return fmt.Errorf("no station provided and no default set\nUsage: metro departures <station>\n       metro departures --here\nOr:    metro config --default-station chatelet")
		}
		query = cfg.DefaultStation
	}

	fmt.Printf("Searching for \"%s\"...\n", query)
	places, err := c.SearchPlaces(query)
	if err != nil {
		return err
	}
	if len(places.Places) == 0 {
		return fmt.Errorf("no results found for \"%s\"", query)
	}

	// Filter: metro stop areas + addresses
	var candidates []model.PRIMPlace
	for _, p := range places.Places {
		if p.Type == "StopArea" && hasMetro(p) {
			candidates = append(candidates, p)
		} else if p.Type == "Address" {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		for _, p := range places.Places {
			if p.Type == "StopArea" || p.Type == "Address" {
				candidates = append(candidates, p)
			}
		}
	}
	if len(candidates) == 0 {
		return fmt.Errorf("no metro stops or addresses found for \"%s\"", query)
	}

	place := candidates[0]
	if len(candidates) > 1 {
		place, err = pickPlace(candidates)
		if err != nil {
			fmt.Printf("  %s\n", err)
			place = candidates[0]
		}
	}

	fmt.Println()

	if place.Type == "StopArea" {
		return showStopAreaDepartures(c, place.ID, place.Name, place.City)
	}
	return showNearbyDepartures(c, place.Name+" "+place.City)
}

// runDeparturesHere uses browser geolocation to find the user's position.
func runDeparturesHere(c *client.Client) error {
	fmt.Println("Locating you... (opening browser)")
	lat, lon, err := location.GetLocation(30 * time.Second)
	if err != nil {
		return fmt.Errorf("could not get location: %w", err)
	}
	fmt.Printf("Found you at %.6f, %.6f\n", lat, lon)
	return showDeparturesAtCoords(c, fmt.Sprintf("%.6f", lon), fmt.Sprintf("%.6f", lat))
}

// showStopAreaDepartures fetches and displays departures for a specific stop area.
func showStopAreaDepartures(c *client.Client, stopID, name, city string) error {
	label := name
	if city != "" {
		label = fmt.Sprintf("\033[1m%s\033[0m (%s)", name, city)
	} else {
		label = fmt.Sprintf("\033[1m%s\033[0m", name)
	}
	fmt.Println(label)
	deps, err := c.Departures(stopID, 15)
	if err != nil {
		return fmt.Errorf("fetching departures: %w", err)
	}
	display.Departures(name, deps.Departures)
	fmt.Println()
	return nil
}

// showNearbyDepartures resolves an address to coordinates, then shows nearby metro departures.
func showNearbyDepartures(c *client.Client, addressQuery string) error {
	fmt.Printf("Finding metro stops near %s...\n", addressQuery)
	navResp, err := c.NavitiaPlaces(addressQuery)
	if err != nil {
		return fmt.Errorf("resolving address: %w", err)
	}

	var lon, lat string
	for _, np := range navResp.Places {
		if np.Address != nil {
			lon, lat = np.Address.Coord.Lon, np.Address.Coord.Lat
			break
		}
		if np.StopArea != nil {
			lon, lat = np.StopArea.Coord.Lon, np.StopArea.Coord.Lat
			break
		}
	}
	if lon == "" {
		return fmt.Errorf("could not resolve coordinates for \"%s\"", addressQuery)
	}

	return showDeparturesAtCoords(c, lon, lat)
}

// showDeparturesAtCoords finds metro stops near coordinates and shows departures for each.
func showDeparturesAtCoords(c *client.Client, lon, lat string) error {
	fmt.Printf("Finding metro stops nearby...\n\n")
	nearby, err := c.PlacesNearby(lon, lat, 500)
	if err != nil {
		return err
	}

	seen := make(map[string]bool)
	var areas []model.StopArea
	for _, pn := range nearby.PlacesNearby {
		var sa *model.StopArea
		if pn.StopPoint != nil && pn.StopPoint.StopArea != nil {
			sa = pn.StopPoint.StopArea
		}
		if sa != nil && !seen[sa.ID] {
			seen[sa.ID] = true
			areas = append(areas, *sa)
		}
	}

	if len(areas) == 0 {
		return fmt.Errorf("no metro stops found within 500m")
	}

	for _, sa := range areas {
		fmt.Printf("\033[1m%s\033[0m\n", sa.Name)
		deps, err := c.Departures(sa.ID, 10)
		if err != nil {
			fmt.Printf("  \033[31mError: %v\033[0m\n", err)
			continue
		}
		display.Departures(sa.Name, deps.Departures)
		fmt.Println()
	}
	return nil
}

func hasMetro(p model.PRIMPlace) bool {
	for _, m := range p.Modes {
		if m == "Metro" {
			return true
		}
	}
	return false
}

func pickPlace(places []model.PRIMPlace) (model.PRIMPlace, error) {
	fmt.Println("\nMultiple results found:")
	for i, p := range places {
		label := p.Type
		if label == "StopArea" {
			label = "Stop"
		}
		metroLines := metroLinesList(p)
		extra := ""
		if metroLines != "" {
			extra = " [" + metroLines + "]"
		}
		if p.City != "" {
			extra += " - " + p.City
		}
		fmt.Printf("  %d. %s (%s%s)\n", i+1, p.Name, label, extra)
	}

	fmt.Print("\nPick a number: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(places) {
		return places[0], fmt.Errorf("invalid choice, using first result")
	}
	return places[idx-1], nil
}

func metroLinesList(p model.PRIMPlace) string {
	var lines []string
	for _, l := range p.Lines {
		for _, m := range l.Mode {
			if m.ID == "physical_mode:Metro" {
				lines = append(lines, "M"+l.ShortName)
				break
			}
		}
	}
	return strings.Join(lines, ", ")
}
