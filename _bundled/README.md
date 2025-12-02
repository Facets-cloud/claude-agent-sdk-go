# Bundled Claude Code CLI Binaries

This directory contains bundled Claude Code CLI binaries for different platforms.

## Supported Platforms

- `claude-darwin-amd64` - macOS Intel (x86_64)
- `claude-darwin-arm64` - macOS Apple Silicon (ARM64)
- `claude-linux-amd64` - Linux AMD64
- `claude-linux-arm64` - Linux ARM64
- `claude-windows-amd64.exe` - Windows AMD64

## Building

To download and bundle CLI binaries for distribution, run:

```bash
go run scripts/download_cli.go
```

This will download the CLI version specified in `cli_version.go` (currently v2.0.56) for all supported platforms.

## Usage

The Go SDK automatically uses bundled CLI binaries when no system installation is found. The `findCLI()` function in `transport_subprocess.go` checks for bundled binaries before falling back to system paths.

## Development

During development, you can test with a local CLI installation. The bundled binaries are only used as a fallback when no other installation is found.
