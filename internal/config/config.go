
package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type LogCfg struct {
	Level string `yaml:"level"` // debug|info|warn|error
	JSON  bool   `yaml:"json"`
}

type AdminCfg struct {
	Listen        string `yaml:"listen"`          // e.g. ":8080"
	MetricsListen string `yaml:"metrics_listen"`  // e.g. ":9100"
	AuthToken     string `yaml:"auth_token"`      // Bearer token for admin API
}

type UsersCfg struct {
	Path string `yaml:"path"` // path to user store JSON file
}

type XrayCfg struct {
	Binary     string   `yaml:"binary"`       // path to xray binary
	ConfigPath string   `yaml:"config_path"`  // path to generated xray config
	Listen     string   `yaml:"listen"`       // "0.0.0.0:443"
	Dest       string   `yaml:"dest"`         // REALITY dest e.g. "www.cloudflare.com:443"
	ServerNames []string `yaml:"server_names"` // SNI serverNames used by REALITY
	PrivateKey string   `yaml:"private_key"`  // REALITY X25519 private key (base64 or hex). Prefer using `xray x25519`
	ShortIDs   []string `yaml:"short_ids"`    // shortIds hex strings (8-16 hex chars)
	Egress     string   `yaml:"egress"`       // "tor" | "direct"
	LogLevel   string   `yaml:"loglevel"`     // xray internal loglevel e.g. "warn"
}

type TorCfg struct {
	Enabled          bool   `yaml:"enabled"`
	Socks            string `yaml:"socks"`     // "tor:9050" (docker) or "127.0.0.1:9050"
	BootstrapTimeout string `yaml:"bootstrap_timeout"` // e.g. "30s"
}

type Config struct {
	Log   LogCfg   `yaml:"log"`
	Admin AdminCfg `yaml:"admin"`
	Users UsersCfg `yaml:"users"`
	Xray  XrayCfg  `yaml:"xray"`
	Tor   TorCfg   `yaml:"tor"`
}

func Default() Config {
	return Config{
		Log: LogCfg{Level: "info", JSON: true},
		Admin: AdminCfg{
			Listen: ":8080", MetricsListen: ":9100", AuthToken: "",
		},
		Users: UsersCfg{Path: "/var/lib/vpnd/users.json"},
		Xray: XrayCfg{
			Binary: "/usr/local/bin/xray",
			ConfigPath: "/var/lib/vpnd/xray_config.json",
			Listen: "0.0.0.0:443",
			Dest: "www.cloudflare.com:443",
			ServerNames: []string{"www.cloudflare.com"},
			PrivateKey: "",
			ShortIDs: []string{"aabbccdd"},
			Egress: "tor",
			LogLevel: "warn",
		},
		Tor: TorCfg{Enabled: true, Socks: "tor:9050", BootstrapTimeout: "30s"},
	}
}

func FromFile(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Admin.Listen == "" { return errors.New("admin.listen is required") }
	if c.Admin.MetricsListen == "" { return errors.New("admin.metrics_listen is required") }
	if c.Users.Path == "" { return errors.New("users.path is required") }
	if c.Xray.Binary == "" { return errors.New("xray.binary is required") }
	if c.Xray.ConfigPath == "" { return errors.New("xray.config_path is required") }
	if c.Xray.Listen == "" { return errors.New("xray.listen is required") }
	if len(c.Xray.ServerNames) == 0 { return errors.New("xray.server_names must not be empty") }
	if c.Xray.Dest == "" { return errors.New("xray.dest is required") }
	if c.Xray.Egress != "tor" && c.Xray.Egress != "direct" {
		return fmt.Errorf("xray.egress must be 'tor' or 'direct', got %q", c.Xray.Egress)
	}
	if c.Tor.Enabled && c.Xray.Egress == "tor" && c.Tor.Socks == "" {
		return errors.New("tor.socks is required when tor.enabled and xray.egress==tor")
	}
	if c.Tor.BootstrapTimeout == "" {
		c.Tor.BootstrapTimeout = "30s"
	}
	if _, err := time.ParseDuration(c.Tor.BootstrapTimeout); err != nil {
		return fmt.Errorf("tor.bootstrap_timeout invalid: %w", err)
	}
	return nil
}

func (c *Config) MustValidate() {
	if err := c.Validate(); err != nil {
		panic(err)
	}
}

func (c *Config) Save(path string) error {
	b, err := yaml.Marshal(c)
	if err != nil { return err }
	return os.WriteFile(path, b, 0644)
}
