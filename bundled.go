package claude

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// Embed bundled CLI binaries for different platforms
// Note: These files are only embedded when they exist in the _bundled/ directory.
// Run 'go run scripts/download_cli.go' to download the CLI binaries.
//
//go:embed _bundled/*
var bundledCLI embed.FS

// getBundledCLIPath returns the path to the bundled CLI binary for the current platform.
// If the binary doesn't exist in the embedded filesystem, returns an empty string.
func getBundledCLIPath() (string, error) {
	// Determine the binary name for the current platform
	var binaryName string
	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			binaryName = "claude-darwin-amd64"
		case "arm64":
			binaryName = "claude-darwin-arm64"
		default:
			return "", nil // Unsupported architecture
		}
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			binaryName = "claude-linux-amd64"
		case "arm64":
			binaryName = "claude-linux-arm64"
		default:
			return "", nil // Unsupported architecture
		}
	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			binaryName = "claude-windows-amd64.exe"
		default:
			return "", nil // Unsupported architecture
		}
	default:
		return "", nil // Unsupported OS
	}

	embeddedPath := "_bundled/" + binaryName

	// Check if the file exists in the embedded filesystem
	if _, err := bundledCLI.Open(embeddedPath); err != nil {
		// Binary not embedded (possibly development mode)
		return "", nil
	}

	// Extract the binary to a temporary location
	tempDir := filepath.Join(os.TempDir(), "claude-agent-sdk-go")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	tempFile := filepath.Join(tempDir, binaryName)

	// Check if we already extracted this binary
	if info, err := os.Stat(tempFile); err == nil {
		// File exists, verify it's not corrupted by checking size
		embeddedFile, err := bundledCLI.Open(embeddedPath)
		if err == nil {
			defer embeddedFile.Close()
			if stat, err := embeddedFile.Stat(); err == nil && stat.Size() == info.Size() {
				// File exists and has correct size, use it
				return tempFile, nil
			}
		}
	}

	// Extract the binary from embedded filesystem
	embeddedFile, err := bundledCLI.Open(embeddedPath)
	if err != nil {
		return "", fmt.Errorf("failed to open embedded CLI: %w", err)
	}
	defer embeddedFile.Close()

	// Create the output file
	outFile, err := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer outFile.Close()

	// Copy the binary
	if _, err := io.Copy(outFile, embeddedFile); err != nil {
		return "", fmt.Errorf("failed to extract CLI binary: %w", err)
	}

	return tempFile, nil
}
