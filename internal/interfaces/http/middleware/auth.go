package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"auth0-server/internal/application/usecases"
	"auth0-server/pkg/errors"
	"auth0-server/pkg/logger"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	authUseCase *usecases.AuthUseCase
	logger      logger.Logger
	timeout     time.Duration
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authUseCase *usecases.AuthUseCase, logger logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authUseCase: authUseCase,
		logger:      logger,
		timeout:     10 * time.Second,
	}
}

// RequireAuth middleware validates JWT tokens
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), m.timeout)
		defer cancel()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.sendError(w, errors.ErrUnauthorized.WithMessage("Authorization header required"))
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.sendError(w, errors.ErrUnauthorized.WithMessage("Invalid authorization header format"))
			return
		}

		token := parts[1]
		claims, err := m.authUseCase.ValidateToken(ctx, token)
		if err != nil {
			m.logger.ErrorContext(ctx, "token validation failed", err, nil)
			m.sendError(w, errors.ErrUnauthorized.WithMessage("Invalid or expired token"))
			return
		}

		// Add user ID to request context
		ctx = context.WithValue(ctx, "userID", claims.Subject)
		ctx = context.WithValue(ctx, "userEmail", claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// CORS middleware for handling cross-origin requests
func CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// RateLimit middleware for basic rate limiting
func RateLimit(requestsPerSecond int) func(http.HandlerFunc) http.HandlerFunc {
	// Simple in-memory rate limiter (in production, use Redis or similar)
	limiter := make(map[string]time.Time)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			clientIP := r.RemoteAddr
			now := time.Now()

			if lastRequest, exists := limiter[clientIP]; exists {
				if now.Sub(lastRequest) < time.Second/time.Duration(requestsPerSecond) {
					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
					return
				}
			}

			limiter[clientIP] = now
			next.ServeHTTP(w, r)
		}
	}
}

// Timeout middleware adds timeout to requests
func Timeout(duration time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// Recovery middleware recovers from panics
func Recovery(logger logger.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.ErrorContext(r.Context(), "panic recovered", nil, map[string]interface{}{
						"panic": err,
						"path":  r.URL.Path,
					})
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		}
	}
}

// sendError sends an error response
func (m *AuthMiddleware) sendError(w http.ResponseWriter, err *errors.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	// Simple JSON encoding
	w.Write([]byte(`{"error":"` + err.Code + `","error_description":"` + err.Message + `"}`))
}
