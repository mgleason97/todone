package util

import (
	"log"
	"os"
	"strings"
)

// MustReadFile reads the contents of a file at the given path, crashing on any error.
func MustReadFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read contents at %s: %v", path, err)
	}
	return strings.TrimSpace(string(data))
}
