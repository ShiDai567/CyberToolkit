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
	CORSOrigins   []string
	DatabaseURL   string
}

func Load() Config {
	loadDotEnv(".env")

	return Config{
		Addr:          getAddr(),
		AdminEmail:    getenvCompat([]string{"ADMIN_EMAIL", "APP_ADMIN_EMAIL"}, "admin@cybertoolkit.local"),
		AdminPassword: getenvCompat([]string{"ADMIN_PASSWORD", "APP_ADMIN_PASSWORD"}, "admin123456"),
		CORSOrigins:   getCORSOrigins(),
		DatabaseURL:   getenvCompat([]string{"DATABASE_URL"}, "postgres://postgres:postgres@localhost:5432/cybertoolkit?sslmode=disable"),
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getenvCompat(keys []string, fallback string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return fallback
}

func getAddr() string {
	port := getenvCompat([]string{"PORT"}, "")
	if port != "" {
		if strings.HasPrefix(port, ":") {
			return port
		}
		return ":" + port
	}

	return getenvCompat([]string{"APP_ADDR"}, ":8080")
}

func getCORSOrigins() []string {
	raw := getenvCompat([]string{"CORS_ALLOWED_ORIGINS"}, "http://localhost:3000,http://127.0.0.1:3000")
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))

	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		origins = append(origins, origin)
	}

	return origins
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
