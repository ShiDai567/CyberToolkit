package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Addr             string
	AdminEmail       string
	AdminPassword    string
	AdminDisplayName string
	CORSOrigins      []string
	DatabaseURL      string
	RedisURL         string
}

// currentVersion is bumped whenever the config schema changes.
// When a config file with an older version is loaded, it is merged
// with the new defaults and re-written so the user sees new fields.
const currentVersion = 1

// yamlConfig mirrors Config for YAML serialization.
type yamlConfig struct {
	Version int `yaml:"version"`
	Server  struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	CORS struct {
		AllowedOrigins []string `yaml:"allowed_origins"`
	} `yaml:"cors"`
	Database struct {
		PostgreSQL struct {
			Enabled  bool   `yaml:"enabled"`
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			DBName   string `yaml:"dbname"`
			SSLMode  string `yaml:"sslmode"`
		} `yaml:"postgresql"`
	} `yaml:"database"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
}

const configPath = "data/config.yaml"

func Load() Config {
	yc := loadOrCreate(configPath)

	addr := fmt.Sprintf(":%d", yc.Server.Port)

	// Build postgres connection URL from individual fields.
	pg := yc.Database.PostgreSQL
	if !pg.Enabled {
		panic("config: database.postgresql is not enabled")
	}
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		pg.User, pg.Password,
		pg.Host, pg.Port,
		pg.DBName, pg.SSLMode,
	)

	// Build redis connection URL from individual fields.
	var redisURL string
	if yc.Redis.Password != "" {
		redisURL = fmt.Sprintf("redis://:%s@%s:%d/%d", yc.Redis.Password, yc.Redis.Host, yc.Redis.Port, yc.Redis.DB)
	} else {
		redisURL = fmt.Sprintf("redis://%s:%d/%d", yc.Redis.Host, yc.Redis.Port, yc.Redis.DB)
	}

	return Config{
		Addr:             addr,
		AdminEmail:       getenv("ADMIN_EMAIL", "admin@cybertoolkit.local"),
		AdminPassword:    getenv("ADMIN_PASSWORD", "admin123456"),
		AdminDisplayName: getenv("ADMIN_USER", "Admin"),
		CORSOrigins:      yc.CORS.AllowedOrigins,
		DatabaseURL:      dbURL,
		RedisURL:         redisURL,
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// loadOrCreate reads the YAML config file, creating it with defaults if absent.
func loadOrCreate(path string) yamlConfig {
	defaults := defaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(fmt.Sprintf("config: read %s: %v", path, err))
		}
		// File does not exist — write defaults and return them.
		writeConfig(path, defaults)
		return defaults
	}

	var yc yamlConfig
	if err := yaml.Unmarshal(data, &yc); err != nil {
		panic(fmt.Sprintf("config: parse %s: %v", path, err))
	}

	// Fill any zero-value fields with defaults so the file can be sparse.
	merge(&yc, defaults)

	// If the on-disk version is outdated, bump it and re-write so the
	// user sees any newly added fields with their default values.
	if yc.Version < currentVersion {
		yc.Version = currentVersion
		writeConfig(path, yc)
	}

	return yc
}

func writeConfig(path string, yc yamlConfig) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		panic(fmt.Sprintf("config: mkdir %s: %v", filepath.Dir(path), err))
	}

	header := "# CyberToolkit backend configuration\n" +
		"# This file is auto-generated on first run. Edit to customise.\n\n"

	data, err := yaml.Marshal(yc)
	if err != nil {
		panic(fmt.Sprintf("config: marshal yaml: %v", err))
	}

	if err := os.WriteFile(path, []byte(header+string(data)), 0o640); err != nil {
		panic(fmt.Sprintf("config: write %s: %v", path, err))
	}
}

// merge fills zero-value fields in dst from src.
func merge(dst *yamlConfig, src yamlConfig) {
	if dst.Server.Port == 0 {
		dst.Server.Port = src.Server.Port
	}
	if len(dst.CORS.AllowedOrigins) == 0 {
		dst.CORS.AllowedOrigins = src.CORS.AllowedOrigins
	}
	if dst.Database.PostgreSQL.Host == "" {
		dst.Database.PostgreSQL.Host = src.Database.PostgreSQL.Host
	}
	if dst.Database.PostgreSQL.Port == 0 {
		dst.Database.PostgreSQL.Port = src.Database.PostgreSQL.Port
	}
	if dst.Database.PostgreSQL.User == "" {
		dst.Database.PostgreSQL.User = src.Database.PostgreSQL.User
	}
	if dst.Database.PostgreSQL.Password == "" {
		dst.Database.PostgreSQL.Password = src.Database.PostgreSQL.Password
	}
	if dst.Database.PostgreSQL.DBName == "" {
		dst.Database.PostgreSQL.DBName = src.Database.PostgreSQL.DBName
	}
	if dst.Database.PostgreSQL.SSLMode == "" {
		dst.Database.PostgreSQL.SSLMode = src.Database.PostgreSQL.SSLMode
	}
	if dst.Redis.Host == "" {
		dst.Redis.Host = src.Redis.Host
	}
	if dst.Redis.Port == 0 {
		dst.Redis.Port = src.Redis.Port
	}
	// Redis.Password 允许为空（无密码），不做 merge
	// Redis.DB 允许为 0（默认库），不做 merge
}

func defaultConfig() yamlConfig {
	var yc yamlConfig
	yc.Version = currentVersion
	yc.Server.Port = 8080
	yc.CORS.AllowedOrigins = []string{
		"http://localhost:3000",
		"http://127.0.0.1:3000",
	}
	yc.Database.PostgreSQL.Enabled = true
	yc.Database.PostgreSQL.Host = "localhost"
	yc.Database.PostgreSQL.Port = 5432
	yc.Database.PostgreSQL.User = "postgres"
	yc.Database.PostgreSQL.Password = "postgres"
	yc.Database.PostgreSQL.DBName = "cybertoolkit"
	yc.Database.PostgreSQL.SSLMode = "disable"
	yc.Redis.Host = "localhost"
	yc.Redis.Port = 6379
	yc.Redis.Password = ""
	yc.Redis.DB = 0
	return yc
}
