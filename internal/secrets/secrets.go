package secrets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultSecretsDir = "/run/secrets"

func Read(key string) (string, error) {
	return read(defaultSecretsDir, key)
}

func read(dir, key string) (string, error) {
	if data, err := os.ReadFile(filepath.Join(dir, key)); err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	if value := os.Getenv(strings.ToUpper(key)); value != "" {
		return value, nil
	}

	return "", fmt.Errorf("secret %q was not found in either /run/secrets/ nor environment variables", key)
}
