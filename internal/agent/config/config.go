package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	AgentID     string   `mapstructure:"agent_id"`
	MasterURL   string   `mapstructure:"master_url"`
	ListenAddr  string   `mapstructure:"listen_addr"`
	LogLevel    string   `mapstructure:"log_level"`
	ModulesPath string   `mapstructure:"modules_path"`
}

func Load() (*Config, error) {
	v := viper.New()
	
	v.SetDefault("agent_id", "unknown-agent")
	v.SetDefault("master_url", "localhost:9090")
	v.SetDefault("listen_addr", "0.0.0.0:8080")
	v.SetDefault("log_level", "info")
	v.SetDefault("modules_path", "/var/lib/gimpel/modules")

	v.SetConfigName("gimpel-agent")
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/gimpel/")
	v.AddConfigPath(".")

	v.SetEnvPrefix("GIMPEL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
