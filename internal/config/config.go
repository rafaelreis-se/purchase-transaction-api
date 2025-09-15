package config

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Path string
}

// LoadConfig loads configuration with default values
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: ":8080",
		},
		Database: DatabaseConfig{
			Path: "transactions.db",
		},
	}
}
