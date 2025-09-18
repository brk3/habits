package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AuthEnabled bool   `yaml:"auth_enabled"`
	AuthToken   string `yaml:"auth_token"`
	DBPath      string `yaml:"db_path"`
	APIBaseURL  string `yaml:"api_base_url"`
	Server      struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
		TLS  struct {
			Enabled  bool   `yaml:"enabled"`
			CertFile string `yaml:"cert_file"`
			KeyFile  string `yaml:"key_file"`
		} `yaml:"tls"`
	} `yaml:"server"`
	OIDCProviders []struct {
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
	} `yaml:"oidc_providers"`
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
	if !c.Server.TLS.Enabled {
		c.Server.TLS.Enabled = false
	}
	if !c.AuthEnabled {
		c.AuthEnabled = false
	}
	for i := range c.OIDCProviders {
		provider := &c.OIDCProviders[i]
		if provider.Scopes == nil {
			provider.Scopes = []string{"openid", "profile"}
		}
	}
}

func (c *Config) finalize() error {
	var err error

	for i := range c.OIDCProviders {
		provider := &c.OIDCProviders[i]
		name := provider.Name
		if name == "" {
			return fmt.Errorf("oidc_providers[%d]: name is required", i)
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

	if c.Server.TLS.Enabled {
		if c.Server.TLS.CertFile == "" || c.Server.TLS.KeyFile == "" {
			// TODO: Also check permissions to access these files?
			return errors.New("server.tls enabled but cert_file or key_file missing")
		}
	}
	return nil
}
