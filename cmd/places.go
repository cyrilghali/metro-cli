package cmd

import (
	"fmt"
	"strings"

	"github.com/cyrilghali/metro-cli/internal/client"
	"github.com/cyrilghali/metro-cli/internal/config"
	"github.com/cyrilghali/metro-cli/internal/model"
	"github.com/spf13/cobra"
)

var placesCmd = &cobra.Command{
	Use:   "places",
	Short: "Manage saved places",
	Long: `Save, list, and remove named places for quick access.

Saved places let you skip the station picker when looking up departures.
For example, save your home station and then just run "metro d home".

Examples:
  metro places                          # list saved places
  metro places save home chatelet       # save "chatelet" as "home"
  metro places save work "la defense"   # save "la defense" as "work"
  metro places default home             # set default for "metro d"
  metro places remove home              # remove saved place

  metro d home                          # use saved place directly
  metro d                               # uses the default place`,
	RunE: runPlacesList,
}

var placesSaveCmd = &cobra.Command{
	Use:   "save <alias> <station or address>",
	Short: "Save a place with a name",
	Long: `Search for a station or address and save it under a short alias.

Examples:
  metro places save home chatelet
  metro places save work "gare de lyon"
  metro places save gym "73 rue rivoli"`,
	Args: cobra.MinimumNArgs(2),
	RunE: runPlacesSave,
}

var placesRemoveCmd = &cobra.Command{
	Use:     "remove <alias>",
	Aliases: []string{"rm", "delete"},
	Short:   "Remove a saved place",
	Args:    cobra.ExactArgs(1),
	RunE:    runPlacesRemove,
}

var placesDefaultCmd = &cobra.Command{
	Use:   "default <alias>",
	Short: "Set the default place for \"metro d\" with no arguments",
	Long: `Mark a saved place as the default, used when you run "metro d" with no station.

Examples:
  metro places default home
  metro places default work`,
	Args: cobra.ExactArgs(1),
	RunE: runPlacesDefault,
}

func init() {
	placesCmd.AddCommand(placesSaveCmd)
	placesCmd.AddCommand(placesRemoveCmd)
	placesCmd.AddCommand(placesDefaultCmd)
	rootCmd.AddCommand(placesCmd)
}

func runPlacesList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if len(cfg.Places) == 0 {
		fmt.Println("No saved places.")
		fmt.Println("\nSave one with:")
		fmt.Println("  metro places save home chatelet")
		return nil
	}

	fmt.Print("Saved places:\n\n")
	for alias, p := range cfg.Places {
		label := p.Type
		if label == "StopArea" {
			label = "Stop"
		}
		city := ""
		if p.City != "" {
			city = " - " + p.City
		}
		def := ""
		if cfg.DefaultPlace == alias {
			def = " (default)"
		}
		fmt.Printf("  \033[1m%-12s\033[0m %s (%s%s)%s\n", alias, p.Name, label, city, def)
	}
	fmt.Println("\nUse with: metro d <alias>")

	return nil
}

func runPlacesSave(cmd *cobra.Command, args []string) error {
	alias := strings.ToLower(args[0])
	query := strings.Join(args[1:], " ")

	c, err := client.New()
	if err != nil {
		return err
	}

	fmt.Printf("Searching for \"%s\"...\n", query)
	places, err := c.SearchPlaces(query)
	if err != nil {
		return err
	}
	if len(places.Places) == 0 {
		return fmt.Errorf("no results found for \"%s\"", query)
	}

	// Filter to stop areas and addresses
	var candidates []model.PRIMPlace
	for _, p := range places.Places {
		if p.Type == "StopArea" || p.Type == "Address" {
			candidates = append(candidates, p)
		}
	}
	if len(candidates) == 0 {
		return fmt.Errorf("no stops or addresses found for \"%s\"", query)
	}

	place := candidates[0]
	if len(candidates) > 1 {
		place, err = pickPlace(candidates)
		if err != nil {
			return err
		}
	}

	// Save to config
	if err := savePlace(alias, place); err != nil {
		return err
	}

	city := ""
	if place.City != "" {
		city = " (" + place.City + ")"
	}
	fmt.Printf("\nSaved \"%s\" as \033[1m%s\033[0m%s\n", alias, place.Name, city)
	fmt.Printf("Now use: metro d %s\n", alias)
	return nil
}

func runPlacesRemove(cmd *cobra.Command, args []string) error {
	alias := strings.ToLower(args[0])

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if _, ok := cfg.Places[alias]; !ok {
		return fmt.Errorf("no saved place named \"%s\"", alias)
	}

	name := cfg.Places[alias].Name
	delete(cfg.Places, alias)

	// Clear default if it pointed to this alias
	if cfg.DefaultPlace == alias {
		cfg.DefaultPlace = ""
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Removed \"%s\" (%s)\n", alias, name)
	return nil
}

func runPlacesDefault(cmd *cobra.Command, args []string) error {
	alias := strings.ToLower(args[0])

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	place, ok := cfg.Places[alias]
	if !ok {
		return fmt.Errorf("no saved place named \"%s\"\nSave one first: metro places save %s <station>", alias, alias)
	}

	cfg.DefaultPlace = alias
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Default set to \"%s\" (%s)\n", alias, place.Name)
	fmt.Println("Now \"metro d\" with no arguments will use this place.")
	return nil
}

// savePlace persists a PRIMPlace as a saved place under the given alias.
func savePlace(alias string, place model.PRIMPlace) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if cfg.Places == nil {
		cfg.Places = make(map[string]config.SavedPlace)
	}

	cfg.Places[alias] = config.SavedPlace{
		Name: place.Name,
		Type: place.Type,
		ID:   place.ID,
		City: place.City,
		Lat:  place.Y,
		Lon:  place.X,
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	return nil
}
