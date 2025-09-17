package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DBPath     string `yaml:"db_path"`
	APIBaseURL string `yaml:"api_base_url"`
	// TODO: Move configs into `App`
	//App           App            `yaml:"app"`
	Server        Server         `yaml:"server"`
	OIDCProviders []OIDCProvider `yaml:"oidc_providers"`
}

//type App struct {
//	DBPath     string `yaml:"db_path"`
//	APIBaseURL string `yaml:"api_base_url"`
//}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	TLS  TLS    `yaml:"tls"`
}

type TLS struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type OIDCProvider struct {
	// This name can be anything the user wants. It is only used inside this config/application
	Name string `yaml:"name"`

	// example issuer: https://idm.example.com/oauth2/openid/<client_id>
	IssuerURL         string   `yaml:"issuer_url"`
	ClientID          string   `yaml:"client_id"`           // Provided by IdP
	ClientSecret      string   `yaml:"client_secret"`       // Provided by IdP
	RedirectURL       string   `yaml:"redirect_url"`        // TODO: derive this value. example: https://habits.example.com/auth/callback
	Scopes            []string `yaml:"scopes"`              // "openid" and "profile" are requirements; the rest is optional
	LogoutRedirectURL string   `yaml:"logout_redirect_url"` // Optional

	issuerURL         *url.URL `yaml:-`
	redirectURL       *url.URL `yaml:-`
	logoutRedirectURL *url.URL `yaml:-`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	// We have valid YAML at this point. Apply all defaults to missing items
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
		c.Server.Host = "127.0.0.1"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 3000
	}
	if c.APIBaseURL == "" {
		c.APIBaseURL = fmt.Sprintf("http://%s:%d/api", c.Server.Host, c.Server.Port)
	}

	// OIDC defaults
	for i := range c.OIDCProviders {
		provider := &c.OIDCProviders[i]
		if provider.Scopes == nil {
			// You might also consider strongly "email". Not required tho.
			provider.Scopes = []string{"openid", "profile"}
		}
	}
}

func (c *Config) finalize() error {
	var err error

	// ensure parsed objects exit for urls
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
	// TODO: validate URLs here? or in finalize? Or validate them here and convert them in finalize?

	// `seen` is used to detect duplicate providers with the same `name` value
	seen := make(map[string]struct{}, len(c.OIDCProviders))

	// loop over all providers and validate them
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

// Addr returns "host:port" for net/http.
func (s Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Can be used to lookup the named oidc provider struct without having to introspect the list
func (c *Config) GetProviderByName(name string) (*OIDCProvider, bool) {
	for i := range c.OIDCProviders {
		if c.OIDCProviders[i].Name == name {
			return &c.OIDCProviders[i], true
		}
	}
	return nil, false
}
