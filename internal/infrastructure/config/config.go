package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server configuration
	Port        string
	Host        string
	Environment string

	// Logging configuration
	LogLevel             string
	EnableRequestLogging bool

	// Analysis configuration
	MaxProjectSize int64
	MaxFileSize    int64
	ScanTimeout    int

	// API configuration
	APIVersion        string
	RateLimitEnabled  bool
	RateLimitRequests int
	RateLimitWindow   int

	// CORS configuration
	AllowedOrigins   []string
	AllowCredentials bool

	// Storage configuration
	StorageType       string
	MaxStoredProjects int
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		// Server defaults
		Port:        getEnv("PORT", "8080"),
		Host:        getEnv("HOST", "0.0.0.0"),
		Environment: getEnv("ENVIRONMENT", "development"),

		// Logging defaults
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		EnableRequestLogging: getEnvBool("ENABLE_REQUEST_LOGGING", true),

		// Analysis defaults
		MaxProjectSize: getEnvInt64("MAX_PROJECT_SIZE", 100*1024*1024), // 100MB
		MaxFileSize:    getEnvInt64("MAX_FILE_SIZE", 1024*1024),        // 1MB
		ScanTimeout:    getEnvInt("SCAN_TIMEOUT", 300),                 // 5 minutes

		// API defaults
		APIVersion:        getEnv("API_VERSION", "v1"),
		RateLimitEnabled:  getEnvBool("RATE_LIMIT_ENABLED", false),
		RateLimitRequests: getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnvInt("RATE_LIMIT_WINDOW", 60), // 1 minute

		// CORS defaults
		AllowedOrigins:   getEnvStringSlice("ALLOWED_ORIGINS", []string{"*"}),
		AllowCredentials: getEnvBool("ALLOW_CREDENTIALS", true),

		// Storage defaults
		StorageType:       getEnv("STORAGE_TYPE", "memory"),
		MaxStoredProjects: getEnvInt("MAX_STORED_PROJECTS", 10),
	}

	return config, nil
}

func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Environment) == "production"
}

func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.Environment) == "development"
}

func (c *Config) GetServerAddress() string {
	return c.Host + ":" + c.Port
}

func (c *Config) GetLogLevel() string {
	return strings.ToLower(c.LogLevel)
}

// Helper functions to get environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
