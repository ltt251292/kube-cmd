#!/bin/bash

# install.sh - Script to install Kube Tools from GitHub
# Usage: curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash
# or: ./install.sh [options]

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values (can be overridden by environment variables)
INSTALL_DIR="${KUBE_INSTALL_DIR:-/usr/local/bin}"
BUILD_ONLY="${KUBE_BUILD_ONLY:-false}"
FORCE="${KUBE_FORCE:-false}"
QUIET="${KUBE_QUIET:-false}"
REPO_URL="${KUBE_REPO_URL:-https://github.com/ltt251292/kube-cmd.git}"
BRANCH="${KUBE_BRANCH:-main}"
TEMP_DIR=""
CLEANUP_TEMP=false

# Script information
SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Function to show help
show_help() {
    echo "Kube Tools Installer"
    echo ""
    echo "Usage: $SCRIPT_NAME [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help          Show this help"
    echo "  -d, --dir DIR       Installation directory (default: $INSTALL_DIR)"
    echo "  -b, --build-only    Build only, do not install"
    echo "  -f, --force         Overwrite existing files"
    echo "  -q, --quiet         Quiet mode (less output)"
    echo "  --uninstall         Uninstall all tools"
    echo "  --repo URL          GitHub repository URL (default: $REPO_URL)"
    echo "  --branch BRANCH     Git branch to clone (default: $BRANCH)"
    echo ""
    echo "Environment Variables:"
    echo "  KUBE_INSTALL_DIR    Installation directory (override --dir)"
    echo "  KUBE_BUILD_ONLY     Build only: true/false (override --build-only)"
    echo "  KUBE_FORCE          Overwrite files: true/false (override --force)"
    echo "  KUBE_QUIET          Quiet mode: true/false (override --quiet)"
    echo "  KUBE_REPO_URL       Repository URL (override --repo)"
    echo "  KUBE_BRANCH         Git branch (override --branch)"
    echo ""
    echo "Examples:"
    echo "  # Install from internet (recommended)"
    echo "  curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash"
    echo ""
    echo "  # With environment variables"
    echo "  KUBE_INSTALL_DIR=~/bin curl -fsSL ... | bash"
    echo "  KUBE_BUILD_ONLY=true KUBE_QUIET=true curl -fsSL ... | bash"
    echo ""
    echo "  # Local installation"
    echo "  $SCRIPT_NAME                    # Install all tools"
    echo "  $SCRIPT_NAME --build-only       # Build tools only"
    echo "  $SCRIPT_NAME --dir ~/bin        # Install to ~/bin"
    echo "  $SCRIPT_NAME --uninstall        # Uninstall tools"
}

