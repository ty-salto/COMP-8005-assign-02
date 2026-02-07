package main

import (
	"fmt"
	"os"
	"strings"
)

func loadShadowHash(path, username string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot open shadow file: %w", err)
	}
	lines := strings.Split(string(b), "\n")
	prefix := username + ":"
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			parts := strings.Split(line, ":")
			if len(parts) < 2 || parts[1] == "" {
				return "", fmt.Errorf("malformed or unsupported hash entry")
			}
			return parts[1], nil
		}
	}
	return "", fmt.Errorf("username not in shadow file")
}


func validateHash(fullHash string) (string, error) {
	switch {
	case strings.HasPrefix(fullHash, "$1$"):
		return "md5", nil
	case strings.HasPrefix(fullHash, "$5$"):
		return "sha256", nil
	case strings.HasPrefix(fullHash, "$6$"):
		return "sha512", nil
	case strings.HasPrefix(fullHash, "$2a$") || strings.HasPrefix(fullHash, "$2b$") || strings.HasPrefix(fullHash, "$2y$"):
		return "bcrypt", nil
	case strings.HasPrefix(fullHash, "$y$") || strings.HasPrefix(fullHash, "$7$"):
		return "yescrypt", nil
	default:
		return "", fmt.Errorf("malformed or unsupported hash entry")
	}
}