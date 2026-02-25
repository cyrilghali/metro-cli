package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var checkOnly bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update metro to the latest version",
	Long: `Check for and install the latest version of metro from GitHub releases.

Examples:
  metro update          # download and install the latest version
  metro update --check  # check for updates without installing`,
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().BoolVar(&checkOnly, "check", false, "check for updates without installing")
	rootCmd.AddCommand(updateCmd)
}

const (
	releasesURL = "https://api.github.com/repos/cyrilghali/metro-cli/releases/latest"
	binaryName  = "metro"
)

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func runUpdate(cmd *cobra.Command, args []string) error {
	fmt.Println("Checking for updates...")

	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("checking latest release: %w", err)
	}

	latest := release.TagName
	current := Version
	if !strings.HasPrefix(current, "v") && current != "dev" {
		current = "v" + current
	}

	isDev := current == "dev"

	if checkOnly {
		fmt.Printf("  Current: %s\n", Version)
		fmt.Printf("  Latest:  %s\n", latest)
		if !isDev && current == latest {
			fmt.Println("\nAlready up to date.")
		} else if isDev {
			fmt.Println("\nYou are running a development build.")
			fmt.Println("Run `metro update` to switch to the latest release.")
		} else {
			fmt.Println("\nRun `metro update` to install.")
		}
		return nil
	}

	if !isDev && current == latest {
		fmt.Printf("Already up to date (%s).\n", latest)
		return nil
	}

	// Find the right asset for this OS/arch.
	assetName := expectedAssetName()
	var downloadURL string
	for _, a := range release.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return fmt.Errorf("no release asset found for %s/%s (looking for %s)", runtime.GOOS, runtime.GOARCH, assetName)
	}

	fmt.Printf("Downloading %s -> %s...\n", Version, latest)

	bin, err := downloadAndExtract(downloadURL, assetName)
	if err != nil {
		return fmt.Errorf("downloading update: %w", err)
	}

	if err := replaceBinary(bin); err != nil {
		return err
	}

	fmt.Printf("Updated to %s.\n", latest)
	return nil
}

// fetchLatestRelease queries the GitHub API for the latest release.
func fetchLatestRelease() (*ghRelease, error) {
	req, err := http.NewRequest("GET", releasesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no releases found — this is likely a fresh install from source")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %s", resp.Status)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &rel, nil
}

// expectedAssetName returns the archive name for the current platform.
func expectedAssetName() string {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("metro-cli_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)
}

// downloadAndExtract downloads the archive and extracts the binary.
func downloadAndExtract(url, assetName string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %s", resp.Status)
	}

	archive, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading download: %w", err)
	}

	if strings.HasSuffix(assetName, ".zip") {
		return extractZip(archive)
	}
	return extractTarGz(archive)
}

// extractTarGz extracts the metro binary from a .tar.gz archive.
func extractTarGz(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decompressing: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading archive: %w", err)
		}
		name := filepath.Base(hdr.Name)
		if name == binaryName || name == binaryName+".exe" {
			bin, err := io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("extracting binary: %w", err)
			}
			return bin, nil
		}
	}
	return nil, fmt.Errorf("binary %q not found in archive", binaryName)
}

// extractZip extracts the metro binary from a .zip archive.
func extractZip(data []byte) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("reading zip: %w", err)
	}
	for _, f := range zr.File {
		name := filepath.Base(f.Name)
		if name == binaryName || name == binaryName+".exe" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("opening %s in zip: %w", f.Name, err)
			}
			defer rc.Close()
			bin, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("extracting binary: %w", err)
			}
			return bin, nil
		}
	}
	return nil, fmt.Errorf("binary %q not found in archive", binaryName)
}

// replaceBinary atomically replaces the current executable with the new binary.
func replaceBinary(newBin []byte) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding current binary: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("resolving binary path: %w", err)
	}

	// Get permissions from the current binary.
	info, err := os.Stat(execPath)
	if err != nil {
		return fmt.Errorf("reading binary permissions: %w", err)
	}

	// Write to a temp file in the same directory (ensures same filesystem for rename).
	dir := filepath.Dir(execPath)
	tmp, err := os.CreateTemp(dir, ".metro-update-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w\nTry running with sudo.", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(newBin); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing new binary: %w", err)
	}
	if err := tmp.Chmod(info.Mode()); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("setting permissions: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}

	// Atomic replace.
	if err := os.Rename(tmpPath, execPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replacing binary: %w\nTry running with sudo.", err)
	}

	return nil
}
