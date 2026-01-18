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

type ResourceLimitsConfig struct {
	MaxMemoryMB      int64 `mapstructure:"max_memory_mb"`
	MaxCPUPercent    int   `mapstructure:"max_cpu_percent"`
	MaxOpenFiles     int64 `mapstructure:"max_open_files"`
	MaxProcesses     int64 `mapstructure:"max_processes"`
	NetworkBandwidth int64 `mapstructure:"network_bandwidth_kbps"`
}

type HealthCheckConfig struct {
	Enabled  bool          `mapstructure:"enabled"`
	Interval time.Duration `mapstructure:"interval"`
	Timeout  time.Duration `mapstructure:"timeout"`
	Retries  int           `mapstructure:"retries"`
}

type RestartPolicyConfig struct {
	Policy            string        `mapstructure:"policy"`
	MaxRestarts       int           `mapstructure:"max_restarts"`
	RestartDelay      time.Duration `mapstructure:"restart_delay"`
	BackoffMultiplier float64       `mapstructure:"backoff_multiplier"`
	MaxBackoffDelay   time.Duration `mapstructure:"max_backoff_delay"`
}

type ModuleConfig struct {
	ID         string            `mapstructure:"id"`
	Name       string            `mapstructure:"name"`
	Image      string            `mapstructure:"image"`
	SocketPath string            `mapstructure:"socket_path"`
	Env        map[string]string `mapstructure:"env"`
	Listeners  []ListenerConfig  `mapstructure:"listeners"`

	ExecutionMode  string `mapstructure:"execution_mode"`
	ConnectionMode string `mapstructure:"connection_mode"`
	WorkingDir     string `mapstructure:"working_dir"`

	RequiresRoot        bool `mapstructure:"requires_root"`
	CanHandleRawPackets bool `mapstructure:"can_handle_raw_packets"`

	ResourceLimits ResourceLimitsConfig `mapstructure:"resource_limits"`

	HealthCheck HealthCheckConfig `mapstructure:"health_check"`

	RestartPolicy RestartPolicyConfig `mapstructure:"restart_policy"`
}

type RuntimeConfig struct {
	DefaultExecutionMode  string `mapstructure:"default_execution_mode"`
	DefaultConnectionMode string `mapstructure:"default_connection_mode"`
	EnablePrivileged      bool   `mapstructure:"enable_privileged"`
	EnableContainerd      bool   `mapstructure:"enable_containerd"`
	ContainerdAddress     string `mapstructure:"containerd_address"`
	ContainerdNamespace   string `mapstructure:"containerd_namespace"`
	TrustedKeyFile        string `mapstructure:"trusted_key_file"`
	ModuleCacheDir        string `mapstructure:"module_cache_dir"`
}

type AgentConfig struct {
	AgentID           string        `mapstructure:"agent_id"`
	DataDir           string        `mapstructure:"data_dir"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	RegistrationToken string        `mapstructure:"registration_token"`

	ControlPlane ControlPlaneConfig `mapstructure:"control_plane"`
	Gateway      GatewayConfig      `mapstructure:"gateway"`
	Modules      []ModuleConfig     `mapstructure:"modules"`
	Runtime      RuntimeConfig      `mapstructure:"runtime"`
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
