package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v4"

	"github.com/brk3/habits/internal/logger"
)

type OIDCProvider struct {
	Id                string   `yaml:"id"`
	Name              string   `yaml:"name"`
	IssuerURL         string   `yaml:"issuer_url"`
	ClientID          string   `yaml:"client_id"`
	ClientSecret      string   `yaml:"client_secret"`
	RedirectURL       string   `yaml:"redirect_url"`
	Scopes            []string `yaml:"scopes"`
	LogoutRedirectURL string   `yaml:"logout_redirect_url"`

	issuerURL         *url.URL `yaml:"-"`
	redirectURL       *url.URL `yaml:"-"`
	logoutRedirectURL *url.URL `yaml:"-"`
}

type Config struct {
	AuthEnabled bool   `yaml:"auth_enabled"`
	AuthToken   string `yaml:"auth_token"`
	DBPath      string `yaml:"db_path"`
	APIBaseURL  string `yaml:"api_base_url"`
	LogLevel    string `yaml:"log_level"`

	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
		TLS  struct {
			Enabled  bool   `yaml:"enabled"`
			CertFile string `yaml:"cert_file"`
			KeyFile  string `yaml:"key_file"`
		} `yaml:"tls"`
	} `yaml:"server"`

	OIDCProviders []OIDCProvider `yaml:"oidc_providers"`

	Nudge struct {
		NotifyEmail    string `yaml:"notify_email"`
		ResendAPIKey   string `yaml:"resend_api_key"`
		ThresholdHours int    `yaml:"threshold_hours"`
	} `yaml:"nudge"`

	SLogLevel slog.Level `yaml:"-"`
}

func Load() (*Config, error) {
	path := os.Getenv("HABITS_CONFIG")
	if path == "" {
		path = "config.yaml"
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	cfg.applyDefaults()

	if err := cfg.finalize(); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.DBPath == "" {
		c.DBPath = "habits.db"
	}
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 3000
	}
	if c.APIBaseURL == "" {
		c.APIBaseURL = "http://localhost:3000"
	}
	if c.AuthToken == "" {
		c.AuthToken = "XXX"
	}
	if c.LogLevel == "" {
		c.LogLevel = "debug"
	}

	if c.Nudge.ThresholdHours == 0 {
		c.Nudge.ThresholdHours = 3
	}

	for i := range c.OIDCProviders {
		provider := &c.OIDCProviders[i]
		if provider.Scopes == nil {
			provider.Scopes = []string{"openid", "profile", "offline_access"}
		}
	}
}

func (c *Config) finalize() error {
	var err error

	switch strings.ToLower(c.LogLevel) {
	case "debug":
		c.SLogLevel = slog.LevelDebug
	case "info":
		c.SLogLevel = slog.LevelInfo
	case "warn":
		c.SLogLevel = slog.LevelWarn
	case "error":
		c.SLogLevel = slog.LevelError
	default:
		return fmt.Errorf("invalid log_level: %s", c.LogLevel)
	}

	if c.DBPath != "" {
		if c.DBPath, err = resolvePath(c.DBPath); err != nil {
			return fmt.Errorf("file does not exist > db_path: %w", err)
		}
	}

	if c.Server.TLS.CertFile != "" {
		if c.Server.TLS.CertFile, err = resolvePath(c.Server.TLS.CertFile); err != nil {
			return fmt.Errorf("file does not exist > server.tls.cert_file: %w", err)
		}
	}
	if c.Server.TLS.KeyFile != "" {
		if c.Server.TLS.KeyFile, err = resolvePath(c.Server.TLS.KeyFile); err != nil {
			return fmt.Errorf("file does not exist > server.tls.key_file: %w", err)
		}
	}

	for i := range c.OIDCProviders {
		provider := &c.OIDCProviders[i]
		name := provider.Name
		if name == "" {
			return fmt.Errorf("oidc_providers[%d]: name is required", i)
		}

		if provider.Id == "" {
			return fmt.Errorf("oidc_providers[%d] (%s): id is required", i, name)
		}

		provider.issuerURL, err = url.Parse(provider.IssuerURL)
		if err != nil {
			return fmt.Errorf("oidc_providers[%d] (%s): issuer_url is not a valid URL", i, name)
		}

		provider.redirectURL, err = url.Parse(provider.RedirectURL)
		if err != nil {
			return fmt.Errorf("oidc_providers[%d] (%s): redirect_url is not a valid URL", i, name)
		}

		if provider.LogoutRedirectURL != "" {
			provider.logoutRedirectURL, err = url.Parse(provider.LogoutRedirectURL)
			if err != nil {
				return fmt.Errorf("oidc_providers[%d] (%s): logout_redirect_url is not a valid URL", i, name)
			}
		}
	}
	return nil
}

