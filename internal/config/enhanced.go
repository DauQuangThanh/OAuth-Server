package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver          string
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxConnections  int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Type            string // "memory", "redis"
	RedisURL        string
	DefaultTTL      time.Duration
	MaxSize         int
	CleanupInterval time.Duration
}

// WorkerConfig holds worker pool configuration
type WorkerConfig struct {
	PoolSize        int
	QueueSize       int
	TaskTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	EnableMetrics   bool
	EnableTracing   bool
	MetricsPath     string
	HealthCheckPath string
	MetricsPort     int
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	EnableHTTPS       bool
	CertFile          string
	KeyFile           string
	JWEEncryption     bool
	TokenExpiration   time.Duration
	RefreshExpiration time.Duration
	MaxLoginAttempts  int
	LockoutDuration   time.Duration
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host              string
	Port              int
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	MaxHeaderBytes    int
	EnableCompression bool
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool
	RequestsPerSecond int
	BurstSize         int
	CleanupInterval   time.Duration
}

// EnhancedConfig extends the base config with additional settings
type EnhancedConfig struct {
	*Config // Embed the original config

	Database    DatabaseConfig
	Cache       CacheConfig
	Worker      WorkerConfig
	Monitoring  MonitoringConfig
	Security    SecurityConfig
	Server      ServerConfig
	RateLimit   RateLimitConfig
	Environment string
}

// LoadEnhancedConfig loads the enhanced configuration
func LoadEnhancedConfig() (*EnhancedConfig, error) {
	// Load base config first
	baseConfig, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	config := &EnhancedConfig{
		Config: baseConfig,
	}

	// Load enhanced configurations
	config.loadDatabaseConfig()
	config.loadCacheConfig()
	config.loadWorkerConfig()
	config.loadMonitoringConfig()
	config.loadSecurityConfig()
	config.loadServerConfig()
	config.loadRateLimitConfig()

	config.Environment = getEnvString("ENVIRONMENT", "development")

	return config, nil
}

func (c *EnhancedConfig) loadDatabaseConfig() {
	c.Database = DatabaseConfig{
		Driver:          getEnvString("DB_DRIVER", "memory"), // Default to memory for testing
		Host:            getEnvString("DB_HOST", "localhost"),
		Port:            getEnvInt("DB_PORT", 5432),
		User:            getEnvString("DB_USER", "postgres"),
		Password:        getEnvString("DB_PASSWORD", ""),
		DBName:          getEnvString("DB_NAME", "Auth0_DB"),
		SSLMode:         getEnvString("DB_SSL_MODE", "disable"),
		MaxConnections:  getEnvInt("DB_MAX_CONNECTIONS", 25),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 15*time.Minute),
		ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
	}
}

func (c *EnhancedConfig) loadCacheConfig() {
	c.Cache = CacheConfig{
		Type:            getEnvString("CACHE_TYPE", "memory"),
		RedisURL:        getEnvString("REDIS_URL", ""),
		DefaultTTL:      getEnvDuration("CACHE_DEFAULT_TTL", 10*time.Minute),
		MaxSize:         getEnvInt("CACHE_MAX_SIZE", 1000),
		CleanupInterval: getEnvDuration("CACHE_CLEANUP_INTERVAL", 5*time.Minute),
	}
}

func (c *EnhancedConfig) loadWorkerConfig() {
	c.Worker = WorkerConfig{
		PoolSize:        getEnvInt("WORKER_POOL_SIZE", 10),
		QueueSize:       getEnvInt("WORKER_QUEUE_SIZE", 100),
		TaskTimeout:     getEnvDuration("WORKER_TASK_TIMEOUT", 30*time.Second),
		ShutdownTimeout: getEnvDuration("WORKER_SHUTDOWN_TIMEOUT", 10*time.Second),
	}
}

