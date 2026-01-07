#!/bin/bash
set -e

VERSION=${1:-v0.1.0}

echo "Creating release $VERSION..."

# Run full test suite
echo "Running tests..."
make test

# Build all release binaries and archives
echo "Building releases..."
make release-all
make archives

# Create git tag
echo "Creating git tag..."
git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"

# Create GitHub release
echo "Creating GitHub release..."
gh release create "$VERSION" ./release/* \
  --title "$VERSION" \
  --notes "Release $VERSION of Claude HUD Enhanced

## Changes
- Claude Code integration with transcript parsing
- Beads & Worktrunk integration
- Git status visualization
- System monitoring (CPU, RAM, Disk)
- Session info with todo tracking and cost calculation
- Catppuccin Mocha theme
- Nerd Font icons with ASCII fallback
- Performance optimized with <50ms render latency

## Installation

\`\`\`bash
# Linux AMD64
wget https://github.com/ll931217/claude-hud-enhanced/releases/download/$VERSION/claude-hud-$VERSION-linux-amd64.tar.gz
tar -xzf claude-hud-$VERSION-linux-amd64.tar.gz
sudo cp claude-hud /usr/local/bin/
\`\`\`

See [INSTALLATION.md](https://github.com/ll931217/claude-hud-enhanced/blob/main/docs/INSTALLATION.md) for more details."