func (c *Config) validate() error {
	tlscfg := c.Server.TLS
	if tlscfg.Enabled {
		if tlscfg.CertFile == "" || tlscfg.KeyFile == "" {
			return errors.New("server.tls enabled but cert_file or key_file missing")
		}

		// Check public certificate file:
		// 1. stat: does it exist?
		// 2. open: can we open the file?
		cert := tlscfg.CertFile
		_, err := fileStat(cert)
		if err != nil {
			return fmt.Errorf("server.tls.cert_file: %w", err)
		}
		if err := canOpenFile(cert); err != nil {
			return fmt.Errorf("server.tls.cert_file not readable: %w", err)
		}

		// Check private key file:
		// 1. stat: does it exist?
		// 2. mode: is the private file to permissive?
		// 3. open: can we open the file?
		key := tlscfg.KeyFile
		fiKey, err := fileStat(key)
		if err != nil {
			return fmt.Errorf("server.tls.key_file: %w", err)
		}

		mode := fiKey.Mode().Perm()
		if mode&0o077 != 0 {
			return fmt.Errorf("server.tls.key_file: %s permissions too permissive (%#o); expected 0600", key, mode)
		}

		if err := canOpenFile(key); err != nil {
			return fmt.Errorf("server.tls.key_file not readable: %w", err)
		}

	}

	if len(c.OIDCProviders) == 0 && c.AuthEnabled {
		return errors.New("authentication was enabled, but no OIDC Providers were configured")
	}

	if len(c.OIDCProviders) > 0 && !c.AuthEnabled {
		logger.Warn("OIDC Providers have been configured, but auth is disabled")
	}

	seen := make(map[string]struct{}, len(c.OIDCProviders))
	for i := range c.OIDCProviders {
		provider := &c.OIDCProviders[i]
		name := provider.Name

		if name == "" {
			return fmt.Errorf("oidc_providers[%d]: name is required", i)
		}

		if _, dup := seen[name]; dup {
			return fmt.Errorf("duplicate provider name %q in oidc_providers", name)
		}
		seen[name] = struct{}{}

		if provider.Id == "" {
			return fmt.Errorf("oidc_providers[%d] (%s): id is required", i, name)
		}

		if provider.IssuerURL == "" {
			return fmt.Errorf("oidc_providers[%d] (%q): issuer_url is required", i, name)
		}

		if provider.RedirectURL == "" {
			return fmt.Errorf("oidc_providers[%d] (%q): redirect_url is required", i, name)
		}

		if provider.ClientID == "" {
			return fmt.Errorf("oidc_providers[%d] (%q): client_id is required", i, name)
		}

		if provider.ClientSecret == "" {
			return fmt.Errorf("oidc_providers[%d] (%q): client_secret is required", i, name)
		}
	}
	return nil
}

// you give path, i give _real_ path
func resolvePath(p string) (string, error) {
	if p == "" {
		return "", errors.New("empty path")
	}

	abs, err := filepath.Abs(filepath.Clean(p))
	if err != nil {
		return "", fmt.Errorf("abs path: %w", err)
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		// If the file doesn't exist yet, EvalSymlinks will fail; return absolute path
		return abs, nil
	}
	return resolved, nil
}

// calls OS stat() on the file and returns info
func fileStat(p string) (os.FileInfo, error) {
	fi, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	if !fi.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", p)
	}
	return fi, nil
}

// opens file and closes it; does not read from file
func canOpenFile(p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	if closeErr := f.Close(); closeErr != nil {
		logger.Warn("Failed to close file after checking accessibility", "file", p, "error", closeErr)
	}
	return nil
}
