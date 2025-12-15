package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerAddress string
	JWESecret     string
	Issuer        string
	Domain        string
}

func LoadConfig() (*Config, error) {
	addr := os.Getenv("SERVER_ADDRESS")
	if addr == "" {
		addr = ":8080"
	}

	jweSecret := os.Getenv("JWE_SECRET")
	if jweSecret == "" {
		return nil, fmt.Errorf("JWE_SECRET is required")
	}

	issuer := os.Getenv("ISSUER")
	if issuer == "" {
		issuer = "auth0-server"
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "localhost:8080"
	}

	return &Config{
		ServerAddress: addr,
		JWESecret:     jweSecret,
		Issuer:        issuer,
		Domain:        domain,
	}, nil
}
