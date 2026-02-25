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

var (
	here          bool
	herePort      int
	hereCacheTTL  time.Duration
	modeFlag      string
)

var departuresCmd = &cobra.Command{
	Use:   "departures [station or address]",
	Short: "Show next departures near a location",
	Long: `Show next departures near a station or address.
Active disruptions on displayed lines are shown automatically.

Modes:
  metro   Metro lines M1-M14         (default)
  rer     RER lines A-E
  train   Transilien / suburban rail
  tram    Tramway lines T1-T13
  bus     Bus lines
  all     All transport types

Examples:
  metro departures chatelet
  metro departures "gare de lyon"
  metro departures "73 rue rivoli"
  metro departures --mode rer
  metro departures chatelet --mode all

  # auto-detect location via browser
  metro departures --here
  metro departures --here --port 8080
  metro departures --here --cache 5m`,
	RunE: runDepartures,
}

func init() {
	departuresCmd.Flags().BoolVar(&here, "here", false, "detect location via browser (opens a tab)")
	departuresCmd.Flags().IntVar(&herePort, "port", 0, "fixed port for --here server (default: random)")
	departuresCmd.Flags().DurationVar(&hereCacheTTL, "cache", 0, "reuse cached location within this duration (e.g. 5m, 1h)")
	departuresCmd.Flags().StringVarP(&modeFlag, "mode", "m", "metro", "transport filter (see modes above)")
	rootCmd.AddCommand(departuresCmd)
}

func runDepartures(cmd *cobra.Command, args []string) error {
	c, err := client.New()
	if err != nil {
		return err
	}

	mode, err := model.ParseMode(modeFlag)
	if err != nil {
		return err
	}

	// --here: use browser geolocation
	if here {
		return runDeparturesHere(c, mode)
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

	// Filter: stop areas + addresses
	var candidates []model.PRIMPlace
	for _, p := range places.Places {
		if p.Type == "StopArea" && hasTransport(p, mode) {
			candidates = append(candidates, p)
		} else if p.Type == "Address" {
			candidates = append(candidates, p)
		}
	}
	// Fallback: any stop area or address
	if len(candidates) == 0 {
		for _, p := range places.Places {
			if p.Type == "StopArea" || p.Type == "Address" {
				candidates = append(candidates, p)
			}
		}
	}
	if len(candidates) == 0 {
		return fmt.Errorf("no stops or addresses found for \"%s\"", query)
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
		return showStopAreaDepartures(c, place.ID, place.Name, place.City, mode)
	}
	return showNearbyDepartures(c, place.Name+" "+place.City, mode)
}

// runDeparturesHere uses browser geolocation to find the user's position.
func runDeparturesHere(c *client.Client, mode model.TransportMode) error {
	// Try cache first
	if hereCacheTTL > 0 {
		if lat, lon, err := location.LoadCache(hereCacheTTL); err == nil {
			fmt.Printf("Using cached location (%.6f, %.6f)\n", lat, lon)
			return showDeparturesAtCoords(c, fmt.Sprintf("%.6f", lon), fmt.Sprintf("%.6f", lat), mode)
		}
	}

	fmt.Println("Locating you... (opening browser)")
	lat, lon, err := location.GetLocation(30*time.Second, herePort)
	if err != nil {
		return fmt.Errorf("could not get location: %w", err)
	}
	fmt.Printf("Found you at %.6f, %.6f\n", lat, lon)

	if hereCacheTTL > 0 {
		location.SaveCache(lat, lon)
	}

	return showDeparturesAtCoords(c, fmt.Sprintf("%.6f", lon), fmt.Sprintf("%.6f", lat), mode)
}

// showStopAreaDepartures fetches and displays departures for a specific stop area.
func showStopAreaDepartures(c *client.Client, stopID, name, city string, mode model.TransportMode) error {
	label := name
	if city != "" {
		label = fmt.Sprintf("\033[1m%s\033[0m (%s)", name, city)
	} else {
		label = fmt.Sprintf("\033[1m%s\033[0m", name)
	}
	fmt.Println(label)
	deps, err := c.Departures(stopID, 60, mode.Filter)
	if err != nil {
		return fmt.Errorf("fetching departures: %w", err)
	}
	display.Departures(deps.Departures, deps.Disruptions, mode.IsAll())
	fmt.Println()
	return nil
}

// showNearbyDepartures resolves an address to coordinates, then shows nearby departures.
func showNearbyDepartures(c *client.Client, addressQuery string, mode model.TransportMode) error {
	fmt.Printf("Finding stops near %s...\n", addressQuery)
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

	return showDeparturesAtCoords(c, lon, lat, mode)
}

// showDeparturesAtCoords finds stops near coordinates and shows departures for each.
func showDeparturesAtCoords(c *client.Client, lon, lat string, mode model.TransportMode) error {
	fmt.Printf("Finding stops nearby...\n\n")
	nearby, err := c.PlacesNearby(lon, lat, 500, mode.Filter)
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
		return fmt.Errorf("no stops found within 500m")
	}

	for _, sa := range areas {
		fmt.Printf("\033[1m%s\033[0m\n", sa.Name)
		deps, err := c.Departures(sa.ID, 40, mode.Filter)
		if err != nil {
			fmt.Printf("  \033[31mError: %v\033[0m\n", err)
			continue
		}
		display.Departures(deps.Departures, deps.Disruptions, mode.IsAll())
		fmt.Println()
	}
	return nil
}

// hasTransport checks if a PRIM place has the requested transport mode.
func hasTransport(p model.PRIMPlace, mode model.TransportMode) bool {
	if mode.IsAll() {
		return true
	}
	// Map physical_mode IDs to our mode names
	modeMap := map[string]string{
		"physical_mode:Metro":        "metro",
		"physical_mode:RapidTransit": "rer",
		"physical_mode:LocalTrain":   "train",
		"physical_mode:Tramway":      "tram",
		"physical_mode:Bus":          "bus",
	}
	// Also check the PRIM "modes" array (uses display names)
	nameMap := map[string]string{
		"Metro":   "metro",
		"RER":     "rer",
		"Train":   "train",
		"Tramway": "tram",
		"Bus":     "bus",
	}
	for _, m := range p.Modes {
		if nameMap[m] == mode.Name {
			return true
		}
	}
	for _, l := range p.Lines {
		for _, m := range l.Mode {
			if modeMap[m.ID] == mode.Name {
				return true
			}
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
		lines := linesList(p)
		extra := ""
		if lines != "" {
			extra = " [" + lines + "]"
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

// linesList returns a display string of all transport lines at a place.
func linesList(p model.PRIMPlace) string {
	var lines []string
	for _, l := range p.Lines {
		if len(l.Mode) == 0 {
			continue
		}
		prefix := ""
		for _, m := range l.Mode {
			switch m.ID {
			case "physical_mode:Metro":
				prefix = "M"
			case "physical_mode:RapidTransit":
				prefix = "RER "
			case "physical_mode:Tramway":
				prefix = "T"
			}
			break
		}
		lines = append(lines, prefix+l.ShortName)
	}
	return strings.Join(lines, ", ")
}
