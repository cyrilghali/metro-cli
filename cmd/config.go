package cmd

import (
	"fmt"
	"strings"

	"github.com/cyrilghali/metro-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgToken          string
	cfgDefaultStation string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage metro CLI configuration",
	Long: `View or update the metro CLI configuration stored in ~/.metro.toml.

Examples:
  metro config --token YOUR_API_TOKEN
  metro config --default-station chatelet
  metro config                              # show current config`,
	RunE: runConfig,
}

func init() {
	configCmd.Flags().StringVar(&cfgToken, "token", "", "set PRIM API token")
	configCmd.Flags().StringVar(&cfgDefaultStation, "default-station", "", "set default station for departures")
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	changed := false

	if cfgToken != "" {
		cfg.Token = cfgToken
		changed = true
	}
	if cfgDefaultStation != "" {
		cfg.DefaultStation = cfgDefaultStation
		changed = true
	}

	if changed {
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		if cfgToken != "" {
			fmt.Println("API token saved.")
		}
		if cfgDefaultStation != "" {
			fmt.Printf("Default station set to \"%s\".\n", cfgDefaultStation)
		}
		fmt.Printf("Config saved to %s\n", config.Path())
		return nil
	}

	// Show current config
	fmt.Printf("Config file: %s\n\n", config.Path())
	if cfg.Token != "" {
		masked := strings.Repeat("*", len(cfg.Token))
		if len(cfg.Token) > 8 {
			masked = cfg.Token[:4] + strings.Repeat("*", len(cfg.Token)-8) + cfg.Token[len(cfg.Token)-4:]
		}
		fmt.Printf("  Token:           %s\n", masked)
	} else {
		fmt.Println("  Token:           (not set)")
	}
	if cfg.DefaultStation != "" {
		fmt.Printf("  Default station: %s\n", cfg.DefaultStation)
	} else {
		fmt.Println("  Default station: (not set)")
	}
	if len(cfg.Places) > 0 {
		fmt.Printf("  Saved places:    %d (run \"metro places\" to view)\n", len(cfg.Places))
	} else {
		fmt.Println("  Saved places:    (none)")
	}

	if cfg.Token == "" || cfg.DefaultStation == "" {
		fmt.Println("\nSetup:")
		if cfg.Token == "" {
			fmt.Println("  metro config --token <your-api-token>")
			fmt.Println("  Get a free token at https://prim.iledefrance-mobilites.fr")
		}
		if cfg.DefaultStation == "" {
			fmt.Println("  metro config --default-station <station-name>")
		}
	}

	return nil
}
