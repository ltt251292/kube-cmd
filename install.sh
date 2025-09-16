#!/bin/bash

# install.sh - Script cài đặt Kube Tools từ GitHub
# Usage: curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash
# hoặc: ./install.sh [options]

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
INSTALL_DIR="/usr/local/bin"
BUILD_ONLY=false
FORCE=false
QUIET=false
REPO_URL="https://github.com/ltt251292/kube-cmd.git"  # TODO: Update với actual repo URL
BRANCH="main"
TEMP_DIR=""
CLEANUP_TEMP=false

# Script information
SCRIPT_NAME="$(basename "$0")"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Function để hiển thị help
show_help() {
    echo "Kube Tools Installer"
    echo ""
    echo "Usage: $SCRIPT_NAME [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help          Hiển thị help này"
    echo "  -d, --dir DIR       Thư mục cài đặt (default: $INSTALL_DIR)"
    echo "  -b, --build-only    Chỉ build, không cài đặt"
    echo "  -f, --force         Ghi đè files đã tồn tại"
    echo "  -q, --quiet         Chế độ quiet (ít output)"
    echo "  --uninstall         Gỡ bỏ tất cả tools"
    echo "  --repo URL          GitHub repository URL (default: $REPO_URL)"
    echo "  --branch BRANCH     Git branch để clone (default: $BRANCH)"
    echo ""
    echo "Examples:"
    echo "  # Cài đặt từ internet (khuyến nghị)"
    echo "  curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash"
    echo ""
    echo "  # Cài đặt local"
    echo "  $SCRIPT_NAME                    # Cài đặt tất cả tools"
    echo "  $SCRIPT_NAME --build-only       # Chỉ build tools"
    echo "  $SCRIPT_NAME --dir ~/bin        # Cài đặt vào ~/bin"
    echo "  $SCRIPT_NAME --uninstall        # Gỡ bỏ tools"
}

# Function để log messages
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

# Function để kiểm tra requirements
check_requirements() {
    log "Kiểm tra requirements..."
    
    # Kiểm tra Git (cần để clone repo)
    if ! command -v git &> /dev/null; then
        log_error "Git không được cài đặt. Vui lòng cài đặt Git trước."
        exit 1
    fi
    
    log "Git ✓"
    
    # Kiểm tra Go
    if ! command -v go &> /dev/null; then
        log_error "Go không được cài đặt. Vui lòng cài đặt Go 1.21+ trước."
        exit 1
    fi
    
    # Kiểm tra Go version
    go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    go_major=$(echo $go_version | cut -d. -f1)
    go_minor=$(echo $go_version | cut -d. -f2)
    
    if [[ $go_major -lt 1 ]] || [[ $go_major -eq 1 && $go_minor -lt 21 ]]; then
        log_error "Go version $go_version không được hỗ trợ. Cần Go 1.21+."
        exit 1
    fi
    
    log "Go version $go_version ✓"
    
    # Kiểm tra make
    if ! command -v make &> /dev/null; then
        log_error "Make không được cài đặt."
        exit 1
    fi
    
    log "Make ✓"
    
    # Kiểm tra quyền ghi vào install directory
    if [[ "$BUILD_ONLY" == "false" ]]; then
        if [[ ! -w "$INSTALL_DIR" ]] && [[ "$EUID" -ne 0 ]]; then
            log_warning "Không có quyền ghi vào $INSTALL_DIR. Sẽ cần sudo."
        fi
    fi
}

# Function để clone repository
clone_repo() {
    log "Cloning repository từ $REPO_URL..."
    
    # Tạo temp directory
    TEMP_DIR=$(mktemp -d)
    CLEANUP_TEMP=true
    
    # Clone repository
    if ! git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$TEMP_DIR"; then
        log_error "Failed to clone repository"
        cleanup
        exit 1
    fi
    
    # Chuyển vào thư mục source
    cd "$TEMP_DIR"
    SCRIPT_DIR="$TEMP_DIR"
    
    log_success "Repository cloned successfully"
}

# Function để cleanup
cleanup() {
    if [[ "$CLEANUP_TEMP" == "true" ]] && [[ -n "$TEMP_DIR" ]] && [[ -d "$TEMP_DIR" ]]; then
        log "Cleaning up temporary directory..."
        rm -rf "$TEMP_DIR"
    fi
}

# Function để build tools
build_tools() {
    log "Building kube tools..."
    
    cd "$SCRIPT_DIR"
    
    # Clean trước khi build
    if ! make clean; then
        log_error "Failed to clean build artifacts"
        exit 1
    fi
    
    # Tidy dependencies
    if ! make tidy; then
        log_error "Failed to tidy dependencies"
        exit 1
    fi
    
    # Build tất cả tools
    if ! make build-all; then
        log_error "Failed to build tools"
        exit 1
    fi
    
    log_success "Build completed successfully"
}

# Function để cài đặt tools
install_tools() {
    log "Installing tools to $INSTALL_DIR..."
    
    cd "$SCRIPT_DIR"
    
    # Danh sách tools
    TOOLS=("kube" "kube-pods" "kube-services" "kube-switch-context" "kube-switch-namespace" "kube-logs" "kube-port-forward" "kube-exec")
    
    for tool in "${TOOLS[@]}"; do
        if [[ ! -f "$tool" ]]; then
            log_error "Binary $tool không tồn tại. Vui lòng build trước."
            exit 1
        fi
        
        target="$INSTALL_DIR/$tool"
        
        # Kiểm tra file đã tồn tại
        if [[ -f "$target" ]] && [[ "$FORCE" == "false" ]]; then
            log_warning "$tool đã tồn tại trong $INSTALL_DIR. Sử dụng --force để ghi đè."
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

# Function để uninstall tools
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

# Function để verify installation
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
    
    # Kiểm tra xem đã ở trong repo chưa
    if [[ -f "go.mod" ]] && [[ -f "Makefile" ]] && [[ -d "tools" ]]; then
        log "Đã ở trong thư mục source code, bỏ qua clone."
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
    
    # Clone repository nếu cần
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
