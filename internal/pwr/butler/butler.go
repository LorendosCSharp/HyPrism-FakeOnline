package butler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"HyPrism/internal/env"
	"HyPrism/internal/util"
	"HyPrism/internal/util/download"
)

const (
	butlerVersion = "15.21.0"
	brothURL      = "https://broth.itch.zone/butler/%s-%s/LATEST/archive/default"
)

// InstallButler installs the Butler tool
func InstallButler(ctx context.Context, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) (string, error) {
	butlerDir := env.GetButlerDir()
	butlerPath, err := GetButlerPath()
	if err == nil {
		if _, statErr := os.Stat(butlerPath); statErr == nil {
			fmt.Println("Butler already installed")
			if progressCallback != nil {
				progressCallback("butler", 100, "Butler ready", "", "", 0, 0)
			}
			return butlerPath, nil
		}
	}

	if err := os.MkdirAll(butlerDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create butler directory: %w", err)
	}

	if progressCallback != nil {
		progressCallback("butler", 0, "Downloading Butler...", "", "", 0, 0)
	}

	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Butler only provides darwin-amd64 (no arm64), so on macOS we always use amd64
	// which runs through Rosetta 2 on Apple Silicon
	if osName == "darwin" {
		arch = "amd64"
	}

	url := fmt.Sprintf(brothURL, osName, arch)
	fmt.Printf("Butler download URL: %s\n", url)
	archivePath := filepath.Join(env.GetCacheDir(), "butler.zip")

	if err := download.DownloadWithProgress(archivePath, url, "butler", 0.8, progressCallback); err != nil {
		return "", fmt.Errorf("failed to download butler: %w", err)
	}

	if progressCallback != nil {
		progressCallback("butler", 90, "Extracting Butler...", "", "", 0, 0)
	}

	if err := util.ExtractZip(archivePath, butlerDir); err != nil {
		return "", fmt.Errorf("failed to extract butler: %w", err)
	}

	os.Remove(archivePath)

	butlerPath, err = GetButlerPath()
	if err != nil {
		return "", err
	}

	// Make executable on Unix
	if runtime.GOOS != "windows" {
		os.Chmod(butlerPath, 0755)
	}

	// Verify butler works
	cmd := exec.Command(butlerPath, "version")
	util.HideConsoleWindow(cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("butler verification failed: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("Butler installed: %s\n", string(output))

	if progressCallback != nil {
		progressCallback("butler", 100, "Butler installed", "", "", 0, 0)
	}

	return butlerPath, nil
}

// GetButlerPath returns the path to the Butler executable
func GetButlerPath() (string, error) {
	butlerDir := env.GetButlerDir()
	
	name := "butler"
	if runtime.GOOS == "windows" {
		name = "butler.exe"
	}

	path := filepath.Join(butlerDir, name)
	return path, nil
}
