package cmd

import (
	"fmt"

	"github.com/cyrilghali/metro-cli/internal/client"
	"github.com/cyrilghali/metro-cli/internal/display"
	"github.com/spf13/cobra"
)

var lineFilter string

var disruptionsCmd = &cobra.Command{
	Use:   "disruptions",
	Short: "Show current metro disruptions",
	Long: `Show current traffic disruptions on Paris metro lines.

Examples:
  metro disruptions
  metro disruptions --line M14
  metro disruptions --line 1`,
	RunE: runDisruptions,
}

func init() {
	disruptionsCmd.Flags().StringVar(&lineFilter, "line", "", "filter by line (e.g. M1, 14)")
	rootCmd.AddCommand(disruptionsCmd)
}

func runDisruptions(cmd *cobra.Command, args []string) error {
	c, err := client.New()
	if err != nil {
		return err
	}

	fmt.Println("Fetching metro disruptions...")
	resp, err := c.MetroLines()
	if err != nil {
		return err
	}

	fmt.Println()
	display.DisruptionsSummary(resp, lineFilter)
	return nil
}
