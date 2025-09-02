package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	App         AppConfig         `yaml:"app"`
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	Redis       RedisConfig       `yaml:"redis"`
	Logging     LoggingConfig     `yaml:"logging"`
	CORS        CORSConfig        `yaml:"cors"`
	JWT         JWTConfig         `yaml:"jwt"`
	RateLimit   RateLimitConfig   `yaml:"rate_limit"`
	ExternalAPI ExternalAPIConfig `yaml:"external_apis"`
	FileUpload  FileUploadConfig  `yaml:"file_upload"`
	HealthCheck HealthCheckConfig `yaml:"health_check"`
	Metrics     MetricsConfig     `yaml:"metrics"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
	Debug       bool   `yaml:"debug"`
}

type ServerConfig struct {
	Host                    string        `yaml:"host"`
	Port                    int           `yaml:"port"`
	ReadTimeout             time.Duration `yaml:"read_timeout"`
	WriteTimeout            time.Duration `yaml:"write_timeout"`
	IdleTimeout             time.Duration `yaml:"idle_timeout"`
	GracefulShutdownTimeout time.Duration `yaml:"graceful_shutdown_timeout"`
}

type DatabaseConfig struct {
	Driver            string        `yaml:"driver"`
	URI               string        `yaml:"uri"`
	Host              string        `yaml:"host"`
	Port              int           `yaml:"port"`
	Name              string        `yaml:"name"`
	Username          string        `yaml:"username"`
	Password          string        `yaml:"password"`
	SSLMode           string        `yaml:"ssl_mode"`
	MaxPoolSize       int           `yaml:"max_pool_size"`
	MinPoolSize       int           `yaml:"min_pool_size"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
}

type RedisConfig struct {
	Host               string `yaml:"host"`
	Port               int    `yaml:"port"`
	Password           string `yaml:"password"`
	Database           int    `yaml:"database"`
	PoolSize           int    `yaml:"pool_size"`
	MinIdleConnections int    `yaml:"min_idle_connections"`
}

type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
}

type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

type JWTConfig struct {
	Secret            string        `yaml:"secret"`
	Expiration        time.Duration `yaml:"expiration"`
	RefreshExpiration time.Duration `yaml:"refresh_expiration"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `yaml:"requests_per_minute"`
	Burst             int `yaml:"burst"`
}

type ExternalAPIConfig struct {
	Timeout       time.Duration `yaml:"timeout"`
	RetryAttempts int           `yaml:"retry_attempts"`
	RetryDelay    time.Duration `yaml:"retry_delay"`
}

type FileUploadConfig struct {
	MaxSize      int64    `yaml:"max_size"`
	AllowedTypes []string `yaml:"allowed_types"`
	UploadPath   string   `yaml:"upload_path"`
}

type HealthCheckConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
}

type MetricsConfig struct {
	Enabled   bool   `yaml:"enabled"`
	Endpoint  string `yaml:"endpoint"`
	Namespace string `yaml:"namespace"`
}

// Global config instance
var cfg *Config

// Load reads the configuration from a YAML file
func Load(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	cfg = &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override with environment variables if set
	overrideWithEnv()

	return nil
}

// Get returns the global configuration instance
func Get() *Config {
	if cfg == nil {
		panic("configuration not loaded. Call config.Load() first")
	}
	return cfg
}

// overrideWithEnv overrides configuration values with environment variables
func overrideWithEnv() {
	if env := os.Getenv("APP_ENVIRONMENT"); env != "" {
		cfg.App.Environment = env
	}
	if env := os.Getenv("SERVER_PORT"); env != "" {
		fmt.Sscanf(env, "%d", &cfg.Server.Port)
	}
	if env := os.Getenv("DB_HOST"); env != "" {
		cfg.Database.Host = env
	}
	if env := os.Getenv("DB_PORT"); env != "" {
		fmt.Sscanf(env, "%d", &cfg.Database.Port)
	}
	if env := os.Getenv("DB_NAME"); env != "" {
		cfg.Database.Name = env
	}
	if env := os.Getenv("DB_USERNAME"); env != "" {
		cfg.Database.Username = env
	}
	if env := os.Getenv("DB_PASSWORD"); env != "" {
		cfg.Database.Password = env
	}
	if env := os.Getenv("REDIS_HOST"); env != "" {
		cfg.Redis.Host = env
	}
	if env := os.Getenv("REDIS_PORT"); env != "" {
		fmt.Sscanf(env, "%d", &cfg.Redis.Port)
	}
	if env := os.Getenv("REDIS_PASSWORD"); env != "" {
		cfg.Redis.Password = env
	}
	if env := os.Getenv("JWT_SECRET"); env != "" {
		cfg.JWT.Secret = env
	}
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	switch c.Database.Driver {
	case "mongodb":
		if c.Database.URI != "" {
			return c.Database.URI
		}
		return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			c.Database.Username, c.Database.Password, c.Database.Host,
			c.Database.Port, c.Database.Name)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Database.Host, c.Database.Port, c.Database.Username,
			c.Database.Password, c.Database.Name, c.Database.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.Database.Username, c.Database.Password, c.Database.Host,
			c.Database.Port, c.Database.Name)
	case "sqlite":
		return c.Database.Name
	default:
		return ""
	}
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetRedisAddress returns the Redis address
func (c *Config) GetRedisAddress() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}
