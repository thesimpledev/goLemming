//go:build !windows

package main

import (
	"fmt"
	"os"
)

func checkPlatform() {
	fmt.Fprintln(os.Stderr, "Error: GoLemming only runs on Windows.")
	fmt.Fprintln(os.Stderr, "It requires Windows APIs for screen capture and input simulation.")
	os.Exit(1)
}
