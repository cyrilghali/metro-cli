package cmd

import (
	"fmt"

	"github.com/cyrilghali/metro-cli/internal/client"
	"github.com/cyrilghali/metro-cli/internal/display"
	"github.com/cyrilghali/metro-cli/internal/model"
	"github.com/spf13/cobra"
)

var (
	lineFilter      string
	disruptionMode  string
)

var disruptionsCmd = &cobra.Command{
	Use:   "disruptions",
	Short: "Show current traffic disruptions",
	Long: `Show current traffic disruptions across Ile-de-France transport lines.
Useful for a network-wide overview. For disruptions at a specific station,
use "metro departures <station>" instead (disruptions are shown inline).

Modes:
  metro   Metro lines M1-M14         (default)
  rer     RER lines A-E
  train   Transilien / suburban rail
  tram    Tramway lines T1-T13
  bus     Bus lines
  all     All transport types

Examples:
  metro disruptions
  metro disruptions --line M14
  metro disruptions --mode rer
  metro disruptions --mode all --line A`,
	RunE: runDisruptions,
}

func init() {
	disruptionsCmd.Flags().StringVar(&lineFilter, "line", "", "filter by line (e.g. M1, A, T3)")
	disruptionsCmd.Flags().StringVarP(&disruptionMode, "mode", "m", "metro", "transport filter (see modes above)")
	rootCmd.AddCommand(disruptionsCmd)
}

func runDisruptions(cmd *cobra.Command, args []string) error {
	c, err := client.New()
	if err != nil {
		return err
	}

	mode, err := model.ParseMode(disruptionMode)
	if err != nil {
		return err
	}

	if mode.IsAll() {
		return showAllDisruptions(c)
	}

	fmt.Printf("Fetching %s disruptions...\n\n", mode.Name)
	resp, err := c.Lines(mode.Filter, mode.MaxLines)
	if err != nil {
		return err
	}

	display.DisruptionsSummary(resp, lineFilter, mode)
	return nil
}

func showAllDisruptions(c *client.Client) error {
	fmt.Println("Fetching disruptions...")
	for _, name := range model.ModeNames {
		m := model.Modes[name]
		resp, err := c.Lines(m.Filter, m.MaxLines)
		if err != nil {
			fmt.Printf("  \033[31mError fetching %s: %v\033[0m\n", name, err)
			continue
		}
		display.DisruptionsSummary(resp, lineFilter, m)
		fmt.Println()
	}
	return nil
}
