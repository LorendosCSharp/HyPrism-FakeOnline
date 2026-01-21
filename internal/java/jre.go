package java

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"HyPrism/internal/env"
	"HyPrism/internal/util"
	"HyPrism/internal/util/download"
)

// JREPlatform represents a JRE download for a specific platform
type JREPlatform struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

// JREJSON represents the JRE configuration
type JREJSON struct {
	Version     string                            `json:"version"`
	DownloadURL map[string]map[string]JREPlatform `json:"download_url"`
}

const (
	// Use Adoptium Temurin JRE directly instead of referencing another launcher
	jreConfigURL = "https://raw.githubusercontent.com/yyyumeniku/HyPrism/main/jre.json"
	jreVersion   = "25"
)

// DownloadJRE downloads the Java Runtime Environment
func DownloadJRE(ctx context.Context, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	jreDir := env.GetJREDir()
	javaBin := getJavaBinaryName()
	javaPath := filepath.Join(jreDir, javaBin)

	// Check if JRE already exists
	if _, err := os.Stat(javaPath); err == nil {
		fmt.Println("Java Runtime already installed")
		if progressCallback != nil {
			progressCallback("jre", 100, "Java Runtime ready", "", "", 0, 0)
		}
		return nil
	}

	if progressCallback != nil {
		progressCallback("jre", 0, "Downloading Java Runtime...", "", "", 0, 0)
	}

	// Fetch JRE config
	jreConfig, err := fetchJREConfig(ctx)
	if err != nil {
		// Fall back to default Adoptium URL
		return downloadAdoptiumJRE(ctx, jreDir, progressCallback)
	}

	// Get platform-specific URL
	osName := runtime.GOOS
	arch := runtime.GOARCH
	
	if osName == "darwin" {
		osName = "macos"
	}
	
	if arch == "amd64" {
		arch = "x64"
	}

	platform, ok := jreConfig.DownloadURL[osName]
	if !ok {
		return downloadAdoptiumJRE(ctx, jreDir, progressCallback)
	}

	archConfig, ok := platform[arch]
	if !ok {
		return downloadAdoptiumJRE(ctx, jreDir, progressCallback)
	}

	// Download JRE
	archiveExt := ".tar.gz"
	if runtime.GOOS == "windows" {
		archiveExt = ".zip"
	}

	archivePath := filepath.Join(env.GetCacheDir(), "jre"+archiveExt)

	if err := download.DownloadWithProgress(archivePath, archConfig.URL, "jre", 0.8, progressCallback); err != nil {
		return fmt.Errorf("failed to download JRE: %w", err)
	}

	// Verify checksum
	if archConfig.SHA256 != "" {
		if progressCallback != nil {
			progressCallback("jre", 85, "Verifying checksum...", "", "", 0, 0)
		}
		if err := util.VerifySHA256(archivePath, archConfig.SHA256); err != nil {
			os.Remove(archivePath)
			return fmt.Errorf("JRE checksum verification failed: %w", err)
		}
	}

	// Extract
	if progressCallback != nil {
		progressCallback("jre", 90, "Extracting Java Runtime...", "", "", 0, 0)
	}

	if err := util.ExtractArchive(archivePath, jreDir); err != nil {
		return fmt.Errorf("failed to extract JRE: %w", err)
	}

	// Cleanup
	os.Remove(archivePath)

	// Find and move java binary to expected location
	if err := normalizeJREStructure(jreDir); err != nil {
		return fmt.Errorf("failed to normalize JRE structure: %w", err)
	}

	if progressCallback != nil {
		progressCallback("jre", 100, "Java Runtime installed", "", "", 0, 0)
	}

	return nil
}

func fetchJREConfig(ctx context.Context) (*JREJSON, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	
	req, err := http.NewRequestWithContext(ctx, "GET", jreConfigURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JRE config: %d", resp.StatusCode)
	}

	var config JREJSON
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func downloadAdoptiumJRE(ctx context.Context, jreDir string, progressCallback func(stage string, progress float64, message string, currentFile string, speed string, downloaded, total int64)) error {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	if arch == "amd64" {
		arch = "x64"
	}
	if arch == "arm64" {
		arch = "aarch64"
	}
	if osName == "darwin" {
		osName = "mac"
	}

	archiveType := "tar.gz"
	if osName == "windows" {
		archiveType = "zip"
	}

	url := fmt.Sprintf(
		"https://api.adoptium.net/v3/binary/latest/%s/ga/%s/%s/jre/hotspot/normal/eclipse?project=jdk",
		jreVersion, osName, arch,
	)

	archivePath := filepath.Join(env.GetCacheDir(), "jre."+archiveType)

	if err := download.DownloadWithProgress(archivePath, url, "jre", 0.8, progressCallback); err != nil {
		return fmt.Errorf("failed to download JRE from Adoptium: %w", err)
	}

	if progressCallback != nil {
		progressCallback("jre", 90, "Extracting Java Runtime...", "", "", 0, 0)
	}

	if err := util.ExtractArchive(archivePath, jreDir); err != nil {
		return fmt.Errorf("failed to extract JRE: %w", err)
	}

	os.Remove(archivePath)

	if err := normalizeJREStructure(jreDir); err != nil {
		return fmt.Errorf("failed to normalize JRE structure: %w", err)
	}

	if progressCallback != nil {
		progressCallback("jre", 100, "Java Runtime installed", "", "", 0, 0)
	}

	return nil
}

func normalizeJREStructure(jreDir string) error {
	// JRE archives often have a version directory, we need to move contents up
	entries, err := os.ReadDir(jreDir)
	if err != nil {
		return err
	}

	// If there's a single directory, move its contents up
	if len(entries) == 1 && entries[0].IsDir() {
		subDir := filepath.Join(jreDir, entries[0].Name())
		
		// On macOS, the structure is different
		if runtime.GOOS == "darwin" {
			contentsDir := filepath.Join(subDir, "Contents", "Home")
			if _, err := os.Stat(contentsDir); err == nil {
				subDir = contentsDir
			}
		}

		// Move all files from subDir to jreDir
		subEntries, err := os.ReadDir(subDir)
		if err != nil {
			return err
		}

		for _, entry := range subEntries {
			oldPath := filepath.Join(subDir, entry.Name())
			newPath := filepath.Join(jreDir, entry.Name())
			
			if err := os.Rename(oldPath, newPath); err != nil {
				// Try copy instead
				if entry.IsDir() {
					if err := util.CopyDir(oldPath, newPath); err != nil {
						return err
					}
				} else {
					if err := util.CopyFile(oldPath, newPath); err != nil {
						return err
					}
				}
			}
		}

		// Remove the now-empty directory
		os.RemoveAll(filepath.Join(jreDir, entries[0].Name()))
	}

	return nil
}

func getJavaBinaryName() string {
	if runtime.GOOS == "windows" {
		return filepath.Join("bin", "java.exe")
	}
	return filepath.Join("bin", "java")
}

// GetJavaExec returns the path to the Java executable
func GetJavaExec() (string, error) {
	jreDir := env.GetJREDir()
	javaBin := getJavaBinaryName()
	javaPath := filepath.Join(jreDir, javaBin)

	if _, err := os.Stat(javaPath); err != nil {
		return "", fmt.Errorf("Java not found at %s", javaPath)
	}

	// Make sure it's executable on Unix systems
	if runtime.GOOS != "windows" {
		os.Chmod(javaPath, 0755)
	}

	return javaPath, nil
}
