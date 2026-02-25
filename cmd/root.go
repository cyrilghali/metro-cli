package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:               "metro",
	Short:             "Paris metro departures and disruptions",
	Long:              "A CLI tool to check next metro departures near you and current traffic disruptions in Paris.",
	Version:           Version,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
