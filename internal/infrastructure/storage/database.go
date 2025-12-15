package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// DatabaseConfig holds PostgreSQL connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string

	// Connection pool settings
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultDatabaseConfig returns a default configuration for PostgreSQL
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "",
		DBName:          "Auth0_DB",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 15 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// ConnectPostgreSQL establishes a connection to PostgreSQL database
func ConnectPostgreSQL(config *DatabaseConfig) (*sql.DB, error) {
	// Build connection string using URI format (more reliable)
	var connStr string
	if config.Password != "" {
		connStr = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s",
			config.User, config.Password, config.Host, config.Port, config.DBName, config.SSLMode,
		)
	} else {
		connStr = fmt.Sprintf(
			"postgres://%s@%s:%d/%s?sslmode=%s",
			config.User, config.Host, config.Port, config.DBName, config.SSLMode,
		)
	}

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// CreateDatabaseIfNotExists creates the database if it doesn't exist
func CreateDatabaseIfNotExists(config *DatabaseConfig) error {
	// Connect to postgres database to create the target database
	tempConfig := *config
	tempConfig.DBName = "postgres" // Connect to default postgres database

	var connStr string
	if tempConfig.Password != "" {
		connStr = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s",
			tempConfig.User, tempConfig.Password, tempConfig.Host, tempConfig.Port, tempConfig.DBName, tempConfig.SSLMode,
		)
	} else {
		connStr = fmt.Sprintf(
			"postgres://%s@%s:%d/%s?sslmode=%s",
			tempConfig.User, tempConfig.Host, tempConfig.Port, tempConfig.DBName, tempConfig.SSLMode,
		)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer db.Close()

	// Check if database exists
	var exists bool
	checkQuery := "SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = $1)"
	err = db.QueryRow(checkQuery, config.DBName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	// Create database if it doesn't exist
	if !exists {
		createQuery := fmt.Sprintf(`CREATE DATABASE "%s"`, config.DBName)
		_, err = db.Exec(createQuery)
		if err != nil {
			return fmt.Errorf("failed to create database %s: %w", config.DBName, err)
		}
	}

	return nil
}
