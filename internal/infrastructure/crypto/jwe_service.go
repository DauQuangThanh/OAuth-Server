package crypto

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"

	"auth0-server/internal/domain/auth"
)

// JWETokenService implements high-performance JWE token operations
// with connection pooling and concurrent token generation
type JWETokenService struct {
	encryptionKey []byte
	signingKey    []byte
	issuer        string
	audience      []string

	// Performance optimizations
	signerPool    sync.Pool
	encrypterPool sync.Pool
	mutex         sync.RWMutex
}

// NewJWETokenService creates a new JWE token service
func NewJWETokenService(secretKey, issuer string, audience []string) *JWETokenService {
	// Derive encryption and signing keys from the secret
	encKey := make([]byte, 32) // 256-bit key for AES-256
	sigKey := make([]byte, 32) // 256-bit key for HMAC

	// Use the secret to derive keys (in production, use proper key derivation)
	copy(encKey, []byte(secretKey + "_enc")[:32])
	copy(sigKey, []byte(secretKey + "_sig")[:32])

	service := &JWETokenService{
		encryptionKey: encKey,
		signingKey:    sigKey,
		issuer:        issuer,
		audience:      audience,
	}

	// Initialize object pools for better performance
	service.signerPool = sync.Pool{
		New: func() interface{} {
			signer, _ := jose.NewSigner(
				jose.SigningKey{Algorithm: jose.HS256, Key: service.signingKey},
				(&jose.SignerOptions{}).WithType("JWT"),
			)
			return signer
		},
	}

	service.encrypterPool = sync.Pool{
		New: func() interface{} {
			encrypter, _ := jose.NewEncrypter(
				jose.A256GCM,
				jose.Recipient{Algorithm: jose.DIRECT, Key: service.encryptionKey},
				nil,
			)
			return encrypter
		},
	}

	return service
}

// GenerateTokenPair creates access and refresh tokens
func (s *JWETokenService) GenerateTokenPair(ctx context.Context, userID, email, name string) (*auth.TokenPair, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	now := time.Now()

	// Generate access token
	accessClaims := &auth.Claims{
		Subject:   userID,
		Issuer:    s.issuer,
		Audience:  s.audience,
		ExpiresAt: now.Add(24 * time.Hour),
		IssuedAt:  now,
		NotBefore: now,
		Email:     email,
		Name:      name,
		Scope:     "openid profile email",
	}

	accessToken, err := s.createEncryptedToken(accessClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &auth.Claims{
		Subject:   userID,
		Issuer:    s.issuer,
		Audience:  s.audience,
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		IssuedAt:  now,
		NotBefore: now,
		Email:     email,
		Name:      name,
	}

	refreshToken, err := s.createEncryptedToken(refreshClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &auth.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    86400, // 24 hours
		Scope:        "openid profile email",
	}, nil
}

// ValidateToken validates a JWE token and returns claims
func (s *JWETokenService) ValidateToken(ctx context.Context, tokenString string) (*auth.Claims, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Parse the JWE token with expected algorithms
	object, err := jose.ParseEncrypted(tokenString, []jose.KeyAlgorithm{jose.DIRECT}, []jose.ContentEncryption{jose.A256GCM})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE token: %w", err)
	}

	// Decrypt the token
	decrypted, err := object.Decrypt(s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Parse the decrypted JWT
	token, err := jwt.ParseSigned(string(decrypted), []jose.SignatureAlgorithm{jose.HS256})
	if err != nil {
		return nil, fmt.Errorf("failed to parse decrypted JWT: %w", err)
	}

	// Verify and extract claims using a map first
	var rawClaims map[string]interface{}
	err = token.Claims(s.signingKey, &rawClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token signature: %w", err)
	}

	// Convert raw claims to auth.Claims with proper time conversion
	claims := &auth.Claims{}

	if sub, ok := rawClaims["sub"].(string); ok {
		claims.Subject = sub
	}
	if iss, ok := rawClaims["iss"].(string); ok {
		claims.Issuer = iss
	}
	if email, ok := rawClaims["email"].(string); ok {
		claims.Email = email
	}
	if name, ok := rawClaims["name"].(string); ok {
		claims.Name = name
	}
	if scope, ok := rawClaims["scope"].(string); ok {
		claims.Scope = scope
	}

	// Handle audience (can be string or []string)
	if aud, ok := rawClaims["aud"]; ok {
		switch v := aud.(type) {
		case string:
			claims.Audience = []string{v}
		case []interface{}:
			claims.Audience = make([]string, len(v))
			for i, a := range v {
				if s, ok := a.(string); ok {
					claims.Audience[i] = s
				}
			}
		}
	}

	// Convert Unix timestamps to time.Time
	if exp, ok := rawClaims["exp"].(float64); ok {
		claims.ExpiresAt = time.Unix(int64(exp), 0)
	}
	if iat, ok := rawClaims["iat"].(float64); ok {
		claims.IssuedAt = time.Unix(int64(iat), 0)
	}
	if nbf, ok := rawClaims["nbf"].(float64); ok {
		claims.NotBefore = time.Unix(int64(nbf), 0)
	}

	// Validate time-based claims
	if time.Now().After(claims.ExpiresAt) {
		return nil, fmt.Errorf("token has expired")
	}

	if time.Now().Before(claims.NotBefore) {
		return nil, fmt.Errorf("token not yet valid")
	}

	return claims, nil
}

// RefreshToken creates a new token pair from a refresh token
func (s *JWETokenService) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Validate the refresh token
	claims, err := s.ValidateToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new token pair
	return s.GenerateTokenPair(ctx, claims.Subject, claims.Email, claims.Name)
}

// RevokeToken revokes a token (placeholder implementation)
func (s *JWETokenService) RevokeToken(ctx context.Context, token string) error {
	// In a production environment, you would implement token blacklisting
	// For now, this is a placeholder
	return nil
}

// createEncryptedToken creates a JWE token from claims
func (s *JWETokenService) createEncryptedToken(claims *auth.Claims) (string, error) {
	// Get signer from pool
	signer := s.signerPool.Get().(jose.Signer)
	defer s.signerPool.Put(signer)

	// Create custom claims map
	customClaims := map[string]interface{}{
		"email": claims.Email,
		"name":  claims.Name,
		"exp":   claims.ExpiresAt.Unix(),
		"iat":   claims.IssuedAt.Unix(),
		"nbf":   claims.NotBefore.Unix(),
		"sub":   claims.Subject,
		"iss":   claims.Issuer,
		"aud":   claims.Audience,
	}
	if claims.Scope != "" {
		customClaims["scope"] = claims.Scope
	}

	// Serialize claims to JSON
	claimsBytes, err := json.Marshal(customClaims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	// Sign the token
	signedToken, err := signer.Sign(claimsBytes)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	serializedJWT, err := signedToken.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize signed token: %w", err)
	}

	// Get encrypter from pool
	encrypter := s.encrypterPool.Get().(jose.Encrypter)
	defer s.encrypterPool.Put(encrypter)

	// Encrypt the signed JWT
	encrypted, err := encrypter.Encrypt([]byte(serializedJWT))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt token: %w", err)
	}

	encryptedToken, err := encrypted.CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to serialize encrypted token: %w", err)
	}

	return encryptedToken, nil
}
