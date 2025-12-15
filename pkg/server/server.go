package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"auth0-server/pkg/logger"
)

// Config contains server configuration
type Config struct {
	Address         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	MaxHeaderBytes  int
}

// DefaultConfig returns default server configuration optimized for high concurrency
func DefaultConfig() *Config {
	return &Config{
		Address:         ":8080",
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
		MaxHeaderBytes:  1 << 20, // 1MB
	}
}

// Server represents a high-performance HTTP server
type Server struct {
	httpServer *http.Server
	logger     logger.Logger
	config     *Config
}

// New creates a new server instance
func New(handler http.Handler, config *Config, logger logger.Logger) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	httpServer := &http.Server{
		Addr:           config.Address,
		Handler:        handler,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		IdleTimeout:    config.IdleTimeout,
		MaxHeaderBytes: config.MaxHeaderBytes,
	}

	return &Server{
		httpServer: httpServer,
		logger:     logger,
		config:     config,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info(fmt.Sprintf("Starting server on %s", s.config.Address), nil)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop() error {
	s.logger.Info("Shutting down server...", nil)

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Server forced to shutdown", err, nil)
		return err
	}

	s.logger.Info("Server shutdown complete", nil)
	return nil
}
