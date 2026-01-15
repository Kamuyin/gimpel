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

type ControlPlaneConfig struct {
	Address string    `mapstructure:"address"`
	TLS     TLSConfig `mapstructure:"tls"`
}

type GatewayConfig struct {
	Address        string        `mapstructure:"address"`
	TLS            TLSConfig     `mapstructure:"tls"`
	FlushInterval  time.Duration `mapstructure:"flush_interval"`
	BatchSize      int           `mapstructure:"batch_size"`
	BufferPath     string        `mapstructure:"buffer_path"`
	MaxBufferBytes int64         `mapstructure:"max_buffer_bytes"`
}

type ListenerConfig struct {
	ID              string `mapstructure:"id"`
	Protocol        string `mapstructure:"protocol"`
	Port            int    `mapstructure:"port"`
	ModuleID        string `mapstructure:"module_id"`
	HighInteraction bool   `mapstructure:"high_interaction"`
}

type ModuleConfig struct {
	ID         string            `mapstructure:"id"`
	Name       string            `mapstructure:"name"`
	Image      string            `mapstructure:"image"`
	SocketPath string            `mapstructure:"socket_path"`
	Env        map[string]string `mapstructure:"env"`
	Listeners  []ListenerConfig  `mapstructure:"listeners"`
}

type AgentConfig struct {
	AgentID           string        `mapstructure:"agent_id"`
	DataDir           string        `mapstructure:"data_dir"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	RegistrationToken string        `mapstructure:"registration_token"`

	ControlPlane ControlPlaneConfig `mapstructure:"control_plane"`
	Gateway      GatewayConfig      `mapstructure:"gateway"`
	Modules      []ModuleConfig     `mapstructure:"modules"`
}

func (c *AgentConfig) Validate() error {
	if c.ControlPlane.Address == "" {
		return fmt.Errorf("control_plane.address is required")
	}
	if c.Gateway.Address == "" {
		return fmt.Errorf("gateway.address is required")
	}
	if c.DataDir == "" {
		c.DataDir = "/var/lib/gimpel"
	}
	if c.HeartbeatInterval == 0 {
		c.HeartbeatInterval = 30 * time.Second
	}
	if c.Gateway.FlushInterval == 0 {
		c.Gateway.FlushInterval = 5 * time.Second
	}
	if c.Gateway.BatchSize == 0 {
		c.Gateway.BatchSize = 100
	}
	if c.Gateway.BufferPath == "" {
		c.Gateway.BufferPath = c.DataDir + "/events"
	}
	if c.Gateway.MaxBufferBytes == 0 {
		c.Gateway.MaxBufferBytes = 100 * 1024 * 1024
	}

	for i := range c.Modules {
		mod := &c.Modules[i]
		for j := range mod.Listeners {
			if mod.Listeners[j].ModuleID == "" {
				mod.Listeners[j].ModuleID = mod.ID
			}
		}
	}

	return nil
}

func Load(path string) (*AgentConfig, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	v.SetEnvPrefix("GIMPEL")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg AgentConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}
