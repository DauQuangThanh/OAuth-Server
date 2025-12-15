package usecases

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"auth0-server/internal/domain/account"
	"auth0-server/internal/domain/auth"
)

// AuthUseCase handles authentication business logic
type AuthUseCase struct {
	accountUseCase     *AccountUseCase
	tokenService       auth.TokenService
	authorizationCodes map[string]*auth.AuthorizationCode // In-memory store for demo
}

// NewAuthUseCase creates a new authentication use case
func NewAuthUseCase(accountUseCase *AccountUseCase, tokenService auth.TokenService) *AuthUseCase {
	return &AuthUseCase{
		accountUseCase:     accountUseCase,
		tokenService:       tokenService,
		authorizationCodes: make(map[string]*auth.AuthorizationCode),
	}
}

// Authenticate authenticates an account and returns tokens
func (uc *AuthUseCase) Authenticate(ctx context.Context, email, password string) (*auth.TokenPair, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Validate credentials
	acc, err := uc.accountUseCase.ValidateCredentials(ctx, email, password)
	if err != nil {
		return nil, err
	}

	// Generate token pair
	return uc.tokenService.GenerateTokenPair(ctx, acc.ID, acc.Email, acc.Name)
}

// ValidateToken validates a token and returns claims
func (uc *AuthUseCase) ValidateToken(ctx context.Context, token string) (*auth.Claims, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	return uc.tokenService.ValidateToken(ctx, token)
}

// RefreshAuthentication refreshes an authentication session
func (uc *AuthUseCase) RefreshAuthentication(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	return uc.tokenService.RefreshToken(ctx, refreshToken)
}

// GetAccountProfile gets account profile information from a token (maintains Auth0 compatibility as "user" profile)
func (uc *AuthUseCase) GetAccountProfile(ctx context.Context, token string) (*account.AccountProfile, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Validate token
	claims, err := uc.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Get account details
	acc, err := uc.accountUseCase.GetAccount(ctx, claims.Subject)
	if err != nil {
		return nil, err
	}

	return &account.AccountProfile{
		ID:            acc.ID,
		Email:         acc.Email,
		EmailVerified: acc.Verified,
		Name:          acc.Name,
		Nickname:      acc.Nickname,
		Picture:       acc.Picture,
	}, nil
}

// CreateAuthorizationCode creates an authorization code for OAuth 2.1 flow
func (uc *AuthUseCase) CreateAuthorizationCode(ctx context.Context, email, password, clientID, redirectURI, scope, codeChallenge, codeChallengeMethod string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	// Authenticate the user (this is internal to the authorization server, not a password grant)
	acc, err := uc.accountUseCase.GetAccountByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// Verify password
	if !uc.accountUseCase.VerifyPassword(acc.Password, password) {
		return "", fmt.Errorf("authentication failed: invalid credentials")
	}

	// Generate authorization code
	codeBytes := make([]byte, 32)
	if _, err := rand.Read(codeBytes); err != nil {
		return "", fmt.Errorf("failed to generate authorization code: %w", err)
	}

	code := base64.URLEncoding.EncodeToString(codeBytes)

	// Store authorization code
	authCode := &auth.AuthorizationCode{
		Code:                code,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		Scope:               scope,
		AccountID:           acc.ID,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		ExpiresAt:           time.Now().Add(10 * time.Minute), // 10 minute expiry
		Used:                false,
	}

	uc.authorizationCodes[code] = authCode

	return code, nil
}

// ExchangeCodeForTokens exchanges an authorization code for tokens (OAuth 2.1 with PKCE)
func (uc *AuthUseCase) ExchangeCodeForTokens(ctx context.Context, code, clientID, codeVerifier, redirectURI string) (*auth.TokenPair, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Retrieve authorization code
	authCode, exists := uc.authorizationCodes[code]
	if !exists {
		return nil, fmt.Errorf("invalid authorization code")
	}

	// Check if code is expired
	if time.Now().After(authCode.ExpiresAt) {
		delete(uc.authorizationCodes, code)
		return nil, fmt.Errorf("authorization code expired")
	}

	// Check if code has been used (one-time use)
	if authCode.Used {
		delete(uc.authorizationCodes, code)
		return nil, fmt.Errorf("authorization code already used")
	}

	// Validate client ID
	if authCode.ClientID != clientID {
		return nil, fmt.Errorf("invalid client ID")
	}

	// Validate redirect URI
	if authCode.RedirectURI != redirectURI {
		return nil, fmt.Errorf("invalid redirect URI")
	}

	// Validate PKCE
	if !uc.validatePKCE(authCode.CodeChallenge, codeVerifier, authCode.CodeChallengeMethod) {
		return nil, fmt.Errorf("PKCE validation failed")
	}

	// Mark code as used
	authCode.Used = true

	// Get account details
	acc, err := uc.accountUseCase.GetAccount(ctx, authCode.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Generate tokens
	tokenPair, err := uc.tokenService.GenerateTokenPair(ctx, acc.ID, acc.Email, acc.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Clean up authorization code
	delete(uc.authorizationCodes, code)

	return tokenPair, nil
}

// validatePKCE validates PKCE challenge and verifier
func (uc *AuthUseCase) validatePKCE(codeChallenge, codeVerifier, method string) bool {
	if method != "S256" {
		return false // OAuth 2.1 requires S256
	}

	// Calculate SHA256 of code verifier
	hash := sha256.Sum256([]byte(codeVerifier))
	expectedChallenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])

	return expectedChallenge == codeChallenge
}