# Function to log messages
log() {
    if [[ "$QUIET" == "false" ]]; then
        echo -e "${BLUE}[INFO]${NC} $1"
    fi
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Function to check requirements
check_requirements() {
    log "Checking requirements..."
    
    # Check Git (needed to clone repo)
    if ! command -v git &> /dev/null; then
        log_error "Git is not installed. Please install Git first."
        exit 1
    fi
    
    log "Git ✓"
    
    # Check Go
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21+ first."
        exit 1
    fi
    
    # Check Go version
    go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    go_major=$(echo $go_version | cut -d. -f1)
    go_minor=$(echo $go_version | cut -d. -f2)
    
    if [[ $go_major -lt 1 ]] || [[ $go_major -eq 1 && $go_minor -lt 21 ]]; then
        log_error "Go version $go_version is not supported. Need Go 1.21+."
        exit 1
    fi
    
    log "Go version $go_version ✓"
    
    # Check make
    if ! command -v make &> /dev/null; then
        log_error "Make is not installed."
        exit 1
    fi
    
    log "Make ✓"
    
    # Check write permission to install directory
    if [[ "$BUILD_ONLY" == "false" ]]; then
        if [[ ! -w "$INSTALL_DIR" ]] && [[ "$EUID" -ne 0 ]]; then
            log_warning "No write permission to $INSTALL_DIR. Will need sudo."
        fi
    fi
}

# Function to clone repository
clone_repo() {
    log "Cloning repository from $REPO_URL..."
    
    # Create temp directory
    TEMP_DIR=$(mktemp -d)
    CLEANUP_TEMP=true
    
    # Clone repository
    if ! git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$TEMP_DIR"; then
        log_error "Failed to clone repository"
        cleanup
        exit 1
    fi
    
    # Change to source directory
    cd "$TEMP_DIR"
    SCRIPT_DIR="$TEMP_DIR"
    
    log_success "Repository cloned successfully"
}

# Function to cleanup
cleanup() {
    if [[ "$CLEANUP_TEMP" == "true" ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        log "Cleaning up temporary directory..."
        rm -rf "$TEMP_DIR"
    fi
}

# Function to build tools
build_tools() {
    log "Building kube tools..."
    
    cd "$SCRIPT_DIR"
    
    # Clean before build
    if ! make clean; then
        log_error "Failed to clean build artifacts"
        exit 1
    fi
    
    # Tidy dependencies
    if ! make tidy; then
        log_error "Failed to tidy dependencies"
        exit 1
    fi
    
    # Build all tools
    if ! make build-all; then
        log_error "Failed to build tools"
        exit 1
    fi
    
    log_success "Build completed successfully"
}

# Function to install tools
install_tools() {
    log "Installing tools to $INSTALL_DIR..."
    
    cd "$SCRIPT_DIR"
    
    # List of tools
    TOOLS=("kube" "kube-pods" "kube-services" "kube-switch-context" "kube-switch-namespace" "kube-logs" "kube-port-forward" "kube-exec")
    
    for tool in "${TOOLS[@]}"; do
        if [[ ! -f "$tool" ]]; then
            log_error "Binary $tool does not exist. Please build first."
            exit 1
        fi
        
        target="$INSTALL_DIR/$tool"
        
        # Check if file already exists
        if [[ -f "$target" ]] && [[ "$FORCE" == "false" ]]; then
            log_warning "$tool already exists in $INSTALL_DIR. Use --force to overwrite."
            continue
        fi
        
        # Copy file
        if [[ -w "$INSTALL_DIR" ]] || [[ "$EUID" -eq 0 ]]; then
            cp "$tool" "$target"
            chmod +x "$target"
        else
            sudo cp "$tool" "$target"
            sudo chmod +x "$target"
        fi
        
        log "Installed $tool"
    done
    
    log_success "All tools installed successfully to $INSTALL_DIR"
}

# Function to uninstall tools
uninstall_tools() {
    log "Uninstalling kube tools from $INSTALL_DIR..."
    
    TOOLS=("kube" "kube-pods" "kube-services" "kube-switch-context" "kube-switch-namespace" "kube-logs" "kube-port-forward" "kube-exec")
    
    for tool in "${TOOLS[@]}"; do
        target="$INSTALL_DIR/$tool"
        
        if [[ -f "$target" ]]; then
            if [[ -w "$INSTALL_DIR" ]] || [[ "$EUID" -eq 0 ]]; then
                rm -f "$target"
            else
                sudo rm -f "$target"
            fi
            log "Removed $tool"
        else
            log "$tool not found (already removed?)"
        fi
    done
    
    log_success "All tools uninstalled successfully"
}

# Function to verify installation
verify_installation() {
    log "Verifying installation..."
    
    TOOLS=("kube" "kube-pods" "kube-services" "kube-switch-context" "kube-switch-namespace" "kube-logs" "kube-port-forward" "kube-exec")
    
    missing_tools=()
    for tool in "${TOOLS[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [[ ${#missing_tools[@]} -eq 0 ]]; then
        log_success "All tools are properly installed and available in PATH"
        echo ""
        echo "Try running: kube"
    else
        log_warning "Some tools are not available in PATH: ${missing_tools[*]}"
        log_warning "Make sure $INSTALL_DIR is in your PATH"
    fi
}

# Main function
main() {
    local uninstall=false
    local need_clone=true
    
    # Convert environment variables from string to boolean
    if [[ "$BUILD_ONLY" == "true" ]] || [[ "$BUILD_ONLY" == "1" ]]; then
        BUILD_ONLY=true
    else
        BUILD_ONLY=false
    fi
    
    if [[ "$FORCE" == "true" ]] || [[ "$FORCE" == "1" ]]; then
        FORCE=true
    else
        FORCE=false
    fi
    
    if [[ "$QUIET" == "true" ]] || [[ "$QUIET" == "1" ]]; then
        QUIET=true
    else
        QUIET=false
    fi
    
    # Check if already in repo
    if [[ -f "go.mod" ]] && [[ -f "Makefile" ]] && [[ -d "tools" ]]; then
        log "Already in source code directory, skipping clone."
        need_clone=false
    fi
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -d|--dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            -b|--build-only)
                BUILD_ONLY=true
                shift
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -q|--quiet)
                QUIET=true
                shift
                ;;
            --uninstall)
                uninstall=true
                shift
                ;;
            --repo)
                REPO_URL="$2"
                shift 2
                ;;
            --branch)
                BRANCH="$2"
                shift 2
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # Header
    if [[ "$QUIET" == "false" ]]; then
        echo "================================"
        echo "    Kube Tools Installer"
        echo "================================"
        echo ""
    fi
    
    # Uninstall mode
    if [[ "$uninstall" == "true" ]]; then
        uninstall_tools
        exit 0
    fi
    
    # Normal installation flow
    check_requirements
    
    # Clone repository if needed
    if [[ "$need_clone" == "true" ]]; then
        clone_repo
        # Set cleanup trap
        trap cleanup EXIT
    fi
    
    build_tools
    
    if [[ "$BUILD_ONLY" == "false" ]]; then
        install_tools
        verify_installation
    else
        log_success "Build completed. Binaries are ready in $SCRIPT_DIR"
    fi
    
    echo ""
    log_success "Installation completed!"
    
    if [[ "$BUILD_ONLY" == "false" ]]; then
        echo ""
        echo "Next steps:"
        echo "  1. Make sure $INSTALL_DIR is in your PATH"
        echo "  2. Run 'kube' to see available tools"
        echo "  3. Try 'kube-pods --help' for help on individual tools"
    fi
}

# Run main function
main "$@"
