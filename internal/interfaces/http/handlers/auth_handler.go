package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"auth0-server/internal/application/usecases"
	"auth0-server/internal/domain/account"
	"auth0-server/pkg/errors"
	"auth0-server/pkg/logger"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authUseCase    *usecases.AuthUseCase
	accountUseCase *usecases.AccountUseCase
	logger         logger.Logger
	timeout        time.Duration
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	authUseCase *usecases.AuthUseCase,
	accountUseCase *usecases.AccountUseCase,
	logger logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		authUseCase:    authUseCase,
		accountUseCase: accountUseCase,
		logger:         logger,
		timeout:        30 * time.Second, // Configurable timeout
	}
}

// TokenHandler handles OAuth2 token requests
func (h *AuthHandler) TokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	if r.Method != http.MethodPost {
		h.sendError(w, errors.ErrMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	grantType := r.FormValue("grant_type")
	switch grantType {
	case "authorization_code":
		h.handleAuthorizationCodeGrant(ctx, w, r)
	case "refresh_token":
		h.handleRefreshToken(ctx, w, r)
	default:
		h.sendError(w, errors.ErrUnsupportedGrantType, http.StatusBadRequest)
	}
}

// handleAuthorizationCodeGrant handles authorization code grant type with PKCE
func (h *AuthHandler) handleAuthorizationCodeGrant(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	clientID := r.FormValue("client_id")
	codeVerifier := r.FormValue("code_verifier")
	redirectURI := r.FormValue("redirect_uri")

	if code == "" || clientID == "" || codeVerifier == "" {
		h.sendError(w, errors.ErrInvalidRequest.WithMessage("code, client_id, and code_verifier are required"), http.StatusBadRequest)
		return
	}

	h.logger.InfoContext(ctx, "attempting authorization code exchange", map[string]interface{}{
		"client_id": clientID,
		"code":      code[:8] + "...", // Log only first 8 chars for security
	})

	tokenPair, err := h.authUseCase.ExchangeCodeForTokens(ctx, code, clientID, codeVerifier, redirectURI)
	if err != nil {
		h.logger.ErrorContext(ctx, "authorization code exchange failed", err, map[string]interface{}{
			"client_id": clientID,
		})
		h.sendError(w, errors.ErrInvalidGrant, http.StatusUnauthorized)
		return
	}

	h.logger.InfoContext(ctx, "authorization code exchange successful", map[string]interface{}{
		"client_id": clientID,
	})

	h.sendJSON(w, tokenPair, http.StatusOK)
}

// handleRefreshToken handles refresh token grant type
func (h *AuthHandler) handleRefreshToken(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	refreshToken := r.FormValue("refresh_token")

	if refreshToken == "" {
		h.sendError(w, errors.ErrInvalidRequest.WithMessage("refresh_token is required"), http.StatusBadRequest)
		return
	}

	tokenPair, err := h.authUseCase.RefreshAuthentication(ctx, refreshToken)
	if err != nil {
		h.logger.ErrorContext(ctx, "token refresh failed", err, nil)
		h.sendError(w, errors.ErrInvalidGrant, http.StatusUnauthorized)
		return
	}

	h.sendJSON(w, tokenPair, http.StatusOK)
}

// AuthorizeHandler handles OAuth 2.1 authorization requests with PKCE
func (h *AuthHandler) AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		h.sendError(w, errors.ErrMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// Parse OAuth 2.1 authorization parameters
	responseType := r.URL.Query().Get("response_type")
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	scope := r.URL.Query().Get("scope")
	codeChallenge := r.URL.Query().Get("code_challenge")
	codeChallengeMethod := r.URL.Query().Get("code_challenge_method")

	// Validate required parameters
	if responseType != "code" {
		h.sendAuthorizationError(w, redirectURI, "unsupported_response_type", "Only 'code' response type is supported", state)
		return
	}

	if clientID == "" || redirectURI == "" || codeChallenge == "" {
		h.sendAuthorizationError(w, redirectURI, "invalid_request", "client_id, redirect_uri, and code_challenge are required", state)
		return
	}

	// PKCE is mandatory in OAuth 2.1
	if codeChallengeMethod != "S256" {
		h.sendAuthorizationError(w, redirectURI, "invalid_request", "code_challenge_method must be S256", state)
		return
	}

	h.logger.InfoContext(ctx, "authorization request received", map[string]interface{}{
		"client_id":    clientID,
		"redirect_uri": redirectURI,
		"scope":        scope,
	})

	// For this demo, we'll show a simple login form
	// In production, this would check if user is authenticated and show consent
	if r.Method == http.MethodGet {
		h.renderLoginForm(w, clientID, redirectURI, state, scope, codeChallenge, codeChallengeMethod)
		return
	}

	// Handle POST - user submitted login credentials
	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		h.renderLoginForm(w, clientID, redirectURI, state, scope, codeChallenge, codeChallengeMethod)
		return
	}

	// Authenticate user (internal method, not password grant)
	authCode, err := h.authUseCase.CreateAuthorizationCode(ctx, email, password, clientID, redirectURI, scope, codeChallenge, codeChallengeMethod)
	if err != nil {
		h.logger.ErrorContext(ctx, "authentication failed in authorization flow", err, map[string]interface{}{
			"email":     email,
			"client_id": clientID,
		})
		h.renderLoginForm(w, clientID, redirectURI, state, scope, codeChallenge, codeChallengeMethod)
		return
	}

	// Redirect back to client with authorization code
	redirectURL := redirectURI + "?code=" + authCode
	if state != "" {
		redirectURL += "&state=" + state
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// UserInfoHandler handles account info requests (maintains Auth0 compatibility)
func (h *AuthHandler) UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	if r.Method != http.MethodGet {
		h.sendError(w, errors.ErrMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.sendError(w, errors.ErrUnauthorized.WithMessage("Authorization header required"), http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		h.sendError(w, errors.ErrUnauthorized.WithMessage("Invalid authorization header format"), http.StatusUnauthorized)
		return
	}

	token := parts[1]
	accountProfile, err := h.authUseCase.GetAccountProfile(ctx, token)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get account profile", err, nil)
		h.sendError(w, errors.ErrUnauthorized, http.StatusUnauthorized)
		return
	}

	h.sendJSON(w, accountProfile, http.StatusOK)
}

// SignupHandler handles account registration
func (h *AuthHandler) SignupHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	if r.Method != http.MethodPost {
		h.sendError(w, errors.ErrMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	var req account.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, errors.ErrInvalidRequest.WithMessage("Invalid JSON"), http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		h.sendError(w, errors.ErrInvalidRequest.WithMessage("email and password are required"), http.StatusBadRequest)
		return
	}

	h.logger.InfoContext(ctx, "attempting account registration", map[string]interface{}{
		"email": req.Email,
	})

	newAccount, err := h.accountUseCase.CreateAccount(ctx, req.Email, req.Password, req.Name)
	if err != nil {
		h.logger.ErrorContext(ctx, "account registration failed", err, map[string]interface{}{
			"email": req.Email,
		})

		if strings.Contains(err.Error(), "already exists") {
			h.sendError(w, errors.ErrUserExists, http.StatusConflict)
			return
		}
		h.sendError(w, errors.ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	h.logger.InfoContext(ctx, "account registration successful", map[string]interface{}{
		"email":      req.Email,
		"account_id": newAccount.ID,
	})

	// Return account info (without password) - maintain Auth0 compatibility
	response := map[string]interface{}{
		"account_id":     newAccount.ID,
		"email":          newAccount.Email,
		"name":           newAccount.Name,
		"email_verified": newAccount.Verified,
		"created_at":     newAccount.CreatedAt,
	}

	h.sendJSON(w, response, http.StatusCreated)
}

// GetAccountsHandler handles account listing requests (maintains backward compatibility as GetUsersHandler)
func (h *AuthHandler) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	if r.Method != http.MethodGet {
		h.sendError(w, errors.ErrMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	accounts, err := h.accountUseCase.ListAccounts(ctx, limit, offset)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to list accounts", err, nil)
		h.sendError(w, errors.ErrInternalServerError, http.StatusInternalServerError)
		return
	}

	// Convert to response format (without passwords) - maintain Auth0 compatibility
	response := make([]map[string]interface{}, len(accounts))
	for i, acc := range accounts {
		response[i] = map[string]interface{}{
			"user_id":        acc.ID, // Keep user_id for Auth0 compatibility
			"account_id":     acc.ID, // Also provide account_id
			"email":          acc.Email,
			"name":           acc.Name,
			"email_verified": acc.Verified,
			"created_at":     acc.CreatedAt,
			"updated_at":     acc.UpdatedAt,
		}
	}

	h.sendJSON(w, response, http.StatusOK)
}

// sendJSON sends a JSON response
func (h *AuthHandler) sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode JSON response", err, nil)
	}
}

// sendError sends an error response
func (h *AuthHandler) sendError(w http.ResponseWriter, err *errors.AppError, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":             err.Code,
		"error_description": err.Message,
	}

	if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
		h.logger.Error("failed to encode error response", encodeErr, nil)
	}
}

