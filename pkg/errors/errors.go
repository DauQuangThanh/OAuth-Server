package errors

import "fmt"

// AppError represents an application error
type AppError struct {
	Code    string `json:"error"`
	Message string `json:"error_description"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WithMessage returns a new AppError with the specified message
func (e *AppError) WithMessage(message string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: message,
	}
}

// Common application errors
var (
	ErrInvalidRequest       = &AppError{Code: "invalid_request", Message: "The request is invalid"}
	ErrInvalidGrant         = &AppError{Code: "invalid_grant", Message: "Invalid credentials"}
	ErrUnsupportedGrantType = &AppError{Code: "unsupported_grant_type", Message: "Grant type not supported"}
	ErrUnauthorized         = &AppError{Code: "unauthorized", Message: "Authentication required"}
	ErrForbidden            = &AppError{Code: "forbidden", Message: "Access denied"}
	ErrNotFound             = &AppError{Code: "not_found", Message: "Resource not found"}
	ErrMethodNotAllowed     = &AppError{Code: "method_not_allowed", Message: "Method not allowed"}
	ErrUserExists           = &AppError{Code: "account_exists", Message: "Account already exists"}
	ErrInternalServerError  = &AppError{Code: "server_error", Message: "Internal server error"}
	ErrServiceUnavailable   = &AppError{Code: "service_unavailable", Message: "Service temporarily unavailable"}
)
