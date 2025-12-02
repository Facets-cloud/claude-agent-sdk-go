package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

const (
	cliVersion = "2.0.56"
	installURL = "https://claude.ai/install.sh"
)

func main() {
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Println("Claude Code CLI Download Script")
	fmt.Printf("Downloading CLI version: %s\n", cliVersion)
	fmt.Println("=" + string(make([]byte, 60)))

	bundledDir := filepath.Join("..", "_bundled")
	if err := os.MkdirAll(bundledDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating bundled directory: %v\n", err)
		os.Exit(1)
	}

	// For the current platform, download and copy the CLI
	currentPlatform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("\nDownloading CLI for current platform: %s\n", currentPlatform)

	if err := downloadForCurrentPlatform(bundledDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading CLI: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n" + string(make([]byte, 60)))
	fmt.Println("CLI download complete!")
	fmt.Println("=" + string(make([]byte, 60)))
	fmt.Println("\nNote: For a complete multi-platform build, you need to:")
	fmt.Println("1. Run this script on each target platform (macOS, Linux, Windows)")
	fmt.Println("2. Or use a CI/CD system to build binaries for all platforms")
	fmt.Println("3. Copy all binaries to the _bundled/ directory")
}

func downloadForCurrentPlatform(bundledDir string) error {
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
			return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
		}
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			binaryName = "claude-linux-amd64"
		case "arm64":
			binaryName = "claude-linux-arm64"
		default:
			return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
		}
	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			binaryName = "claude-windows-amd64.exe"
		default:
			return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
		}
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// Download the install script
	fmt.Println("Downloading install script...")
	resp, err := http.Get(installURL)
	if err != nil {
		return fmt.Errorf("failed to download install script: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download install script: HTTP %d", resp.StatusCode)
	}

	// Save script to temp file
	tempScript := filepath.Join(os.TempDir(), "claude-install.sh")
	f, err := os.Create(tempScript)
	if err != nil {
		return fmt.Errorf("failed to create temp script: %w", err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return fmt.Errorf("failed to save install script: %w", err)
	}
	f.Close()

	// Make script executable
	if err := os.Chmod(tempScript, 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// Run install script
	fmt.Printf("Installing CLI version %s...\n", cliVersion)
	cmd := exec.Command("bash", tempScript, cliVersion)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install CLI: %w", err)
	}

	// Wait a bit for installation to complete
	time.Sleep(2 * time.Second)

	// Find the installed CLI
	cliPath, err := findInstalledCLI()
	if err != nil {
		return fmt.Errorf("failed to find installed CLI: %w", err)
	}

	fmt.Printf("Found installed CLI at: %s\n", cliPath)

	// Copy to bundled directory
	targetPath := filepath.Join(bundledDir, binaryName)
	fmt.Printf("Copying to: %s\n", targetPath)

	src, err := os.Open(cliPath)
	if err != nil {
		return fmt.Errorf("failed to open source CLI: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create target file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy CLI: %w", err)
	}

	// Print size info
	info, err := os.Stat(targetPath)
	if err == nil {
		sizeMB := float64(info.Size()) / (1024 * 1024)
		fmt.Printf("Binary size: %.2f MB\n", sizeMB)
	}

	return nil
}

func findInstalledCLI() (string, error) {
	// Check common installation locations
	locations := []string{
		filepath.Join(os.Getenv("HOME"), ".local", "bin", "claude"),
		"/usr/local/bin/claude",
		filepath.Join(os.Getenv("HOME"), "node_modules", ".bin", "claude"),
	}

	for _, path := range locations {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Check PATH
	path, err := exec.LookPath("claude")
	if err == nil {
		return path, nil
	}

	return "", fmt.Errorf("could not find installed Claude CLI")
}
