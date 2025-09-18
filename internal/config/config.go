package config

import (
	"fmt"
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
	} `yaml:"oidc_providers"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	cfg.applyDefaults()

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
