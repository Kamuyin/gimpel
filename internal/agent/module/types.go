package module

import (
	"context"
	"net"
	"time"
)

type ModuleState int

const (
	ModuleStateStopped ModuleState = iota
	ModuleStateStarting
	ModuleStateRunning
	ModuleStateFailed
)

type ExecutionMode string

const (
	ExecutionModeUserspace ExecutionMode = "userspace"

	ExecutionModeRoot ExecutionMode = "root"

	ExecutionModeContainerd ExecutionMode = "containerd"

	ExecutionModeSystemd ExecutionMode = "systemd"
)

type ConnectionMode string

const (
	ConnectionModeFDPass ConnectionMode = "fdpass"

	ConnectionModeTCPRelay ConnectionMode = "tcp_relay"

	ConnectionModeProxy ConnectionMode = "proxy"
)

type ModuleCapabilities struct {
	Protocols []string `yaml:"protocols" json:"protocols"`

	Ports []int `yaml:"ports" json:"ports"`

	RequiresRoot bool `yaml:"requires_root" json:"requires_root"`

	SupportsHighInteraction bool `yaml:"supports_high_interaction" json:"supports_high_interaction"`

	CanHandleRawPackets bool `yaml:"can_handle_raw_packets" json:"can_handle_raw_packets"`

	MaxConcurrentConnections int `yaml:"max_concurrent_connections" json:"max_concurrent_connections"`
}

type ResourceLimits struct {
	MaxMemoryMB int64 `yaml:"max_memory_mb" json:"max_memory_mb"`

	MaxCPUPercent int `yaml:"max_cpu_percent" json:"max_cpu_percent"`

	MaxOpenFiles int64 `yaml:"max_open_files" json:"max_open_files"`

	MaxProcesses int64 `yaml:"max_processes" json:"max_processes"`

	NetworkBandwidthKbps int64 `yaml:"network_bandwidth_kbps" json:"network_bandwidth_kbps"`
}

type ModuleSpec struct {
	ID string `yaml:"id" json:"id"`

	Name string `yaml:"name" json:"name"`

	Image string `yaml:"image" json:"image"`

	ExecutionMode ExecutionMode `yaml:"execution_mode" json:"execution_mode"`

	ConnectionMode ConnectionMode `yaml:"connection_mode" json:"connection_mode"`

	SocketPath string `yaml:"socket_path" json:"socket_path"`

	Env map[string]string `yaml:"env" json:"env"`

	WorkingDir string `yaml:"working_dir" json:"working_dir"`

	Capabilities ModuleCapabilities `yaml:"capabilities" json:"capabilities"`

	ResourceLimits ResourceLimits `yaml:"resource_limits" json:"resource_limits"`

	RestartPolicy RestartPolicy `yaml:"restart_policy" json:"restart_policy"`

	HealthCheck HealthCheckConfig `yaml:"health_check" json:"health_check"`
}

type RestartPolicy struct {
	Policy string `yaml:"policy" json:"policy"`

	MaxRestarts int `yaml:"max_restarts" json:"max_restarts"`

	RestartDelay time.Duration `yaml:"restart_delay" json:"restart_delay"`

	BackoffMultiplier float64 `yaml:"backoff_multiplier" json:"backoff_multiplier"`

	MaxBackoffDelay time.Duration `yaml:"max_backoff_delay" json:"max_backoff_delay"`
}

type HealthCheckConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`

	Interval time.Duration `yaml:"interval" json:"interval"`

	Timeout time.Duration `yaml:"timeout" json:"timeout"`

	Retries int `yaml:"retries" json:"retries"`
}

type ConnectionRequest struct {
	ConnectionID string

	ListenerID string

	ModuleID string

	SourceIP string

	SourcePort uint32

	DestIP string

	DestPort uint32

	Protocol string

	Timestamp time.Time

	Conn net.Conn

	FD int

	Metadata map[string]string
}

type ConnectionHandler interface {
	HandleConnection(ctx context.Context, req *ConnectionRequest) error

	SupportsConnectionMode(mode ConnectionMode) bool
}

type ModuleRuntime interface {
	Name() string

	Type() ExecutionMode

	Start(ctx context.Context, spec *ModuleSpec) (*ModuleInstance, error)

	Stop(ctx context.Context, instance *ModuleInstance) error

	Signal(ctx context.Context, instance *ModuleInstance, signal int) error

	IsRunning(ctx context.Context, instance *ModuleInstance) bool

	Logs(ctx context.Context, instance *ModuleInstance, lines int) ([]string, error)
}

type ModuleInstance struct {
	ID string

	Spec *ModuleSpec

	PID int

	ContainerID string

	SocketPath string

	DataPort int

	StartedAt time.Time

	State ModuleState

	RestartCount int

	LastError error

	Conn net.Conn

	StopFunc func()

	Metrics *ModuleMetrics
}

type ModuleMetrics struct {
	ConnectionsTotal int64

	ConnectionsActive int64

	BytesReceived int64

	BytesSent int64

	ErrorsTotal int64

	AvgResponseTimeMs float64

	MemoryUsageBytes int64

	CPUUsagePercent float64

	LastHealthCheck time.Time

	HealthChecksPassed int64

	HealthChecksFailed int64
}

type ConnectionInfo struct {
	ConnectionID string
	SourceIP     string
	SourcePort   uint32
	DestIP       string
	DestPort     uint32
	Protocol     string
	FD           int
}