// sendAuthorizationError sends an OAuth 2.1 authorization error response
func (h *AuthHandler) sendAuthorizationError(w http.ResponseWriter, redirectURI, errorCode, errorDescription, state string) {
	if redirectURI == "" {
		// Can't redirect, send direct error response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response := map[string]interface{}{
			"error":             errorCode,
			"error_description": errorDescription,
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Redirect with error parameters
	redirectURL := redirectURI + "?error=" + errorCode + "&error_description=" + errorDescription
	if state != "" {
		redirectURL += "&state=" + state
	}
	http.Redirect(w, nil, redirectURL, http.StatusFound)
}

// renderLoginForm renders a simple login form for the authorization flow
func (h *AuthHandler) renderLoginForm(w http.ResponseWriter, clientID, redirectURI, state, scope, codeChallenge, codeChallengeMethod string) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>OAuth 2.1 Authorization</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 50px auto; padding: 20px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; }
        input { width: 100%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; }
        button { background: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; }
        .info { background: #f8f9fa; padding: 15px; border-radius: 4px; margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="info">
        <h3>Authorization Request</h3>
        <p><strong>Client ID:</strong> %s</p>
        <p><strong>Scope:</strong> %s</p>
        <p>Please sign in to authorize this application.</p>
    </div>
    
    <form method="POST">
        <div class="form-group">
            <label for="email">Email:</label>
            <input type="email" id="email" name="email" required>
        </div>
        <div class="form-group">
            <label for="password">Password:</label>
            <input type="password" id="password" name="password" required>
        </div>
        <input type="hidden" name="client_id" value="%s">
        <input type="hidden" name="redirect_uri" value="%s">
        <input type="hidden" name="state" value="%s">
        <input type="hidden" name="scope" value="%s">
        <input type="hidden" name="code_challenge" value="%s">
        <input type="hidden" name="code_challenge_method" value="%s">
        <button type="submit">Authorize</button>
    </form>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(html, clientID, scope, clientID, redirectURI, state, scope, codeChallenge, codeChallengeMethod)))
}
