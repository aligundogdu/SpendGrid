#!/bin/bash


set -e

REPO="aligundogdu/SpendGrid"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="spendgrid"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and Architecture
detect_platform() {
    OS=$(uname -s)
    ARCH=$(uname -m)
    
    case $OS in
        Darwin)
            case $ARCH in
                x86_64)
                    PLATFORM="darwin-amd64"
                    ;;
                arm64)
                    PLATFORM="darwin-arm64"
                    ;;
                *)
                    echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
                    exit 1
                    ;;
            esac
            ;;
        Linux)
            case $ARCH in
                x86_64)
                    PLATFORM="linux-amd64"
                    ;;
                amd64)
                    PLATFORM="linux-amd64"
                    ;;
                *)
                    echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
                    exit 1
                    ;;
            esac
            ;;
        *)
            echo -e "${RED}Error: Unsupported operating system: $OS${NC}"
            exit 1
            ;;
    esac
}

# Get latest version
get_latest_version() {
    LATEST_URL="https://api.github.com/repos/$REPO/releases/latest"
    VERSION=$(curl -s "$LATEST_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ]; then
        echo -e "${RED}Error: Could not determine latest version${NC}"
        exit 1
    fi
    
    echo "Latest version: $VERSION"
}

# Download binary
download_binary() {
    DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/${BINARY_NAME}-${PLATFORM}"
    TEMP_DIR=$(mktemp -d)
    TEMP_FILE="$TEMP_DIR/$BINARY_NAME"
    
    echo -e "${YELLOW}Downloading SpendGrid $VERSION for $PLATFORM...${NC}"
    
    if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_FILE"; then
        echo -e "${RED}Error: Failed to download binary${NC}"
        rm -rf "$TEMP_DIR"
        exit 1
    fi
    
    chmod +x "$TEMP_FILE"
}

# Install binary
install_binary() {
    echo -e "${YELLOW}Installing to $INSTALL_DIR...${NC}"
    
    # Check if we need sudo
    if [ ! -w "$INSTALL_DIR" ]; then
        if command -v sudo &> /dev/null; then
            sudo mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
        else
            echo -e "${RED}Error: Cannot write to $INSTALL_DIR and sudo is not available${NC}"
            echo "Please run as root or install sudo"
            rm -rf "$TEMP_DIR"
            exit 1
        fi
    else
        mv "$TEMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    rm -rf "$TEMP_DIR"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        INSTALLED_VERSION=$("$BINARY_NAME" version 2>/dev/null || echo "unknown")
        echo -e "${GREEN}✓ SpendGrid installed successfully!${NC}"
        echo -e "${GREEN}  Location: $(which $BINARY_NAME)${NC}"
        echo -e "${GREEN}  Version: $INSTALLED_VERSION${NC}"
    else
        echo -e "${RED}Error: Installation verification failed${NC}"
        echo "Please check that $INSTALL_DIR is in your PATH"
        exit 1
    fi
}

# Main
main() {
    echo "================================"
    echo "  SpendGrid Installer"
    echo "================================"
    echo ""
    
    detect_platform
    get_latest_version
    download_binary
    install_binary
    verify_installation
    
    echo ""
    echo -e "${GREEN}Installation complete!${NC}"
    echo ""
    echo "Quick start:"
    echo "  spendgrid init        # Initialize database"
    echo "  spendgrid --help      # Show help"
}

main