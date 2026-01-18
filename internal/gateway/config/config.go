package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type TLSConfig struct {
	CertFile   string `mapstructure:"cert_file"`
	KeyFile    string `mapstructure:"key_file"`
	CAFile     string `mapstructure:"ca_file"`
	SkipVerify bool   `mapstructure:"skip_verify"`
}

type GatewayConfig struct {
	ListenAddress string        `mapstructure:"listen_address"`
	TLS           TLSConfig     `mapstructure:"tls"`
	LogLevel      string        `mapstructure:"log_level"`
	FlushInterval time.Duration `mapstructure:"flush_interval"`
}

func (c *GatewayConfig) Validate() error {
	if c.ListenAddress == "" {
		return fmt.Errorf("listen_address is required")
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	return nil
}

func Load(path string) (*GatewayConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	v.SetEnvPrefix("GIMPEL_GATEWAY")
	v.AutomaticEnv()

	v.SetDefault("listen_address", ":8081")
	v.SetDefault("log_level", "info")

	if err := v.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg GatewayConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}
