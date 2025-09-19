package config

import "os"

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Treasury TreasuryConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Path string
}

type TreasuryConfig struct {
	BaseURL        string
	TimeoutSeconds int
}

// LoadConfig loads configuration with default values
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", ":8080"),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "transactions.db"),
		},
		Treasury: TreasuryConfig{
			BaseURL:        getEnv("TREASURY_BASE_URL", "https://api.fiscaldata.treasury.gov/services/api/v1/accounting/od/rates_of_exchange"),
			TimeoutSeconds: getEnvInt("TREASURY_TIMEOUT_SECONDS", 30),
		},
	}
}

// getEnv gets an environment variable with a default fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as integer with a default fallback
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue := parseInt(value); intValue > 0 {
			return intValue
		}
	}
	return defaultValue
}

// parseInt safely parses string to int
func parseInt(s string) int {
	result := 0
	for _, char := range s {
		if char >= '0' && char <= '9' {
			result = result*10 + int(char-'0')
		} else {
			return 0 // Invalid character, return 0
		}
	}
	return result
}
