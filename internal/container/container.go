package container

import (
	"context"
	"database/sql"
	"fmt"

	"auth0-server/internal/application/ports"
	"auth0-server/internal/application/usecases"
	"auth0-server/internal/config"
	"auth0-server/internal/domain/account"
	"auth0-server/internal/domain/auth"
	"auth0-server/internal/infrastructure/cache"
	"auth0-server/internal/infrastructure/crypto"
	"auth0-server/internal/infrastructure/monitoring"
	"auth0-server/internal/infrastructure/storage"
	"auth0-server/internal/infrastructure/workers"
	"auth0-server/internal/interfaces/http/handlers"
	"auth0-server/internal/interfaces/http/middleware"
	"auth0-server/pkg/logger"
)

// Container holds all application dependencies
type Container struct {
	Config *config.EnhancedConfig
	Logger logger.Logger

	// Infrastructure
	Database   *sql.DB
	Cache      ports.CacheRepository
	WorkerPool *workers.WorkerPool
	Metrics    *monitoring.MetricsCollector
	Health     *monitoring.HealthChecker

	// Services
	PasswordHasher account.PasswordHasher
	TokenService   auth.TokenService
	IDGenerator    *crypto.IDGenerator

	// Repositories
	AccountRepository account.Repository

	// Use Cases
	AccountUseCase *usecases.AccountUseCase
	AuthUseCase    *usecases.AuthUseCase

	// Handlers
	AuthHandler   *handlers.AuthHandler
	ConfigHandler *handlers.ConfigHandler

	// Middleware
	AuthMiddleware *middleware.AuthMiddleware
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.EnhancedConfig) (*Container, error) {
	c := &Container{
		Config: cfg,
		Logger: logger.NewStandardLogger(),
	}

	if err := c.initializeInfrastructure(); err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	if err := c.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	if err := c.initializeRepositories(); err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	if err := c.initializeUseCases(); err != nil {
		return nil, fmt.Errorf("failed to initialize use cases: %w", err)
	}

	if err := c.initializeHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}

	if err := c.initializeHealthChecks(); err != nil {
		return nil, fmt.Errorf("failed to initialize health checks: %w", err)
	}

	return c, nil
}

// initializeInfrastructure sets up infrastructure components
func (c *Container) initializeInfrastructure() error {
	// Initialize monitoring
	c.Metrics = monitoring.NewMetricsCollector()
	c.Health = monitoring.NewHealthChecker()

	// Initialize worker pool
	c.WorkerPool = workers.NewWorkerPool(c.Config.Worker.PoolSize, c.Config.Worker.QueueSize)
	c.WorkerPool.Start()

	// Initialize cache
	c.Cache = cache.NewInMemoryCache(c.Config.Cache.MaxSize)

	// Initialize database with fallback
	if err := c.initializeDatabase(); err != nil {
		c.Logger.Error("Database initialization failed, continuing with in-memory storage", err, nil)
	}

	return nil
}

// initializeDatabase sets up database connection with fallback
func (c *Container) initializeDatabase() error {
	if c.Config.Database.Driver != "postgres" {
		return nil // Skip database initialization for non-postgres drivers
	}

	dbConfig := &storage.DatabaseConfig{
		Host:            c.Config.Database.Host,
		Port:            c.Config.Database.Port,
		User:            c.Config.Database.User,
		Password:        c.Config.Database.Password,
		DBName:          c.Config.Database.DBName,
		SSLMode:         c.Config.Database.SSLMode,
		MaxOpenConns:    c.Config.Database.MaxConnections,
		MaxIdleConns:    c.Config.Database.MaxIdleConns,
		ConnMaxLifetime: c.Config.Database.ConnMaxLifetime,
		ConnMaxIdleTime: c.Config.Database.ConnMaxIdleTime,
	}

	// Create database if it doesn't exist
	if err := storage.CreateDatabaseIfNotExists(dbConfig); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Connect to database
	db, err := storage.ConnectPostgreSQL(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test database schema
	ctx, cancel := context.WithTimeout(context.Background(), c.Config.Database.ConnMaxIdleTime)
	defer cancel()

	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM accounts").Scan(&count)
	if err != nil {
		db.Close()
		return fmt.Errorf("database schema not ready: %w", err)
	}

	c.Database = db
	c.Logger.Info("PostgreSQL connected successfully", map[string]interface{}{
		"database": dbConfig.DBName,
		"accounts": count,
	})

	return nil
}

// initializeServices sets up business services
func (c *Container) initializeServices() error {
	c.PasswordHasher = crypto.DefaultPasswordHasher()
	c.IDGenerator = crypto.NewIDGenerator()

	// Cast to the correct interface
	jweService := crypto.NewJWETokenService(c.Config.JWESecret, c.Config.Issuer, []string{"auth0-server"})
	c.TokenService = jweService

	return nil
}

// initializeRepositories sets up data repositories
func (c *Container) initializeRepositories() error {
	if c.Config.Database.Driver == "memory" {
		c.Logger.Info("Using in-memory account repository", nil)
		c.AccountRepository = storage.NewInMemoryAccountRepository(c.Logger)
	} else if c.Database != nil {
		c.Logger.Info("Using PostgreSQL account repository", nil)
		c.AccountRepository = storage.NewPostgresAccountRepository(c.Database, c.Logger)
	} else {
		return fmt.Errorf("database connection is required for PostgreSQL account repository")
	}

	return nil
}

// initializeUseCases sets up application use cases
func (c *Container) initializeUseCases() error {
	c.AccountUseCase = usecases.NewAccountUseCase(c.AccountRepository, c.PasswordHasher, c.IDGenerator)
	c.AuthUseCase = usecases.NewAuthUseCase(c.AccountUseCase, c.TokenService)

	return nil
}

// initializeHandlers sets up HTTP handlers
func (c *Container) initializeHandlers() error {
	c.AuthHandler = handlers.NewAuthHandler(c.AuthUseCase, c.AccountUseCase, c.Logger)
	c.ConfigHandler = handlers.NewConfigHandler(c.Config.Config, c.Logger)
	c.AuthMiddleware = middleware.NewAuthMiddleware(c.AuthUseCase, c.Logger)

	return nil
}

// initializeHealthChecks sets up health check endpoints
func (c *Container) initializeHealthChecks() error {
	// Account repository health check
	c.Health.AddCheck("account_repository", func(ctx context.Context) error {
		// Test basic operation
		_, err := c.AccountRepository.GetByID(ctx, "health-check-non-existent")
		if err != nil && err.Error() == "account not found" {
			return nil // Expected for non-existent account
		}
		return err
	})

	// Database health check
	if c.Database != nil {
		c.Health.AddCheck("database", func(ctx context.Context) error {
			return c.Database.PingContext(ctx)
		})
	}

	// Cache health check
	c.Health.AddCheck("cache", func(ctx context.Context) error {
		testKey := "health-check"
		if err := c.Cache.Set(ctx, testKey, "test", 1); err != nil {
			return err
		}
		defer c.Cache.Delete(ctx, testKey)
		_, err := c.Cache.Get(ctx, testKey)
		return err
	})

	return nil
}

// Close gracefully shuts down all resources
func (c *Container) Close() error {
	var errs []error

	// Stop worker pool
	if c.WorkerPool != nil {
		c.WorkerPool.Stop()
	}

	// Close database
	if c.Database != nil {
		if err := c.Database.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close database: %w", err))
		}
	}

	// Close cache
	if c.Cache != nil {
		if err := c.Cache.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close cache: %w", err))
		}
	}

	// Close account repository if it has a Close method
	if closer, ok := c.AccountRepository.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close account repository: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}

	return nil
}
