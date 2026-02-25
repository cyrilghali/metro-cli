package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Install shell completions",
	Long: `Install shell completions for metro.

Detects your current shell and installs completions automatically.
Supports bash, zsh, and fish.

Examples:
  metro completion            # auto-detect shell and install
  metro completion --print    # print the script to stdout instead`,
	RunE: runCompletion,
}

var completionPrint bool

func init() {
	completionCmd.Flags().BoolVar(&completionPrint, "print", false, "print completion script to stdout instead of installing")
	rootCmd.AddCommand(completionCmd)
}

func runCompletion(cmd *cobra.Command, args []string) error {
	shell := detectShell()
	if shell == "" {
		return fmt.Errorf("could not detect your shell\n\nSet the SHELL environment variable, or use one of:\n  metro completion --print > <file>")
	}

	// --print: just dump to stdout
	if completionPrint {
		return generateCompletion(shell, os.Stdout)
	}

	// Generate the script into a buffer
	var buf bytes.Buffer
	if err := generateCompletion(shell, &buf); err != nil {
		return err
	}

	// Determine install path
	dest, err := completionPath(shell)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	// Write the file
	if err := os.WriteFile(dest, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing %s: %w\n\nTry: metro completion --print > <file>", dest, err)
	}

	fmt.Printf("Completions installed for %s:\n  %s\n", shell, dest)

	// Shell-specific post-install hints
	switch shell {
	case "zsh":
		// Check if compinit is likely set up
		fmt.Println("\nRestart your shell, or run:")
		fmt.Println("  source " + dest)
	case "bash":
		fmt.Println("\nRestart your shell, or run:")
		fmt.Printf("  source %s\n", dest)
	case "fish":
		fmt.Println("\nCompletions will be available in new fish sessions.")
	}

	return nil
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ""
	}
	base := filepath.Base(shell)
	switch {
	case strings.Contains(base, "zsh"):
		return "zsh"
	case strings.Contains(base, "bash"):
		return "bash"
	case strings.Contains(base, "fish"):
		return "fish"
	}
	return ""
}

func generateCompletion(shell string, w io.Writer) error {
	switch shell {
	case "bash":
		return rootCmd.GenBashCompletionV2(w, true)
	case "zsh":
		return rootCmd.GenZshCompletion(w)
	case "fish":
		return rootCmd.GenFishCompletion(w, true)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

func completionPath(shell string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	switch shell {
	case "bash":
		// Linux: ~/.local/share/bash-completion/completions/metro
		// macOS: ~/.local/share/bash-completion/completions/metro
		// (works with bash-completion v2, doesn't require root)
		return filepath.Join(home, ".local", "share", "bash-completion", "completions", "metro"), nil
	case "zsh":
		// ~/.zsh/completions/_metro (add to fpath in .zshrc if needed)
		return filepath.Join(home, ".zsh", "completions", "_metro"), nil
	case "fish":
		// ~/.config/fish/completions/metro.fish (standard fish path)
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, ".config", "fish", "completions", "metro.fish"), nil
		}
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(home, ".config")
		}
		return filepath.Join(configDir, "fish", "completions", "metro.fish"), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}
}
