package utils

import (
	"fmt"
	"time"
)

// FormatAge chuyển đổi duration thành format age giống kubectl
// Ví dụ: 5m, 1h, 2d
func FormatAge(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

// FormatBytes chuyển đổi bytes thành định dạng human-readable
// Ví dụ: 1024 -> 1Ki, 1048576 -> 1Mi
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ci", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatCPU chuyển đổi CPU millicores thành format readable
// Ví dụ: 1000 -> 1, 500 -> 500m
func FormatCPU(millicores int64) string {
	if millicores >= 1000 {
		return fmt.Sprintf("%d", millicores/1000)
	}
	return fmt.Sprintf("%dm", millicores)
}

// TruncateString cắt ngắn string nếu dài hơn maxLength
// Thêm "..." nếu string bị cắt
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	if maxLength <= 3 {
		return s[:maxLength]
	}
	return s[:maxLength-3] + "..."
}
