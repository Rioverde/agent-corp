package configs

import (
	"context"
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/hashicorp/vault-client-go"
)

const (
	requestTimeout = 10 * time.Second
	databaseURLKey = "DATABASE_URL"
)

// Config represents the configuration for the application.
type Config struct {
	HTTPServer HTTPServerConfig `envPrefix:"HTTP_"`
	Database   DatabaseConfig   `envPrefix:"DATABASE_"`
	Vault      VaultConfig      `envPrefix:"VAULT_"`
}

// DatabaseConfig represents the configuration for the database.
type DatabaseConfig struct {
	URL string `env:"URL" envDefault:"postgres://agent:agent@postgres:5432/agent_corp?sslmode=disable"`
}

// HTTPServerConfig represents the configuration for the HTTP server.
type HTTPServerConfig struct {
	Address     string        `env:"ADDR" envDefault:":8080"`
	Timeout     time.Duration `env:"TIMEOUT" envDefault:"20s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" envDefault:"20s"`
}

// VaultConfig represents the configuration for the Vault.
type VaultConfig struct {
	Addr       string `env:"ADDR" envDefault:"http://vault:8200"`
	Token      string `env:"TOKEN" envDefault:"dev-root-token"`
	Required   bool   `env:"REQUIRED" envDefault:"false"`
	MountPath  string `env:"MOUNT_PATH" envDefault:"secret"`
	SecretPath string `env:"SECRET_PATH" envDefault:"auth"`
}

// NewConfig loads configuration from the environment and optionally overrides it
// with values from Vault.
func NewConfig(ctx context.Context) (*Config, error) {
	// parse environment variables
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("parse environment config: %w", err)
	}

	// apply overrides from Vault
	if err := applyVaultOverrides(ctx, &cfg); err != nil {
		if cfg.Vault.Required {
			return nil, err
		}
	}

	return &cfg, nil
}

// applyVaultOverrides reads the Vault secret at the given path
func applyVaultOverrides(ctx context.Context, cfg *Config) error {
	// create vault client
	client, err := vault.New(
		vault.WithAddress(cfg.Vault.Addr),
		vault.WithRequestTimeout(requestTimeout),
	)
	// if the Vault is not required, we can ignore the error
	if err != nil {
		return fmt.Errorf("create vault client: %w", err)
	}

	// set the Vault token
	if err := client.SetToken(cfg.Vault.Token); err != nil {
		return fmt.Errorf("set vault token: %w", err)
	}

	// read the Vault secret
	resp, err := client.Secrets.KvV2Read(
		ctx,
		cfg.Vault.SecretPath,
		vault.WithMountPath(cfg.Vault.MountPath),
	)

	// if the Vault is not required, we can ignore the error
	if err != nil {
		return fmt.Errorf("read vault secret %q from mount %q: %w", cfg.Vault.SecretPath, cfg.Vault.MountPath, err)
	}

	if resp == nil {
		return fmt.Errorf("read vault secret %q from mount %q: empty response", cfg.Vault.SecretPath, cfg.Vault.MountPath)
	}

	// check if the secret contains the database URL
	rawDatabaseURL, ok := resp.Data.Data[databaseURLKey]
	if !ok {
		return fmt.Errorf("vault secret %q does not contain %s", cfg.Vault.SecretPath, databaseURLKey)
	}

	// check if the secret contains the database URL
	databaseURL, ok := rawDatabaseURL.(string)
	if !ok {
		return fmt.Errorf("vault secret %q contains non-string %s", cfg.Vault.SecretPath, databaseURLKey)
	}

	// set the database URL
	cfg.Database.URL = databaseURL
	return nil
}
