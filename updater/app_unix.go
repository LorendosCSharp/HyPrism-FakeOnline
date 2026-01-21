//go:build !windows

package updater

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Apply applies a launcher update on Unix systems and restarts the app
func Apply(tmp string) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create a shell script to replace the binary and restart
	scriptPath := filepath.Join(os.TempDir(), "hyprism-update.sh")
	
	var script string
	if runtime.GOOS == "darwin" {
		// For macOS - the update is a DMG file containing the .app bundle
		// Navigate up from Contents/MacOS/executable to the .app bundle
		appBundlePath := exe
		for i := 0; i < 3; i++ {
			appBundlePath = filepath.Dir(appBundlePath)
		}
		if !strings.HasSuffix(appBundlePath, ".app") {
			return fmt.Errorf("not running from an app bundle: %s", appBundlePath)
		}
		
		appName := filepath.Base(appBundlePath) // e.g., "HyPrism.app"
		appDir := filepath.Dir(appBundlePath)   // e.g., "/Applications"
		mountPoint := filepath.Join(os.TempDir(), "hyprism-dmg-mount")
		
		script = fmt.Sprintf(`#!/bin/bash
set -e

echo "Starting HyPrism update..."

# Wait for app to close
sleep 2

# Create mount point
mkdir -p "%s"

# Mount the DMG
echo "Mounting update package..."
hdiutil attach "%s" -mountpoint "%s" -nobrowse -quiet

# Find the .app in the mounted DMG
APP_IN_DMG=""
for item in "%s"/*.app; do
    if [ -d "$item" ]; then
        APP_IN_DMG="$item"
        break
    fi
done

if [ -z "$APP_IN_DMG" ]; then
    echo "Error: No .app found in update package"
    hdiutil detach "%s" -quiet 2>/dev/null || true
    exit 1
fi

echo "Found app: $APP_IN_DMG"

# Backup old app
echo "Backing up current version..."
rm -rf "%s.old" 2>/dev/null || true
mv "%s" "%s.old" 2>/dev/null || true

# Copy new app
echo "Installing new version..."
cp -R "$APP_IN_DMG" "%s/"

# Remove quarantine attribute (important for macOS security)
echo "Removing quarantine..."
xattr -cr "%s/%s" 2>/dev/null || true

# Set executable permissions
chmod +x "%s/%s/Contents/MacOS/"* 2>/dev/null || true

# Unmount DMG
echo "Cleaning up..."
hdiutil detach "%s" -quiet 2>/dev/null || true
rmdir "%s" 2>/dev/null || true

# Remove backup and temp files
rm -rf "%s.old" 2>/dev/null || true
rm -f "%s" 2>/dev/null || true

echo "Update complete! Launching HyPrism..."

# Restart the application
open "%s/%s"

# Clean up this script
rm -f "%s"
`, mountPoint, tmp, mountPoint, mountPoint, mountPoint,
   appBundlePath, appBundlePath, appBundlePath,
   appDir,
   appDir, appName,
   appDir, appName,
   mountPoint, mountPoint,
   appBundlePath, tmp,
   appDir, appName,
   scriptPath)
	} else {
		// Linux
		script = fmt.Sprintf(`#!/bin/bash
sleep 1
mv "%s" "%s.old" 2>/dev/null
cp "%s" "%s"
chmod +x "%s"
rm -f "%s.old"
rm -f "%s"
# Restart the application
"%s" &
rm -f "%s"
`, exe, exe, tmp, exe, exe, exe, tmp, exe, scriptPath)
	}

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("failed to create update script: %w", err)
	}

	cmd := exec.Command("/bin/bash", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start update script: %w", err)
	}

	os.Exit(0)
	return nil
}
