package config

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/hashicorp/vault-client-go"
	"go.uber.org/zap"
)

const (
	requestTimeout             = 10 * time.Second
	databaseURLKey             = "DATABASE_URL"
	databaseMaxConnsKey        = "DATABASE_MAX_CONNS"
	databaseMinConnsKey        = "DATABASE_MIN_CONNS"
	databaseMaxConnIdleTimeKey = "DATABASE_MAX_CONN_IDLE_TIME"
)

// Config represents the configuration for the application.
type Config struct {
	HTTPServer HTTPServerConfig `envPrefix:"HTTP_"`
	Database   DatabaseConfig   `envPrefix:"DATABASE_"`
	Vault      VaultConfig      `envPrefix:"VAULT_"`
}

// DatabaseConfig represents the configuration for the database.
type DatabaseConfig struct {
	URL               string        `env:"URL" envDefault:"postgres://agent:agent@postgres:5432/agent_corp?sslmode=disable"`
	MaxConns          int32         `env:"MAX_CONNS" envDefault:"10"`
	MinConns          int32         `env:"MIN_CONNS" envDefault:"2"`
	MaxConnIdleTime   time.Duration `env:"MAX_CONN_IDLE_TIME" envDefault:"5m"`
	MaxConnLifetime   time.Duration `env:"MAX_CONN_LIFETIME" envDefault:"30m"`
	HealthCheckPeriod time.Duration `env:"HEALTH_CHECK_PERIOD" envDefault:"1m"`
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

// NewConfig loads configuration from the environment and applies overrides
// from Vault. If Vault is unavailable and not required, it logs a warning
// and continues with environment values.
func NewConfig(ctx context.Context, logger *zap.Logger) (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("parse environment config: %w", err)
	}

	if err := applyVaultOverrides(ctx, &cfg); err != nil {
		if cfg.Vault.Required {
			return nil, err
		}
		logger.Warn("Vault overrides skipped, falling back to environment values", zap.Error(err))
	}

	return &cfg, nil
}

func applyVaultOverrides(ctx context.Context, cfg *Config) error {
	client, err := vault.New(
		vault.WithAddress(cfg.Vault.Addr),
		vault.WithRequestTimeout(requestTimeout),
	)
	if err != nil {
		return fmt.Errorf("create vault client: %w", err)
	}

	if err := client.SetToken(cfg.Vault.Token); err != nil {
		return fmt.Errorf("set vault token: %w", err)
	}

	resp, err := client.Secrets.KvV2Read(
		ctx,
		cfg.Vault.SecretPath,
		vault.WithMountPath(cfg.Vault.MountPath),
	)
	if err != nil {
		return fmt.Errorf("read vault secret %q from mount %q: %w", cfg.Vault.SecretPath, cfg.Vault.MountPath, err)
	}

	if resp == nil {
		return fmt.Errorf("read vault secret %q from mount %q: empty response", cfg.Vault.SecretPath, cfg.Vault.MountPath)
	}

	rawDatabaseURL, ok := resp.Data.Data[databaseURLKey]
	if !ok {
		return fmt.Errorf("vault secret %q does not contain %s", cfg.Vault.SecretPath, databaseURLKey)
	}

	databaseURL, ok := rawDatabaseURL.(string)
	if !ok {
		return fmt.Errorf("vault secret %q contains non-string %s", cfg.Vault.SecretPath, databaseURLKey)
	}

	cfg.Database.URL = databaseURL

	if err := setInt32FromVault(resp.Data.Data, databaseMaxConnsKey, &cfg.Database.MaxConns); err != nil {
		return err
	}

	if err := setInt32FromVault(resp.Data.Data, databaseMinConnsKey, &cfg.Database.MinConns); err != nil {
		return err
	}

	if err := setDurationFromVault(resp.Data.Data, databaseMaxConnIdleTimeKey, &cfg.Database.MaxConnIdleTime); err != nil {
		return err
	}

	return nil
}

func setInt32FromVault(data map[string]any, key string, target *int32) error {
	rawValue, ok := data[key]
	if !ok {
		return nil
	}

	var n int64
	switch v := rawValue.(type) {
	case string:
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return fmt.Errorf("vault secret contains invalid %s: %w", key, err)
		}
		n = parsed
	case json.Number:
		parsed, err := v.Int64()
		if err != nil {
			return fmt.Errorf("vault secret contains invalid %s: %w", key, err)
		}
		n = parsed
	case float64:
		n = int64(v)
	case int:
		n = int64(v)
	case int64:
		n = v
	default:
		return fmt.Errorf("vault secret %s has unsupported type %T", key, rawValue)
	}

	if n > math.MaxInt32 || n < math.MinInt32 {
		return fmt.Errorf("vault secret %s value %d overflows int32", key, n)
	}

	*target = int32(n)
	return nil
}

func setDurationFromVault(data map[string]any, key string, target *time.Duration) error {
	rawValue, ok := data[key]
	if !ok {
		return nil
	}

	s, ok := rawValue.(string)
	if !ok {
		return fmt.Errorf("vault secret %s expected string duration, got %T", key, rawValue)
	}

	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("vault secret contains invalid %s: %w", key, err)
	}

	*target = parsed
	return nil
}
