package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type TLSConfig struct {
	CertFile   string `mapstructure:"cert_file"`
	KeyFile    string `mapstructure:"key_file"`
	CAFile     string `mapstructure:"ca_file"`
	SkipVerify bool   `mapstructure:"skip_verify"`
}

type SandboxConfig struct {
	ListenAddress string    `mapstructure:"listen_address"`
	PublicIP      string    `mapstructure:"public_ip"`
	TLS           TLSConfig `mapstructure:"tls"`
	LogLevel      string    `mapstructure:"log_level"`
}

func (c *SandboxConfig) Validate() error {
	if c.ListenAddress == "" {
		return fmt.Errorf("listen_address is required")
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	if c.PublicIP == "" {
		c.PublicIP = "127.0.0.1"
	}
	return nil
}

func Load(path string) (*SandboxConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	v.SetEnvPrefix("GIMPEL_SANDBOX")
	v.AutomaticEnv()

	v.SetDefault("listen_address", ":5000")
	v.SetDefault("log_level", "info")

	if err := v.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg SandboxConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}
