package auth

import (
	"context"
	"time"
)

// TokenType represents different types of tokens
type TokenType string

const (
	AccessToken  TokenType = "access_token"
	RefreshToken TokenType = "refresh_token"
	IDToken      TokenType = "id_token"
)

// Token represents an authentication token
type Token struct {
	Value     string    `json:"token"`
	Type      TokenType `json:"type"`
	ExpiresAt time.Time `json:"expires_at"`
	Subject   string    `json:"subject"`
	Audience  []string  `json:"audience,omitempty"`
	Scopes    []string  `json:"scopes,omitempty"`
}

// TokenPair represents a complete token response
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
}

// Claims represents token claims
type Claims struct {
	Subject   string    `json:"sub"`
	Issuer    string    `json:"iss"`
	Audience  []string  `json:"aud"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
	NotBefore time.Time `json:"nbf"`
	// Custom claims
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
	Scope string `json:"scope,omitempty"`
}

// TokenService defines the interface for token operations
type TokenService interface {
	GenerateTokenPair(ctx context.Context, userID, email, name string) (*TokenPair, error)
	ValidateToken(ctx context.Context, token string) (*Claims, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
	RevokeToken(ctx context.Context, token string) error
}

// Authenticator defines the interface for authentication operations
type Authenticator interface {
	Authenticate(ctx context.Context, email, password string) (*TokenPair, error)
	ValidateToken(ctx context.Context, token string) (*Claims, error)
	RefreshAuthentication(ctx context.Context, refreshToken string) (*TokenPair, error)
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// TokenRequest represents an OAuth2 token request
type TokenRequest struct {
	GrantType    string `json:"grant_type" validate:"required"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// AuthorizationCode represents an OAuth 2.1 authorization code with PKCE
type AuthorizationCode struct {
	Code                string    `json:"code"`
	ClientID            string    `json:"client_id"`
	RedirectURI         string    `json:"redirect_uri"`
	Scope               string    `json:"scope"`
	AccountID           string    `json:"account_id"`
	CodeChallenge       string    `json:"code_challenge"`
	CodeChallengeMethod string    `json:"code_challenge_method"`
	ExpiresAt           time.Time `json:"expires_at"`
	Used                bool      `json:"used"`
}

// PKCEChallenge represents PKCE challenge data
type PKCEChallenge struct {
	Challenge       string `json:"challenge"`
	ChallengeMethod string `json:"challenge_method"`
	Verifier        string `json:"verifier,omitempty"`
}

// AuthorizationCodeService defines operations for authorization codes
type AuthorizationCodeService interface {
	CreateAuthorizationCode(ctx context.Context, accountID, clientID, redirectURI, scope, codeChallenge, codeChallengeMethod string) (string, error)
	ExchangeCodeForTokens(ctx context.Context, code, clientID, codeVerifier, redirectURI string) (*TokenPair, error)
	ValidatePKCE(codeChallenge, codeVerifier, method string) bool
}
