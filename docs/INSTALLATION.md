# Installation Guide

This guide covers installation of Claude HUD Enhanced on various platforms.

## Requirements

- Go 1.25.5 or later
- Git (for building from source)
- Unix-like OS (Linux, macOS) - Windows users should use WSL2

## Installation Methods

### Method 1: Pre-built Binaries (Recommended)

Download pre-built binaries from the [releases page](https://github.com/ll931217/claude-hud-enhanced/releases).

#### Linux

```bash
# Download and extract
wget https://github.com/ll931217/claude-hud-enhanced/releases/latest/download/claude-hud-linux-amd64.tar.gz
tar -xzf claude-hud-linux-amd64.tar.gz

# Install to /usr/local/bin
sudo cp claude-hud /usr/local/bin/
sudo chmod +x /usr/local/bin/claude-hud

# Verify installation
claude-hud --version
```

#### macOS (Apple Silicon)

```bash
# Download and extract
curl -L -o claude-hud-darwin-arm64.tar.gz https://github.com/ll931217/claude-hud-enhanced/releases/latest/download/claude-hud-darwin-arm64.tar.gz
tar -xzf claude-hud-darwin-arm64.tar.gz

# Install to /usr/local/bin
sudo cp claude-hud /usr/local/bin/
sudo chmod +x /usr/local/bin/claude-hud

# Verify installation
claude-hud --version
```

#### macOS (Intel)

```bash
# Download and extract
curl -L -o claude-hud-darwin-amd64.tar.gz https://github.com/ll931217/claude-hud-enhanced/releases/latest/download/claude-hud-darwin-amd64.tar.gz
tar -xzf claude-hud-darwin-amd64.tar.gz

# Install to /usr/local/bin
sudo cp claude-hud /usr/local/bin/
sudo chmod +x /usr/local/bin/claude-hud

# Verify installation
claude-hud --version
```

### Method 2: Build from Source

#### Prerequisites

Install Go 1.25.5 or later:

**Linux (Debian/Ubuntu):**
```bash
sudo apt-get update
sudo apt-get install golang-go
```

**Linux (Fedora/RHEL):**
```bash
sudo dnf install golang
```

**macOS (using Homebrew):**
```bash
brew install go
```

#### Building

```bash
# Clone the repository
git clone https://github.com/ll931217/claude-hud-enhanced.git
cd claude-hud-enhanced

# Build for your current platform
make build

# The binary will be in bin/claude-hud
sudo cp bin/claude-hud /usr/local/bin/

# Verify installation
claude-hud --version
```

#### Cross-Compilation

Build for other platforms:

```bash
# Linux AMD64
make release GOOS=linux GOARCH=amd64

# macOS ARM64 (Apple Silicon)
make release GOOS=darwin GOARCH=arm64

# macOS AMD64 (Intel)
make release GOOS=darwin GOARCH=amd64

# Build all platforms at once
make release-all
```

Binaries will be in the `release/` directory.

### Method 3: Using Go Install

If you have Go installed, you can install directly:

```bash
go install github.com/ll931217/claude-hud-enhanced/cmd/claude-hud@latest
```

The binary will be installed to `$GOPATH/bin` (usually `~/go/bin/`).

Make sure `$GOPATH/bin` is in your PATH:

```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH=$PATH:$(go env GOPATH)/bin
```

## Configuration

After installation, create the configuration directory and file:

```bash
# Create config directory
mkdir -p ~/.config/claude-hud

# Copy example config
cp config.example.yaml ~/.config/claude-hud/config.yaml

# Edit configuration
nano ~/.config/claude-hud/config.yaml
```

## Verification

Verify your installation:

```bash
# Show version
claude-hud --version

# Show detailed build information
claude-hud --build-info

# Run the statusline
claude-hud
```

## Updating

### Using Pre-built Binaries

```bash
# Download new release
wget https://github.com/ll931217/claude-hud-enhanced/releases/latest/download/claude-hud-linux-amd64.tar.gz
tar -xzf claude-hud-linux-amd64.tar.gz

# Replace old binary
sudo cp claude-hud /usr/local/bin/

# Verify
claude-hud --version
```

### Building from Source

```bash
cd claude-hud-enhanced
git pull origin main
make build
sudo cp bin/claude-hud /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/ll931217/claude-hud-enhanced/cmd/claude-hud@latest
```

## Uninstallation

### Remove Binary

```bash
sudo rm /usr/local/bin/claude-hud
```

### Remove Configuration

```bash
# Remove config directory
rm -rf ~/.config/claude-hud
```

## Troubleshooting

### Permission Denied

If you get a "permission denied" error:

```bash
# Make the binary executable
chmod +x claude-hud
```

### Command Not Found

If `claude-hud` command is not found:

1. Verify the binary is in your PATH:
   ```bash
   echo $PATH
   which claude-hud
   ```

2. If not found, add the installation directory to PATH:
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   export PATH=$PATH:/usr/local/bin
   ```

3. Reload your shell configuration:
   ```bash
   source ~/.bashrc  # or source ~/.zshrc
   ```

### Go Version Too Old

If you get an error about Go version:

```bash
# Check Go version
go version

# Update Go
# macOS
brew upgrade go

# Linux (Debian/Ubuntu)
sudo apt-get install golang-go-1.25

# Or download from https://golang.org/dl/
```

## Next Steps

- [Configuration Guide](CONFIGURATION.md) - How to configure Claude HUD Enhanced
- [Usage Guide](USAGE.md) - How to use Claude HUD Enhanced
- [Examples](../examples/) - Example configurations and usage