func (c *EnhancedConfig) loadMonitoringConfig() {
	c.Monitoring = MonitoringConfig{
		EnableMetrics:   getEnvBool("ENABLE_METRICS", true),
		EnableTracing:   getEnvBool("ENABLE_TRACING", true),
		MetricsPath:     getEnvString("METRICS_PATH", "/metrics"),
		HealthCheckPath: getEnvString("HEALTH_CHECK_PATH", "/health"),
		MetricsPort:     getEnvInt("METRICS_PORT", 0), // 0 means use same port as main server
	}
}

func (c *EnhancedConfig) loadSecurityConfig() {
	c.Security = SecurityConfig{
		EnableHTTPS:       getEnvBool("ENABLE_HTTPS", false),
		CertFile:          getEnvString("CERT_FILE", ""),
		KeyFile:           getEnvString("KEY_FILE", ""),
		JWEEncryption:     getEnvBool("JWE_ENCRYPTION", true),
		TokenExpiration:   getEnvDuration("TOKEN_EXPIRATION", 1*time.Hour),
		RefreshExpiration: getEnvDuration("REFRESH_EXPIRATION", 24*time.Hour),
		MaxLoginAttempts:  getEnvInt("MAX_LOGIN_ATTEMPTS", 5),
		LockoutDuration:   getEnvDuration("LOCKOUT_DURATION", 15*time.Minute),
	}
}

func (c *EnhancedConfig) loadServerConfig() {
	// Parse SERVER_ADDRESS if provided, otherwise use individual host/port
	if serverAddr := getEnvString("SERVER_ADDRESS", ""); serverAddr != "" {
		// Parse :port or host:port format
		if strings.HasPrefix(serverAddr, ":") {
			// Just port specified
			if port, err := strconv.Atoi(serverAddr[1:]); err == nil {
				c.Server.Host = "localhost"
				c.Server.Port = port
			} else {
				c.Server.Host = "localhost"
				c.Server.Port = 8080
			}
		} else if host, portStr, err := net.SplitHostPort(serverAddr); err == nil {
			if port, err := strconv.Atoi(portStr); err == nil {
				c.Server.Host = host
				c.Server.Port = port
			} else {
				c.Server.Host = "localhost"
				c.Server.Port = 8080
			}
		} else {
			// Fallback to defaults
			c.Server.Host = "localhost"
			c.Server.Port = 8080
		}
	} else {
		// Use individual environment variables
		c.Server.Host = getEnvString("SERVER_HOST", "localhost")
		c.Server.Port = getEnvInt("SERVER_PORT", 8080)
	}

	c.Server.ReadTimeout = getEnvDuration("SERVER_READ_TIMEOUT", 15*time.Second)
	c.Server.WriteTimeout = getEnvDuration("SERVER_WRITE_TIMEOUT", 15*time.Second)
	c.Server.IdleTimeout = getEnvDuration("SERVER_IDLE_TIMEOUT", 60*time.Second)
	c.Server.ShutdownTimeout = getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 15*time.Second)
	c.Server.MaxHeaderBytes = getEnvInt("SERVER_MAX_HEADER_BYTES", 1<<20) // 1MB
	c.Server.EnableCompression = getEnvBool("ENABLE_COMPRESSION", true)
}

func (c *EnhancedConfig) loadRateLimitConfig() {
	c.RateLimit = RateLimitConfig{
		Enabled:           getEnvBool("RATE_LIMIT_ENABLED", true),
		RequestsPerSecond: getEnvInt("RATE_LIMIT_RPS", 100),
		BurstSize:         getEnvInt("RATE_LIMIT_BURST", 200),
		CleanupInterval:   getEnvDuration("RATE_LIMIT_CLEANUP", 1*time.Minute),
	}
}

// IsDevelopment returns true if running in development environment
func (c *EnhancedConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production environment
func (c *EnhancedConfig) IsProduction() bool {
	return c.Environment == "production"
}

// GetServerAddress returns the full server address
func (c *EnhancedConfig) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// Helper functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
