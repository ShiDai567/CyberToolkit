package config

import (
	"bufio"
	"os"
	"strings"
)

type Config struct {
	Addr          string
	AdminEmail    string
	AdminPassword string
	AdminToken    string
}

func Load() Config {
	loadDotEnv(".env")

	return Config{
		Addr:          getenv("APP_ADDR", ":8080"),
		AdminEmail:    getenv("APP_ADMIN_EMAIL", "admin@cybertoolkit.local"),
		AdminPassword: getenv("APP_ADMIN_PASSWORD", "admin123456"),
		AdminToken:    getenv("APP_ADMIN_TOKEN", "dev-admin-token"),
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}

		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		_ = os.Setenv(key, value)
	}
}
