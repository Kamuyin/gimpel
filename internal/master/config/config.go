package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type TLSConfig struct {
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
	CAFile   string `mapstructure:"ca_file"`
}

type CAConfig struct {
	CertFile     string        `mapstructure:"cert_file"`
	KeyFile      string        `mapstructure:"key_file"`
	Organization string        `mapstructure:"organization"`
	ValidityDays int           `mapstructure:"validity_days"`
	KeySize      int           `mapstructure:"key_size"`
	AutoGenerate bool          `mapstructure:"auto_generate"`
	CRLPath      string        `mapstructure:"crl_path"`
	TTL          time.Duration `mapstructure:"ttl"`
}

type RegistryConfig struct {
	StaleTimeout    time.Duration `mapstructure:"stale_timeout"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

type SandboxConfig struct {
	Nodes []string `mapstructure:"nodes"`
}

type ModuleStoreConfig struct {
	DataDir        string `mapstructure:"data_dir"`
}

type MasterConfig struct {
	ListenAddress      string   `mapstructure:"listen_address"`
	RESTAddress        string   `mapstructure:"rest_address"`
	DataDir            string   `mapstructure:"data_dir"`

	TLS         TLSConfig         `mapstructure:"tls"`
	CA          CAConfig          `mapstructure:"ca"`
	Registry    RegistryConfig    `mapstructure:"registry"`
	Sandbox     SandboxConfig     `mapstructure:"sandbox"`
	ModuleStore ModuleStoreConfig `mapstructure:"module_store"`
}

func (c *MasterConfig) Validate() error {
	if c.ListenAddress == "" {
		c.ListenAddress = ":9090"
	}
	if c.DataDir == "" {
		c.DataDir = "/var/lib/gimpel-master"
	}
	if c.CA.Organization == "" {
		c.CA.Organization = "Gimpel"
	}
	if c.CA.ValidityDays == 0 {
		c.CA.ValidityDays = 365
	}
	if c.CA.KeySize == 0 {
		c.CA.KeySize = 2048
	}
	if c.CA.CertFile == "" {
		c.CA.CertFile = c.DataDir + "/ca.crt"
	}
	if c.CA.KeyFile == "" {
		c.CA.KeyFile = c.DataDir + "/ca.key"
	}
	if c.Registry.StaleTimeout == 0 {
		c.Registry.StaleTimeout = 5 * time.Minute
	}
	if c.Registry.CleanupInterval == 0 {
		c.Registry.CleanupInterval = 1 * time.Minute
	}

	if c.ModuleStore.DataDir == "" {
		c.ModuleStore.DataDir = c.DataDir + "/modules"
	}

	return nil
}

func Load(path string) (*MasterConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	v.SetEnvPrefix("GIMPEL_MASTER")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg MasterConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}
