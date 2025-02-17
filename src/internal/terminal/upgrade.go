package terminal

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"file4you/version"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/inconshreveable/go-update"
)

type Upgrade struct {
	term *Terminal
}

func NewUpgrade(term *Terminal) *Upgrade {
	return &Upgrade{
		term: term,
	}
}

func (up *Upgrade) CheckForUpgrade() {
	if os.Getenv("DESKTOP_CLEANER_SKIP_UPGRADE") != "" {
		return
	}

	if version.Version == "development" {
		return
	}

	up.term.ToggleSpinner(true, "Checking for updates ...")
	defer up.term.ToggleSpinner(false, "")
	// TODO: Migrate to Desktop Cleaner's version URL
	latestVersionURL := "https://raw.githubusercontent.com/ZanzyTHEbar/DesktopCleaner/tree/main/cli-version.txt"
	resp, err := http.Get(latestVersionURL)
	if err != nil {
		log.Println("Error checking latest version:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return
	}

	versionStr := string(body)
	versionStr = strings.TrimSpace(versionStr)

	latestVersion, err := semver.NewVersion(versionStr)
	if err != nil {
		log.Println("Error parsing latest version:", err)
		return
	}

	currentVersion, err := semver.NewVersion(version.Version)
	if err != nil {
		log.Println("Error parsing current version:", err)
		return
	}

	if latestVersion.GreaterThan(currentVersion) {
		up.term.ToggleSpinner(false, "")
		slog.Info(fmt.Sprintf("A new version of DesktopCleaner is available: %s", ColorHiGreen.Bold(true).Render(versionStr)))
		slog.Info(fmt.Sprintf("Current version: %s\n", ColorHiCyan.Bold(true).Render(version.Version)))
		confirmed := up.term.ConfirmYesNo("Upgrade to the latest version?")

		if confirmed {
			up.term.ResumeSpinner()
			err := up.DoUpgrade(latestVersion.String())
			if err != nil {
				up.term.OutputErrorAndExit("Failed to upgrade: %v", err)
				return
			}
			up.term.ToggleSpinner(false, "")
			up.RestartDesktopCleaner()
		} else {
			fmt.Println("Note: set DESKTOP_CLEANER_SKIP_UPGRADE=1 to stop upgrade prompts")
		}
	}
}

func (up *Upgrade) DoUpgrade(version string) error {
	tag := fmt.Sprintf("cli/v%s", version)
	escapedTag := url.QueryEscape(tag)

	downloadURL := fmt.Sprintf("https://github.com/ZanzyTHEbar/DesktopCleaner/releases/download/%s/desktop_cleaner_%s_%s_%s.tar.gz", escapedTag, version, runtime.GOOS, runtime.GOARCH)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download the update: %w", err)
	}
	defer resp.Body.Close()

	// Create a temporary file to save the downloaded archive
	tempFile, err := os.CreateTemp("", "*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up file afterwards

	// Copy the response body to the temporary file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save the downloaded archive: %w", err)
	}

	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to seek in temporary file: %w", err)
	}

	// Now, extract the binary from the tempFile
	gzr, err := gzip.NewReader(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tarReader := tar.NewReader(gzr)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Check if the current file is the binary
		if header.Typeflag == tar.TypeReg && (header.Name == "desktop_cleaner" || header.Name == "desktop_cleaner.exe") {
			err = update.Apply(tarReader, update.Options{})
			if err != nil {
				if errors.Is(err, fs.ErrPermission) {
					return fmt.Errorf("failed to apply update due to permission error; please try running your command again with 'sudo': %w", err)
				}
				return fmt.Errorf("failed to apply update: %w", err)
			}
			break
		}
	}

	return nil
}

func (up *Upgrade) RestartDesktopCleaner() {
	exe, err := os.Executable()
	if err != nil {
		up.term.OutputErrorAndExit("Failed to determine executable path: %v", err)
	}

	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		up.term.OutputErrorAndExit("Failed to restart: %v", err)
	}

	err = cmd.Wait()

	// If the process exited with an error, exit with the same error code
	if exitErr, ok := err.(*exec.ExitError); ok {
		os.Exit(exitErr.ExitCode())
	} else if err != nil {
		up.term.OutputErrorAndExit("Failed to restart: %v", err)
	}

	os.Exit(0)
}
