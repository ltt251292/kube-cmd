# Kube Tools

A collection of command-line tools written in Go to work with Kubernetes clusters conveniently and efficiently.

Instead of a single tool, kube provides multiple specialized tools in the `kube-*` format, similar to kubectl plugins.

## Tools

- 🚀 **kube-pods**: List pods with clean output (colored status, IP, node, image versions)
- 🔧 **kube-services**: List services
- 🔄 **kube-switch-context**: Quickly switch between kube contexts
- 📁 **kube-switch-namespace**: Switch namespace in current context
- 📋 **kube-logs**: Tail logs with multiple options
- 🔌 **kube-port-forward**: Port-forward to pods or services
- 💻 **kube-exec**: Execute commands inside containers
- 📦 **kube-deploy**: Update Deployment image and wait for rollout (or list deployments)
- 🔁 **kube-rollout**: Restart or show rollout status of a Deployment

## Installation

### 🚀 Quick install from Internet (Recommended)

```bash
# Single command install
curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash

# Or download the script and run
wget https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh
chmod +x install.sh
./install.sh
```

### ⚙️ Installation options

```bash
# Install to a different directory
curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash -s -- --dir ~/bin

# Build only, do not install
curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash -s -- --build-only

# Uninstall all tools
curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash -s -- --uninstall
```

### 🌍 Environment Variables

You can override settings via environment variables:

```bash
# Install to a different directory
KUBE_INSTALL_DIR=~/bin curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash

# Build only, quiet mode
KUBE_BUILD_ONLY=true KUBE_QUIET=true curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash

# Use a different fork
KUBE_REPO_URL=https://github.com/your-fork/kube-cmd.git KUBE_BRANCH=develop curl -fsSL ... | bash

# Force overwrite existing files
KUBE_FORCE=true curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash
```

### 🛠️ Build from source (Developers)

```bash
git clone https://github.com/ltt251292/kube-cmd.git
cd kube

# Automated install
./install.sh

# Or build manually
make build-all
sudo make install-all
```

### Requirements

**To use:**
- Kubectl configured
- Access to a Kubernetes cluster

**To build from source:**
- Go 1.21+
- Git
- Make

### OS-specific dependency installation

To build from source you need: Git, Go (>= 1.21), Make. Below are quick setup instructions per operating system.

#### macOS (Homebrew)

```bash
# Install Homebrew if missing (see https://brew.sh)
# /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

brew update
brew install git go make

# (Optional) Install kubectl if needed
brew install kubectl

# Verify versions
git --version
go version   # ensure >= 1.21
make --version
```

Note: If not using Homebrew, you can install Xcode Command Line Tools to get `git` and `make`:

```bash
xcode-select --install
```

#### Debian/Ubuntu

```bash
sudo apt-get update -y
sudo apt-get install -y git make

# Install Go >= 1.21:
# Option 1 (recommended): download from the official site for the latest version
GO_VERSION=1.22.6
curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz -o /tmp/go.tgz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tgz
echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Option 2 (via apt - may not be the newest):
# sudo apt-get install -y golang

# (Optional) kubectl
sudo apt-get install -y kubectl || true

# Verify
git --version
go version
make --version
```

#### RHEL/CentOS/Fedora

```bash
# Git and make
sudo dnf install -y git make || sudo yum install -y git make

# Go >= 1.21 (recommended: download from official site)
GO_VERSION=1.22.6
curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz -o /tmp/go.tgz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tgz
echo 'export PATH="/usr/local/go/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# (Optional) kubectl
sudo dnf install -y kubectl || sudo yum install -y kubectl || true

# Verify
git --version
go version
make --version
```

#### Arch Linux

```bash
sudo pacman -Sy --noconfirm git go base-devel

# (Optional) kubectl
sudo pacman -Sy --noconfirm kubectl || true

# Verify
git --version
go version
make --version
```

#### openSUSE

```bash
sudo zypper refresh
sudo zypper install -y git make go

# (Optional) kubectl
sudo zypper install -y kubectl || true

# Verify
git --version
go version
make --version
```

#### Windows (WSL recommended)

We recommend using WSL (Ubuntu) and following the Debian/Ubuntu instructions above. If using native Windows, please install:

