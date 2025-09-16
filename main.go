package main

import (
	"fmt"
	"os"

	"kube/cmd"
)

// main là entry point của ứng dụng kube
// Khởi tạo root command và thực thi CLI
func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
