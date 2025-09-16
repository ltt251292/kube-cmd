# Kube Tools

Một collection các command line tools được viết bằng Go để làm việc với Kubernetes cluster một cách thuận tiện và hiệu quả hơn.

Thay vì một tool đơn lẻ, kube cung cấp nhiều tools chuyên biệt theo format `kube-*` giống như kubectl plugins.

## Các Tools

- 🚀 **kube-pods**: Xem danh sách pods với format đẹp
- 🔧 **kube-services**: Xem danh sách services
- 🔄 **kube-switch-context**: Chuyển đổi nhanh giữa các contexts
- 📁 **kube-switch-namespace**: Chuyển đổi namespace trong context hiện tại
- 📋 **kube-logs**: Xem logs real-time với nhiều tùy chọn
- 🔌 **kube-port-forward**: Forward ports từ local đến pods
- 💻 **kube-exec**: Thực thi commands trong containers

## Cài đặt

### 🚀 Cài đặt nhanh từ internet (Khuyến nghị)

```bash
# Cài đặt một lệnh duy nhất
curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash

# Hoặc download script và chạy
wget https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh
chmod +x install.sh
./install.sh
```

### ⚙️ Tùy chọn cài đặt

```bash
# Cài đặt vào thư mục khác
curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash -s -- --dir ~/bin

# Chỉ build, không cài đặt
curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash -s -- --build-only

# Gỡ bỏ tất cả tools
curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash -s -- --uninstall
```

### 🌍 Environment Variables

Bạn có thể override các settings bằng environment variables:

```bash
# Cài đặt vào thư mục khác
KUBE_INSTALL_DIR=~/bin curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash

# Chỉ build, chế độ quiet
KUBE_BUILD_ONLY=true KUBE_QUIET=true curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash

# Sử dụng repo fork khác
KUBE_REPO_URL=https://github.com/your-fork/kube-cmd.git KUBE_BRANCH=develop curl -fsSL ... | bash

# Force override files đã tồn tại
KUBE_FORCE=true curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash
```

### 🛠️ Build từ source (Developers)

```bash
git clone https://github.com/ltt251292/kube-cmd.git
cd kube

# Cài đặt tự động
./install.sh

# Hoặc build thủ công
make build-all
sudo make install-all
```

### Yêu cầu

**Để sử dụng:**
- Kubectl đã được cấu hình  
- Quyền truy cập vào Kubernetes cluster

**Để build từ source:**
- Go 1.21+
- Git
- Make

## Sử dụng

### Xem danh sách tools

```bash
# Xem tất cả tools có sẵn
kube
```

### Xem resources

```bash
# Xem pods
kube-pods
kube-pods -A  # Tất cả namespaces

# Xem services
kube-services
kube-services -n my-namespace
```

### Chuyển đổi context và namespace

```bash
# Xem danh sách contexts
kube-switch-context

# Chuyển sang context khác
kube-switch-context my-context

# Xem namespace hiện tại
kube-switch-namespace

# Chuyển sang namespace khác
kube-switch-namespace my-namespace
```

### Xem logs

```bash
# Xem logs của pod
kube-logs my-pod

# Follow logs real-time
kube-logs my-pod -f

# Xem 100 dòng cuối
kube-logs my-pod -t 100

# Xem logs của container cụ thể
kube-logs my-pod --container container-name

# Xem logs với timestamps
kube-logs my-pod --timestamps
```

### Port forwarding

```bash
# Forward port 8080 local -> 80 remote
kube-port-forward my-pod 8080:80

# Forward cùng port (3000 -> 3000)
kube-port-forward my-pod 3000
```

### Thực thi commands

```bash
# Mở bash shell
kube-exec my-pod -- bash

# Thực thi command cụ thể
kube-exec my-pod -- ls -la /app

# Exec vào container cụ thể
kube-exec my-pod --container container-name -- env
```

### Sử dụng với flags global

```bash
# Chỉ định namespace
kube-pods -n kube-system

# Chỉ định context
kube-pods -c my-context

# Kết hợp cả hai
kube-pods -n kube-system -c my-context
```

## Configuration

Tool sử dụng kubeconfig mặc định từ `~/.kube/config`. Bạn có thể:

1. Sử dụng biến môi trường `KUBECONFIG`
2. Chỉ định context và namespace với flags `-c` và `-n`

```bash
# Set biến môi trường KUBECONFIG
export KUBECONFIG=/path/to/kubeconfig
kube-pods

# Hoặc sử dụng flags
kube-pods -c my-context -n my-namespace
```

## Installation Options

### 📋 Script Options

```bash
# Xem tất cả options
curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash -s -- --help

# Command line options:
--dir DIR          # Thư mục cài đặt (default: /usr/local/bin)
--build-only       # Chỉ build, không cài đặt  
--force            # Ghi đè files đã tồn tại
--quiet            # Chế độ quiet (ít output)
--uninstall        # Gỡ bỏ tất cả tools
--repo URL         # Custom repository URL
--branch BRANCH    # Git branch (default: main)

# Environment variables (override command options):
KUBE_INSTALL_DIR   # Thư mục cài đặt
KUBE_BUILD_ONLY    # true/false - Chỉ build
KUBE_FORCE         # true/false - Ghi đè files
KUBE_QUIET         # true/false - Chế độ quiet
KUBE_REPO_URL      # Repository URL
KUBE_BRANCH        # Git branch
```

### 🗑️ Gỡ bỏ (Uninstall)

```bash
# Gỡ bỏ tất cả tools
curl -fsSL https://raw.githubusercontent.com/ltt251292/kube-cmd/main/install.sh | bash -s -- --uninstall

# Hoặc nếu đã có script local
./install.sh --uninstall

# Hoặc dùng Makefile
sudo make uninstall-all
```

## Danh sách Tools

| Tool | Mô tả | Flags chính |
|------|-------|-----------|
| `kube-pods` | Xem pods | `-A`, `-n`, `-c` |
| `kube-services` | Xem services | `-A`, `-n`, `-c` |
| `kube-switch-context` | Chuyển context | - |
| `kube-switch-namespace` | Chuyển namespace | - |
| `kube-logs` | Xem logs | `-f`, `-t`, `--container` |
| `kube-port-forward` | Port forwarding | `-n`, `-c` |
| `kube-exec` | Exec vào pod | `--container`, `-t`, `-i` |

## Ví dụ workflow thường dùng

### 1. Khởi tạo và explore cluster

```bash
# Xem contexts có sẵn
kube-switch-context

# Chuyển sang context cần thiết
kube-switch-context production

# Chuyển sang namespace
kube-switch-namespace my-app

# Xem pods
kube-pods
```

### 2. Debug ứng dụng

```bash
# Xem pods có vấn đề
kube-pods

# Xem logs realtime
kube-logs problematic-pod -f

# Exec vào pod để debug
kube-exec problematic-pod -- bash

# Port forward để test local
kube-port-forward my-app-pod 8080:80
```

### 3. Monitoring

```bash
# Xem tất cả resources
kube-pods -A
kube-services -A

# Monitor logs
kube-logs app-pod -f --timestamps
```

## Góp ý và Issues

Nếu bạn gặp lỗi hoặc có ý tưởng cải thiện, hãy tạo issue hoặc pull request.

## License

MIT License