- Git for Windows
- Go 1.21+ (installer from go.dev)
- Make (GNUWin32 make or `choco install make`)

After installation, ensure `go`, `git`, and `make` are in your PATH and meet the required versions.

## Usage

### Show tools

```bash
# Show all available tools
kube
```

### View resources

```bash
# List pods
kube-pods
kube-pods -A  # All namespaces

# List services
kube-services
kube-services -n my-namespace
```

### Switch context and namespace

```bash
# List contexts
kube-switch-context

# Switch to another context
kube-switch-context my-context

# Show current namespace
kube-switch-namespace

# Switch to another namespace
kube-switch-namespace my-namespace
```

### Logs

```bash
# Show pod logs
kube-logs my-pod

# Follow logs real-time
kube-logs my-pod -f

# Show last 100 lines
kube-logs my-pod -t 100

# Logs of specific container
kube-logs my-pod --container container-name

# Include timestamps in output
kube-logs my-pod --timestamps
```

### Port forwarding

```bash
# Forward port 8080 local -> 80 remote
kube-port-forward my-pod 8080:80

# Forward same port (3000 -> 3000)
kube-port-forward my-pod 3000
```

### Exec into Pods

```bash
# Open bash shell
kube-exec my-pod -- bash

# Execute specific command
kube-exec my-pod -- ls -la /app

# Exec into a specific container
kube-exec my-pod --container container-name -- env
```

### Using global flags

```bash
# Specify namespace
kube-pods -n kube-system

# Specify context
kube-pods -c my-context

# Combine both
kube-pods -n kube-system -c my-context
```

## Configuration

Tools use kubeconfig from `~/.kube/config` by default. You can:

1. Use `KUBECONFIG` environment variable
2. Specify context and namespace with flags `-c` and `-n`

```bash
# Set KUBECONFIG environment variable
export KUBECONFIG=/path/to/kubeconfig
kube-pods

# Or use flags
kube-pods -c my-context -n my-namespace
```

## Installation Options

### 📋 Script Options

```bash
# Show all options
curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash -s -- --help

# Command line options:
--dir DIR          # Install directory (default: /usr/local/bin)
--build-only       # Build only, do not install  
--force            # Overwrite existing files
--quiet            # Quiet mode
--uninstall        # Uninstall all tools
--repo URL         # Custom repository URL
--branch BRANCH    # Git branch (default: main)

# Environment variables (override command options):
KUBE_INSTALL_DIR   # Install directory
KUBE_BUILD_ONLY    # true/false - Build only
KUBE_FORCE         # true/false - Overwrite files
KUBE_QUIET         # true/false - Quiet mode
KUBE_REPO_URL      # Repository URL
KUBE_BRANCH        # Git branch
```

### 🗑️ Uninstall

```bash
# Remove all tools
curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash -s -- --uninstall

# Or if you already have the script locally
./install.sh --uninstall

# Or use Makefile
sudo make uninstall-all
```

## Tool List

| Tool | Description | Main Flags |
|------|-------|-----------|
| `kube-pods` | List pods | `-A`, `-n`, `-c` |
| `kube-services` | List services | `-A`, `-n`, `-c` |
| `kube-switch-context` | Switch context | - |
| `kube-switch-namespace` | Switch namespace | - |
| `kube-logs` | Show logs | `-f`, `-t`, `--container` |
| `kube-port-forward` | Port forwarding | `-n`, `-c` |
| `kube-exec` | Exec into pod | `--container`, `-t`, `-i` |

## Common workflows

### 1. Initialize and explore the cluster

```bash
# List available contexts
kube-switch-context

# Switch to desired context
kube-switch-context production

# Switch namespace
kube-switch-namespace my-app

# List pods
kube-pods
```

### 2. Debug applications

```bash
# List problematic pods
kube-pods

# Tail logs
kube-logs problematic-pod -f

# Exec into pod for debugging
kube-exec problematic-pod -- bash

# Port forward to test locally
kube-port-forward my-app-pod 8080:80
```

### 3. Monitoring

```bash
# View all resources
kube-pods -A
kube-services -A

# Monitor logs
kube-logs app-pod -f --timestamps
```

## Feedback and Issues

If you find bugs or have improvement ideas, please open an issue or pull request.

## License

MIT License
