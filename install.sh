#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored message
print_msg() {
    local color=$1
    shift
    echo -e "${color}$*${NC}"
}

# Print error and exit
error_exit() {
    print_msg "$RED" "Error: $*"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    local os arch
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    arch="$(uname -m)"

    case "$os" in
        linux|darwin)
            ;;
        *)
            error_exit "Unsupported OS: $os. Only Linux and macOS are supported."
            ;;
    esac

    case "$arch" in
        x86_64|amd64)
            arch="amd64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        *)
            error_exit "Unsupported architecture: $arch. Only amd64 and arm64 are supported."
            ;;
    esac

    echo "${os}-${arch}"
}

# Get latest release version
get_latest_version() {
    print_msg "$BLUE" "Fetching latest release version..."
    local version
    version=$(curl -s "https://api.github.com/repos/ll931217/claude-hud-enhanced/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$version" ]; then
        error_exit "Failed to fetch latest version from GitHub"
    fi

    echo "$version"
}

# Download and install binary
install_binary() {
    local version=$1
    local platform=$2
    local archive_name="claude-hud-${version}-${platform}.tar.gz"
    local download_url="https://github.com/ll931217/claude-hud-enhanced/releases/download/${version}/${archive_name}"

    print_msg "$BLUE" "Downloading claude-hud ${version} for ${platform}..."

    # Create temp directory
    local tmp_dir
    tmp_dir=$(mktemp -d)

    # Download archive
    if ! curl -fsSL "$download_url" -o "${tmp_dir}/${archive_name}"; then
        rm -rf "$tmp_dir"
        error_exit "Failed to download archive from $download_url"
    fi

    # Extract archive
    print_msg "$BLUE" "Extracting archive..."
    tar -xzf "${tmp_dir}/${archive_name}" -C "$tmp_dir"

    # Find the extracted binary
    local binary_name
    binary_name=$(find "$tmp_dir" -type f -name "claude-hud-*" | head -1 | xargs basename)

    # Make executable
    chmod +x "${tmp_dir}/${binary_name}"

    # Create .claude directory if it doesn't exist
    mkdir -p ~/.claude

    # Stop any running instances
    if [ -f ~/.claude/claude-hud ]; then
        print_msg "$YELLOW" "Stopping existing claude-hud instance..."
        pkill -f claude-hud 2>/dev/null || true
        sleep 1
    fi

    # Install binary
    cp "${tmp_dir}/${binary_name}" ~/.claude/claude-hud
    print_msg "$GREEN" "Binary installed to ~/.claude/claude-hud"

    # Cleanup
    rm -rf "$tmp_dir"
}

# Update Claude Code settings
update_settings() {
    local settings_file="$HOME/.claude/settings.json"

    if [ ! -f "$settings_file" ]; then
        print_msg "$YELLOW" "Claude Code settings file not found at $settings_file"
        print_msg "$YELLOW" "Please add the following to your settings.json:"
        echo ""
        echo '{'
        echo '  "statusLine": {'
        echo '    "command": "~/.claude/claude-hud",'
        echo '    "padding": 0,'
        echo '    "type": "command"'
        echo '  }'
        echo '}'
        return
    fi

    print_msg "$BLUE" "Updating Claude Code settings..."

    # Use jq to update settings if available, otherwise provide manual instructions
    if command -v jq &> /dev/null; then
        # Backup original settings
        cp "$settings_file" "${settings_file}.backup"

        # Update statusLine configuration
        if jq -e '.statusLine' "$settings_file" > /dev/null 2>&1; then
            # Update existing statusLine
            jq '.statusLine = {
                "command": "~/.claude/claude-hud",
                "padding": 0,
                "type": "command"
            }' "$settings_file" > "${settings_file}.tmp" && \
            mv "${settings_file}.tmp" "$settings_file"
        else
            # Add new statusLine section
            jq '.statusLine = {
                "command": "~/.claude/claude-hud",
                "padding": 0,
                "type": "command"
            }' "$settings_file" > "${settings_file}.tmp" && \
            mv "${settings_file}.tmp" "$settings_file"
        fi

        print_msg "$GREEN" "Settings updated successfully"
        print_msg "$YELLOW" "Backup saved to ${settings_file}.backup"
    else
        print_msg "$YELLOW" "jq not found. Please manually update your settings.json:"
        echo ""
        print_msg "$BLUE" "Add or update the statusLine section:"
        echo ''
        echo '  "statusLine": {'
        echo '    "command": "~/.claude/claude-hud",'
        echo '    "padding": 0,'
        echo '    "type": "command"'
        echo '  }'
    fi
}

# Print usage instructions
print_usage() {
    echo ""
    print_msg "$GREEN" "╔════════════════════════════════════════════════════════════╗"
    print_msg "$GREEN" "║        Claude HUD Enhanced installed successfully!        ║"
    print_msg "$GREEN" "╚════════════════════════════════════════════════════════════╝"
    echo ""
    print_msg "$BLUE" "The statusline will activate automatically in Claude Code."
    echo ""
    print_msg "$YELLOW" "Features:"
    echo "  • Color-coded context progress bar (green/yellow/red)"
    echo "  • Token breakdown at high context usage (≥85%)"
    echo "  • Session duration, model, and cost tracking"
    echo "  • Beads issue tracking integration"
    echo "  • Git status with branch and changes"
    echo "  • Workspace info (path, language, CPU, memory, disk)"
    echo ""
    print_msg "$YELLOW" "To uninstall:"
    echo "  1. Remove ~/.claude/claude-hud"
    echo "  2. Remove or update statusLine section in ~/.claude/settings.json"
    echo ""
    print_msg "$BLUE" "For more information, visit:"
    echo "  https://github.com/ll931217/claude-hud-enhanced"
    echo ""
}

# Main installation flow
main() {
    print_msg "$BLUE" "╔════════════════════════════════════════════════════════════╗"
    print_msg "$BLUE" "║          Claude HUD Enhanced - Install Script             ║"
    print_msg "$BLUE" "╚════════════════════════════════════════════════════════════╝"
    echo ""

    # Detect platform
    local platform
    platform=$(detect_platform)
    print_msg "$GREEN" "Detected platform: $platform"

    # Get latest version
    local version
    version=$(get_latest_version)
    print_msg "$GREEN" "Latest version: $version"

    # Install binary
    install_binary "$version" "$platform"

    # Update settings
    update_settings

    # Print usage
    print_usage
}

main "$@"
