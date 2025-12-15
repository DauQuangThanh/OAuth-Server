package handlers

import (
	"encoding/json"
	"net/http"

	"auth0-server/internal/config"
	"auth0-server/pkg/logger"
)

// ConfigHandler handles configuration-related endpoints
type ConfigHandler struct {
	config *config.Config
	logger logger.Logger
}

// NewConfigHandler creates a new configuration handler
func NewConfigHandler(cfg *config.Config, logger logger.Logger) *ConfigHandler {
	return &ConfigHandler{
		config: cfg,
		logger: logger,
	}
}

// HealthHandler handles health check requests
func (h *ConfigHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// OpenIDConfigurationHandler handles OpenID Connect discovery
func (h *ConfigHandler) OpenIDConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	baseURL := "http://" + h.config.Domain
	if r.TLS != nil {
		baseURL = "https://" + h.config.Domain
	}

	// OAuth 2.1 (draft-ietf-oauth-v2-1-14) compliant configuration
	// References: https://datatracker.ietf.org/doc/draft-ietf-oauth-v2-1/
	// Also implements RFC 9700 (OAuth 2.0 Security Best Practices)
	config := map[string]interface{}{
		"issuer":                 h.config.Issuer,
		"authorization_endpoint": baseURL + "/authorize",
		"token_endpoint":         baseURL + "/oauth/token",
		"userinfo_endpoint":      baseURL + "/userinfo",
		"jwks_uri":               baseURL + "/.well-known/jwks.json",
		"scopes_supported": []string{
			"openid", "profile", "email",
		},
		"response_types_supported": []string{
			"code", // Only authorization code flow per OAuth 2.1 (implicit grant removed)
		},
		"response_modes_supported": []string{
			"query", // OAuth 2.1 default for authorization code flow
		},
		"grant_types_supported": []string{
			"authorization_code", "refresh_token", // OAuth 2.1 compliant grants only (password/implicit removed)
		},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256", "HS256"}, // RS256 REQUIRED per OIDC spec
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic", "none"},
		"claims_supported": []string{
			"sub", "iss", "aud", "exp", "iat", "nbf", "email", "email_verified", "name", "nickname", "picture",
		},
		"code_challenge_methods_supported": []string{
			"S256", // REQUIRED: Only S256 per OAuth 2.1 (plain method removed for security)
		},
		// OAuth 2.1 specific metadata
		"authorization_response_iss_parameter_supported": true,  // RFC 9207 - Authorization Response Issuer Identifier
		"require_pushed_authorization_requests":          false, // PAR not required but supported in future
		"dpop_signing_alg_values_supported":              []string{}, // DPoP support placeholder for future
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		h.logger.Error("Failed to encode OpenID configuration", err, nil)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
