package cmd

import (
	"fmt"
	"os"

	"github.com/cyrilghali/metro-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show metro CLI configuration",
	Long: `Show the current metro CLI configuration.

Config file:   ~/.metro.toml  (saved places and default)
API token:     PRIM_TOKEN environment variable

Examples:
  metro config`,
	RunE: runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Printf("Config file: %s\n\n", config.Path())

	// Token status
	if os.Getenv("PRIM_TOKEN") != "" {
		fmt.Println("  PRIM_TOKEN:     set")
	} else {
		fmt.Println("  PRIM_TOKEN:     (not set)")
	}

	// Default place
	if cfg.DefaultPlace != "" {
		if p, ok := cfg.Places[cfg.DefaultPlace]; ok {
			fmt.Printf("  Default place:  %s (%s)\n", cfg.DefaultPlace, p.Name)
		} else {
			fmt.Printf("  Default place:  %s (missing — run \"metro places save %s <station>\")\n", cfg.DefaultPlace, cfg.DefaultPlace)
		}
	} else {
		fmt.Println("  Default place:  (not set)")
	}

	// Saved places
	if len(cfg.Places) > 0 {
		fmt.Printf("  Saved places:   %d (run \"metro places\" to view)\n", len(cfg.Places))
	} else {
		fmt.Println("  Saved places:   (none)")
	}

	// Setup hints
	needsSetup := os.Getenv("PRIM_TOKEN") == "" || len(cfg.Places) == 0
	if needsSetup {
		fmt.Println("\nSetup:")
		if os.Getenv("PRIM_TOKEN") == "" {
			fmt.Println("  export PRIM_TOKEN=<your-token>")
			fmt.Println("  Get a free token at https://prim.iledefrance-mobilites.fr")
		}
		if len(cfg.Places) == 0 {
			fmt.Println("  metro places save home <station-name>")
			fmt.Println("  metro places default home")
		}
	}

	return nil
}
